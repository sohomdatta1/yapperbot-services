package ybtools

//
// Yapperbot Tools, the internal system bits for Yapperbot and co.
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

import (
	"fmt"
	"regexp"
)

var botBanRegex *regexp.Regexp
var botWhitelistRegex *regexp.Regexp
var botAllowThisBotRegex *regexp.Regexp

const botBanRegexTemplate string = `{{nobots}}|{{bots\|deny=(?:[^,|}]*,)*%[1]s`
const botWhitelistRegexTemplate string = `{{bots\|allow=`
const botAllowRegexTemplate string = `{{bots\|allow=(?:[^}|]*,)*%[1]s[},|]`

func init() {
	botWhitelistRegex = regexp.MustCompile(botWhitelistRegexTemplate)
}

// BotAllowed take a page content and determines if the botUser is allowed
// to edit the page per the applicable templates.
func BotAllowed(pageContent string) bool {
	if settings.BotUser == "" {
		panic("BotAllowed called with no botUser set!")
	}

	// this mess below is only necessary because Go doesn't support regex lookaheads
	// which I know is because there are problems in terms of time guarantees for them
	// but ugh.
	if botWhitelistRegex.MatchString(pageContent) {
		// the page contains a whitelist, we need to check if we're on it
		return botAllowThisBotRegex.MatchString(pageContent)
	}

	// the page doesn't contain a whitelist, return true if we're not blacklisted or nobotted
	return !botBanRegex.MatchString(pageContent)
}

// setupNobotsBot sets the bot name, ready for future calls to BotAllowed.
func setupNobotsBot() {
	botBanRegex = regexp.MustCompile(fmt.Sprintf(botBanRegexTemplate, settings.BotUser))
	botAllowThisBotRegex = regexp.MustCompile(fmt.Sprintf(botAllowRegexTemplate, settings.BotUser))
}
