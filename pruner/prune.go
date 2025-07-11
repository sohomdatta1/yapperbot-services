package main

import (
	"database/sql"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	// needs to be blank-imported to make the driver work
	_ "github.com/go-sql-driver/mysql"
	"github.com/sohomdatta1/yapperbot-services/ybtools"
)

//
// Yapperbot-Pruner, the user pruning bot for Wikipedia
// Copyright (C) 2020 Naypta

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

type preppedStatementsCallback func()

const indeffedUsers, inactiveUsers int8 = 0, 1

const mediaWikiTimestampFormat string = "20060102150405"

const lastEditQueryTemplate string = `SELECT actor_user.actor_name FROM revision_userindex
INNER JOIN actor_user ON actor_user.actor_name = ? AND actor_id = rev_actor
WHERE rev_timestamp > ? LIMIT 1;`

const blockQueryTemplate string = `SELECT ipb_id FROM block_target
INNER JOIN user ON user_name = ? AND user_id = bt_user
INNER JOIN block ON bl_target = bt_id
WHERE bl_expiry = "infinity" AND bl_timestamp < ? LIMIT 1;`

const userRedirectQueryTemplate string = `SELECT rd_title FROM redirect
INNER JOIN page ON page_namespace = 3 AND page_title = ? AND rd_from = page_id
WHERE page_is_redirect = 1 AND rd_namespace = 3 LIMIT 1;`

// regexReplaceCaptureGroupExpression is either insanity or a stroke of genius; one or the other.
// It is a regex that matches on the provided format regex's capture group, as there is only one,
// for the usernames, and will later be used to replace that capture group's contents with the usernames
// we have discovered. Note that it will also match one character before the capture group, which
// should be preserved; this is only because RE2 doesn't support lookbehind... *sigh*
// That character is in capture group 1, so it can be easily restored using Replace.
const regexReplaceCaptureGroupExpression string = `([^\\])\([^?].*?[^\\]\)`

var regexReplaceCaptureGroup *regexp.Regexp

// both Stmt and DB are safe for use by multiple goroutines, so this should be okay
var lastEditQuery, blockQuery, userRedirectQuery *sql.Stmt
var conn *sql.DB

func init() {
	regexReplaceCaptureGroup = regexp.MustCompile(regexReplaceCaptureGroupExpression)
}

func withDatabaseConnection(cb preppedStatementsCallback) {
	var err error

	conn, err = sql.Open("mysql", config.DSN)
	if err != nil {
		ybtools.PanicErr("DSN invalid with error ", err)
	}
	if err := conn.Ping(); err != nil {
		ybtools.PanicErr(err)
	}

	lastEditQuery, err = conn.Prepare(lastEditQueryTemplate)
	if err != nil {
		ybtools.PanicErr("lastEditQuery preparation failed with error ", err)
	}
	defer lastEditQuery.Close()

	blockQuery, err = conn.Prepare(blockQueryTemplate)
	if err != nil {
		ybtools.PanicErr("blockQuery preparation failed with error ", err)
	}
	defer blockQuery.Close()

	userRedirectQuery, err = conn.Prepare(userRedirectQueryTemplate)
	if err != nil {
		ybtools.PanicErr("userRedirectQuery preparation failed with error ", err)
	}
	defer userRedirectQuery.Close()

	cb()
}

// pruneUsersFromList takes a page title, the content of the page, the formatRegex for the users on the page,
// and the configured inactivity timestamp for the page. It then processes the page, pruning the users who it
// needs to remove, and returns the new page content, the number of expired users,
// number of indeffed users, number of renamed users, an array of the expired users,
// and a map of the renamed users (with their old name as the key, and their new name as the value).
func pruneUsersFromList(
	pageTitle string, pageContent string, formatRegex *regexp.Regexp, inactivityTs time.Time, blockTs time.Time) (
	string, int, int, int, []string, map[string]string) {
	var regexBuilder strings.Builder
	var checkedUsers = map[string]bool{}

	var usersToReplace = map[string]string{}
	// in the below map, we use int8 as the key, because we have consts indeffed and inactiveUsers
	var usersToRemove = map[int8][]string{}

	// Remove the (?im) here as it'll later be in a capture group; we add it back in afterwards
	var formatRegexAsString string = strings.TrimPrefix(formatRegex.String(), "(?im)")

	// format in line with https://www.mediawiki.org/wiki/Manual:Timestamp
	var editsSinceStamp string = inactivityTs.Format(mediaWikiTimestampFormat)
	var blockStamp string = blockTs.Format(mediaWikiTimestampFormat)

	for _, match := range formatRegex.FindAllStringSubmatch(pageContent, -1) {
		var outputFromQueryRow string
		// the first regex capture group should always be the username
		if len(match) != 2 {
			ybtools.PanicErr("Match doesn't have exactly one capture group for regex", formatRegex, "- matched:", match)
		}
		var username string = match[1]

		username = strings.Split(username, "/")[0]

		if checkedUsers[username] {
			continue
		}

		checkedUsers[username] = true
		var dbUsername string = usernameCase(username)
		// We have no use whatsoever for the output of this, we just want to see if it errors.
		// That being said, Scan() doesn't let us just pass nothing, so we have to have the
		// slight pain of having a stupid additional variable.
		err := lastEditQuery.QueryRow(dbUsername, editsSinceStamp).Scan(&outputFromQueryRow)

		if err != nil {
			if err == sql.ErrNoRows {
				// they haven't edited in the timeframe, or they have redirected
				// check a redirect for them
				// annoyingly, unlike the user database rows, the page rows have spaces as underscores,
				// so we have to use this with the inverse replacement we use for all the other checks
				err := userRedirectQuery.QueryRow(strings.ReplaceAll(dbUsername, " ", "_")).Scan(&outputFromQueryRow)
				if err == sql.ErrNoRows {
					// No redirect found, just remove them
					log.Println("Queuing", username, "on title", pageTitle, "for pruning")
					usersToRemove[inactiveUsers] = append(usersToRemove[inactiveUsers], username)
					continue
				} else if err != nil {
					ybtools.PanicErr("Failed when querying DB for redirects with error ", err)
				}
				// A redirect was found! Replace underscores with spaces, and if it redirects to a subpage,
				// get the corresponding root page, before a /, for the user (which is, after all, their username).
				outputFromQueryRow = strings.SplitN(strings.ReplaceAll(outputFromQueryRow, "_", " "), "/", 2)[0]
				log.Println("Found a redirect for", username, "so replacing them on", pageTitle, "with", outputFromQueryRow)
				usersToReplace[username] = outputFromQueryRow
				// this is here to make sure that the redirect target is also checked for indefs
				dbUsername = outputFromQueryRow
			} else {
				ybtools.PanicErr("Failed when querying DB for last edits with error ", err)
			}
		}

		// if they still aren't being pruned, check whether they're indefinitely blocked
		err = blockQuery.QueryRow(dbUsername, blockStamp).Scan(&outputFromQueryRow)
		if err == nil {
			// the user is indeffed, as a row has been found
			log.Println("Queuing indeffed user", username, "on title", pageTitle, "for pruning")
			usersToRemove[indeffedUsers] = append(usersToRemove[indeffedUsers], username)
			continue
		} else if err != sql.ErrNoRows {
			log.Fatal("Failed when querying DB for blocks with error ", err)
		}
	}

	/*

		What happens below is a fair bit of regexing. To summarise what's going on:
			1)	We start with a format regex, with the username as a capture group. For example, /{{test\|([^|]*)|example}}/
				to match {{test|username|example}}.
			2)	We create a new regex, builtRegexForRemoval, using regexBuilder. This regex is the format regex, but with the
				capture group (the username selection part) replaced with a list of usernames we want to remove, in (1|2|3) format.
				This might look like /{{test\|(user1|user2|user3)|example}}/. We then replace all occurrences of this with "".
			3)	We then iterate over our users to rename, and use the same trick again to create a third regex, builtRegexForRename,
				using regexRenameBuilder. This regex only matches a single user, so we have to build separate ones for each user,
				but it then allows us to swap out the username wherever it appears for the new username.

		Whew.

	*/

	usersToRemoveTotal := len(usersToRemove[indeffedUsers]) + len(usersToRemove[inactiveUsers])
	if usersToRemoveTotal > 0 {
		regexBuilder.WriteString(`(?im)`) // Always case-insensitive and multiline

		regexUsersToRemove := make([]string, usersToRemoveTotal)
		var i int
		for _, usersForRemoval := range [][]string{usersToRemove[indeffedUsers], usersToRemove[inactiveUsers]} {
			for _, user := range usersForRemoval {
				// Each of these strings needs to be regex-escaped (hence QuoteMeta),
				// but we're also using them in ReplaceAllString, so $ needs to be escaped as $$ too.
				regexUsersToRemove[i] = strings.ReplaceAll(regexp.QuoteMeta(user), "$", "$$")
				i++
			}
		}

		regexBuilder.WriteString(
			// From within the format regex, replace the capture group (there should only be one, per spec)
			// plus the first character before the capture group (see regexReplaceCaptureGroup's comment)
			// with the first character, followed by the usersToRemove list, separated by pipes, in a group
			// (i.e. "match any one of the usernames given" in place of "find a username")
			regexReplaceCaptureGroup.ReplaceAllString(
				formatRegexAsString, ("$1(" + strings.Join(regexUsersToRemove, "|") + ")")))

		regexBuilder.WriteString(`\n?`) // optionally match a newline at the end, to not leave empty space

		builtRegexForRemoval, err := regexp.Compile(regexBuilder.String())
		if err != nil {
			ybtools.PanicErr("Failed to build builtRegexForRemoval from regexBuilder with error ", err)
		}

		pageContent = builtRegexForRemoval.ReplaceAllString(pageContent, "")
	}

	if len(usersToReplace) > 0 {
		captureGroupStringIndex := regexReplaceCaptureGroup.FindStringIndex(formatRegexAsString)
		for old, new := range usersToReplace {
			var regexRenameBuilder strings.Builder

			regexRenameBuilder.WriteString("(?im)(")
			// Write everything up to the start of the capture group, in a capture group this time.
			// Note the +1; this is because regexReplaceCaptureGroup also matches the character
			// before the capture group, as discussed earlier.
			regexRenameBuilder.WriteString(formatRegexAsString[:captureGroupStringIndex[0]+1])
			regexRenameBuilder.WriteString(")")
			regexRenameBuilder.WriteString(regexp.QuoteMeta(old))
			regexRenameBuilder.WriteString("(")
			// Write the rest of the regex, in a capture group again.
			regexRenameBuilder.WriteString(formatRegexAsString[captureGroupStringIndex[1]:])
			regexRenameBuilder.WriteString(")")

			builtRegexForRename, err := regexp.Compile(regexRenameBuilder.String())
			if err != nil {
				ybtools.PanicErr("Failed to build builtRegexForRename from regexRenameBuilder with error ", err)
			}

			// Generate the replacement that we want to use
			b := builtRegexForRename.ExpandString(
				[]byte{},
				"${1}"+strings.ReplaceAll(new, "$", "$$")+"${2}",
				pageContent,
				builtRegexForRename.FindStringSubmatchIndex(pageContent),
			)

			// Now we actually replace each occurrence, putting the relevant bits in the relevant places.
			pageContent = builtRegexForRename.ReplaceAllStringFunc(pageContent, func(string) string {
				// replace User talk links within the match too
				return strings.ReplaceAll(
					// replace very simple user link confusions too; if they're more complicated, for instance a
					// funky styled signature, we won't get into that as it all gets a bit complicated then.
					// the links will still work at least
					strings.ReplaceAll(string(b), "[[User:"+new+"|"+old+"]]", "[[User:"+new+"|"+new+"]]"),
					"User talk:"+old,
					"User talk:"+new,
				)
			})
		}
	}

	return pageContent, len(usersToRemove[inactiveUsers]), len(usersToRemove[indeffedUsers]), len(usersToReplace),
		usersToRemove[inactiveUsers], usersToReplace
}

// Takes a string, s, and converts the first rune to uppercase
// by using the unicode functions in golang's stdlib
// Returns the converted string
func usernameCase(s string) string {
	firstRune, size := utf8.DecodeRuneInString(s)
	if firstRune != utf8.RuneError || size > 1 {
		var upcase rune

		if replacement, exists := charReplaceUpcase[firstRune]; exists {
			upcase = replacement
		} else {
			upcase = unicode.ToUpper(firstRune)
		}
		if upcase != firstRune {
			s = string(upcase) + s[size:]
		}
	}
	return strings.ReplaceAll(s, "_", " ")
}
