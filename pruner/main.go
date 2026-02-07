package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"cgt.name/pkg/go-mwclient"
	"cgt.name/pkg/go-mwclient/params"
	"github.com/karrick/tparse"
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

const templateExpression string = `\s*?((?:\|.*\s*?)*)`
const editSummaryOpening string = `Pruning users as configured on page: processed `
const editSummaryUsersExpired string = `%d inactive user(s)`
const editSummaryUsersIndeffed string = `%d indeffed user(s)`
const editSummaryUsersRenamed string = `%d renamed user(s)`

var templateRegex *regexp.Regexp
var formats = map[string]*regexp.Regexp{}

func init() {
	ybtools.SetupBot(ybtools.BotSettings{
		TaskName:         "Pruner",
		BotUser:          "SodiumBot",
		ToolforgeAccount: "yapping-sodium",
	})
	ybtools.ParseTaskConfig(&config)
}

func main() {
	defer ybtools.SaveEditLimit()

	templateRegex = regexp.MustCompile("{{" + regexp.QuoteMeta(config.ConfigTemplate) + templateExpression + "}}")

	w := ybtools.CreateAndAuthenticateClient(ybtools.DefaultMaxlag)

	formatsJSON := ybtools.LoadJSONFromPageID(config.FormatsJSONPageID)

	for name, regex := range formatsJSON.Map() {
		rString, err := regex.String()
		if err != nil {
			ybtools.PanicErr("Failed to decode regex from formatsJSON with error", err)
		}
		// all these regexes should be case-insensitive and multiline; set this flag on them all
		rCompiled, err := regexp.Compile("(?im)" + rString)
		if err != nil {
			ybtools.PanicErr("Failed to compile regex", rString, "from formatsJSON with error", err)
		}
		formats[name] = rCompiled
	}

	// processArticleInitial just serves as a wrapper around processArticle;
	// the only reason it is here is to set retry to false on the first run,
	// so we can use ybtools ForPageInQuery easily.
	var processArticleInitial = func(pageTitle, pageContent, pageContentModel, revTS, curTS string) {
		log.Println("Processing page", pageTitle)
		processArticle(w, pageTitle, pageContent, pageContentModel, revTS, curTS, false)
	}

	withDatabaseConnection(func() {
		ybtools.ForPageInQuery(params.Values{
			"action":         "query",
			"prop":           "revisions",
			"generator":      "embeddedin",
			"geititle":       config.ConfigTemplate,
			"geifilterredir": "nonredirects",
			"rvprop":         "timestamp|content|contentmodel",
			"rvslots":        "main",
			"curtimestamp":   "1",
		}, processArticleInitial)
	})
}

func enumeratePagePrunerConfig(pageTitle string, pageContent string) (time.Time, time.Time, string, map[string]string, error) {
	match := templateRegex.FindStringSubmatch(pageContent)

	if len(match) == 0 {
		log.Println(pageTitle, "included in transclusions of template, but didn't match regex!")
		return time.Time{}, time.Time{}, "none", map[string]string{}, errors.New("Does not have a valid template configuration")
	}

	var parameters = map[string]string{}
	// only one capture group; zero is the entire match
	// we trim off the first character, as it's a useless pipe always, by definition
	paramStrings := strings.Split(match[1][1:], "|")
	for _, paramString := range paramStrings {
		paramComponents := strings.Split(paramString, "=")
		parameters[strings.TrimSpace(paramComponents[0])] = strings.TrimSpace(paramComponents[1])
	}

	// check required parameters
	for _, param := range []string{"inactivity", "format"} {
		if _, ok := parameters[param]; !ok {
			log.Println(pageTitle, "has an invalid configuration: missing", param)
			return time.Time{}, time.Time{}, "none", map[string]string{}, errors.New("Does not have a valid template configuration")
		}
	}

	// tparse is totally fine with processing things like "year" and "years", but wants "1year" not "1 year" for some reason...
	inactivityTimestamp, err := tparse.AddDuration(time.Now(), "-"+strings.ReplaceAll(parameters["inactivity"], " ", ""))
	if err != nil {
		log.Println(pageTitle, "has an invalid inactivity timeout value, of", parameters["inactivity"], "- error was", err)
		return time.Time{}, time.Time{}, "none", map[string]string{}, errors.New("Does not have a valid template configuration")
	}

	var blockTimestamp time.Time
	// If a parameter `indeffed` is specified, parse that as the blockTimestamp
	if parameters["indeffed"] != "" {
		if parameters["indeffed"] == "0" {
			// special case: immediately.
			// this is a special case because tparse doesn't currently support unitless zero as a duration
			// https://github.com/karrick/tparse/issues/2
			blockTimestamp = time.Now()
		} else {
			blockTimestamp, err = tparse.AddDuration(time.Now(), "-"+strings.ReplaceAll(parameters["indeffed"], " ", ""))
			if err != nil {
				log.Println(pageTitle, "has an invalid indeffed time value of", parameters["indeffed"], "so using default")
			}
		}
	}

	if blockTimestamp.Equal((time.Time{})) {
		// Otherwise, default to removing blocked users after two months
		blockTimestamp = time.Now().AddDate(0, -2, 0)
	}

	return blockTimestamp, inactivityTimestamp, parameters["format"], parameters, nil
}

func processArticle(w *mwclient.Client, pageTitle, pageContent, pageContentModel, revTS, curTS string, retry bool) {
	var (
		blockTimestamp                      time.Time
		inactivityTimestamp                 time.Time
		format                              string
		err                                 error
		numExpired, numIndeffed, numRenamed int
		newPageContent                      string
		expiredUsers                        []string
		parameters                          map[string]string
	)

	switch pageContentModel {
	case "wikitext":
		blockTimestamp, inactivityTimestamp, format, parameters, err = enumeratePagePrunerConfig(pageTitle, pageContent)

		if err != nil {
			log.Printf("Unable to proceed further with `%s` due to the following error `%s`", pageTitle, pageContent)
			return
		}

		formatRegex, ok := formats[format]
		if !ok {
			log.Println(pageTitle, "has an invalid format value, of", format)
			return
		}

		newPageContent, numExpired, numIndeffed, numRenamed, expiredUsers, _ = pruneUsersFromWikitextList(pageTitle, pageContent, formatRegex, inactivityTimestamp, blockTimestamp)
	case "MassMessageListContent":
		var parsedPageContent MassMessageContent
		err := json.Unmarshal([]byte(pageContent), &parsedPageContent)

		if err != nil {
			log.Printf("Unable to correctly parse the contentmodel of the mass message list `%s`, the error is: %s", pageTitle, err)
		}

		blockTimestamp, inactivityTimestamp, format, parameters, err = enumeratePagePrunerConfig(pageTitle, parsedPageContent.Description)

		newPageContent, numExpired, numIndeffed, numRenamed, expiredUsers, _ = pruneUsersFromMMList(pageTitle, parsedPageContent, inactivityTimestamp, blockTimestamp)
	default:
		log.Printf("Incorrect contentmodel, unable to proceed further on `%s`", pageTitle)
	}

	if newPageContent == pageContent {
		log.Println("newPageContent was the same as pageContent on page", pageTitle, "so ignoring")
		return
	}

	var editSummaryBuilder strings.Builder

	editSummaryBuilder.WriteString(editSummaryOpening)

	var summaryActionsTaken []string
	if numExpired != 0 {
		summaryActionsTaken = append(summaryActionsTaken, fmt.Sprintf(editSummaryUsersExpired, numExpired))
	}
	if numIndeffed != 0 {
		summaryActionsTaken = append(summaryActionsTaken, fmt.Sprintf(editSummaryUsersIndeffed, numIndeffed))
	}
	if numRenamed != 0 {
		summaryActionsTaken = append(summaryActionsTaken, fmt.Sprintf(editSummaryUsersRenamed, numRenamed))
	}

	editSummaryBuilder.WriteString(strings.Join(summaryActionsTaken, "; "))

	if !ybtools.CanEdit() {
		return
	}

	err = w.Edit(params.Values{
		"title":          pageTitle,
		"text":           newPageContent,
		"md5":            fmt.Sprintf("%x", md5.Sum([]byte(newPageContent))),
		"summary":        editSummaryBuilder.String(),
		"notminor":       "true",
		"bot":            "true",
		"basetimestamp":  revTS,
		"starttimestamp": curTS,
	})

	if err == nil {
		log.Println("Pruned users on", pageTitle, "so starting notifications")
		var userMessages = map[string]string{}

		expiredMsg, ok := parameters["expiredmsg"]
		if !ok || expiredMsg == "" {
			expiredMsg = config.DefaultExpiredMsgTemplate
		}

		if expiredMsg != "none" {
			for _, user := range expiredUsers {
				userMessages[user] = "{{subst:" + strings.Join([]string{expiredMsg, user, pageTitle, parameters["inactivity"]}, "|") + "}}\n <small>(replacing <span class=\"plainlinks\">[https://en.wikipedia.org/wiki/User:Yapperbot Yapperbot]</span>)</small> ~~~~"
			}

			talkMessageHeader, ok := parameters["talkmsgheader"]
			if !ok || talkMessageHeader == "" {
				talkMessageHeader = config.DefaultTalkMsgHeader
			}

			for user, message := range userMessages {
				userPage, _ := ybtools.FetchWikitextFromTitle("User talk:" + user)
				if ybtools.BotAllowed(userPage) && ybtools.CanEdit() {
					err := w.Edit(params.Values{
						"title":        "User talk:" + user,
						"section":      "new",
						"sectiontitle": talkMessageHeader,
						"summary":      "[[User:Yapperbot/Pruner|Pruner]]: " + talkMessageHeader,
						"notminor":     "true",
						"bot":          "true",
						"text":         message,
						"redirect":     "true",
					})
					if err == nil {
						log.Println("Successfully notified", user, "of their pruning from", pageTitle)
						time.Sleep(5 * time.Second)
					} else {
						switch err := err.(type) {
						case mwclient.APIError:
							switch err.Code {
							case "noedit", "writeapidenied", "blocked":
								ybtools.PanicErr("noedit/writeapidenied/blocked code returned, the bot may have been blocked. Dying")
							default:
								log.Println("Error editing user talk for", user, "meant they couldn't be notified. The error was", err)
							}
						default:
							ybtools.PanicErr("Non-API error returned when trying to notify user ", user, " so dying. Error was ", err)
						}
					}
				}
			}
		}
	} else {
		switch err := err.(type) {
		case mwclient.APIError:
			if err.Code == "editconflict" {
				if retry {
					log.Println("Edit conflicted twice on page", pageTitle, "so skipping")
					return
				}

				log.Println("Edit conflicted on page", pageTitle, "so refetching")
				fetchedContent, revTS, curTS, err := ybtools.FetchWikitextFromTitleWithTimestamps(pageTitle)
				if err != nil {
					log.Println("Returned an error when trying to refetch article, skipping:", err)
					return
				}

				processArticle(w, pageTitle, fetchedContent, pageContentModel, revTS, curTS, true)
				return
			}

			ybtools.PanicErr("API error raised, can't handle, so failing. Error was ", err)
		default:
			if err == mwclient.ErrEditNoChange {
				// Normally this would only happen if two instances ran at the same time or very close to one another.
				// It may also happen in the case of a very long-running process, that misses an update to a page in the mean time.
				// It's very rare, but theoretically possible.
				log.Println("No change made to page", pageTitle, "so assuming something already fixed it and ignoring")
				return
			}
			ybtools.PanicErr("Non-API error raised, can't handle, so failing. Error was ", err)
		}
	}
}
