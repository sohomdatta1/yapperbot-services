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

// BotSettings is a struct storing all the information about the bot
// needed to make the tools library work.
type BotSettings struct {
	TaskName         string
	BotUser          string
	ToolforgeAccount string
}

var settings BotSettings

// SetupBot sets the bot name, ready for future calls to BotAllowed.
func SetupBot(s BotSettings) {
	settings = s
	setupNobotsBot()
	setupTaskConfigFile()
	setKillPage()
	// Kill pages are checked as soon as the mwclient is first authenticated
}

// CanEdit checks if the task has been killed, and then checks if the
// bot is edit limited. This *must* be used for all edits apart from
// those which only affect the bot's own userspace, or those which must
// run even if the task is killed (e.g. updating JSON files which describe
// the run which was just done).
func CanEdit() bool {
	killTaskIfNeeded()
	return EditLimit()
}
