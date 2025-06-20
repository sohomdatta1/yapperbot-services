package ybtools

import (
	"cgt.name/pkg/go-mwclient"
)

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

const killPageNamespace string = "User:"
const killPagePrefix string = "/kill/"

var killPage string

func setKillPage() {
	killPage = killPageNamespace + settings.BotUser + killPagePrefix + settings.TaskName
}

// killTaskIfNeeded checks the killPage for the bot task, and
func killTaskIfNeeded() {
	wt, err := FetchWikitextFromTitle(killPage)
	if err != nil {
		// don't panic if the page is just missing - all that means is that nobody has created it
		if typedErr, ok := err.(mwclient.APIError); ok && typedErr.Code != "missingtitle" {
			// do the panic - something is deeply wrong
			PanicErr("Killed - task kill page couldn't be fetched at ", killPage, " with error ", typedErr.Info)
		}
	}
	if wt != "" {
		// page not empty, kill it!
		PanicErr("Killed - task kill page not empty at ", killPage)
	}
}
