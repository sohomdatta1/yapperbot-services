package main

//
// Yapperbot-FRS, the Feedback Request Service bot for Wikipedia
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
	"errors"
	"log"
	"math/rand"
	"strings"

	"github.com/sohomdatta1/yapperbot-services/frs/src/frslist"
	"github.com/sohomdatta1/yapperbot-services/frs/src/messages"
	"github.com/sohomdatta1/yapperbot-services/frs/src/rfc"

	"cgt.name/pkg/go-mwclient"
)

const maxMsgsToSend int = 15
const minMsgsToSend int = 5

// requestFeedbackFor takes an object that implements frsRequesting and a mwclient instance,
// and processes the feedback request for the frsRequesting object.
func requestFeedbackFor(requester frsRequesting, w *mwclient.Client) (err error) {
	// msgsToSend is a randomly-selected number of messages we want to send out.
	// it evaluates out to any number between max and min
	var msgsToSend int = (rand.Intn(maxMsgsToSend-minMsgsToSend) + minMsgsToSend)

	// headersToSendTo will be our slice of headers that we want to consider users in.
	// it's important that this is a separate array, as we later consider its length
	var headersToSendTo []string

	// allHeader is the header that's "catch all" for all RfCs/GA noms/whatever, where applicable
	var allHeader string

	for _, header := range frslist.GetListHeaders() {
		include, isAllHeader := requester.IncludeHeader(header)
		if include {
			headersToSendTo = append(headersToSendTo, header)
			messages.CleanHeader(header)
		}
		if isAllHeader {
			allHeader = header
		}
	}

	var rfcid string
	if rfc, isRfC := requester.(rfc.RfC); isRfC {
		rfcid = rfc.ID
	}

	// Early return if the page is in userspace, prevent
	// https://en.wikipedia.org/wiki/Wikipedia:Bots/Noticeboard#c-Dw31415-20260109145700-Dw31415_-_DwAlphaBot_-_SodiumBot_conflict_on_RfCHistory
	if strings.HasPrefix(requester.PageTitle(), "User:") {
		return nil
	}

	if len(headersToSendTo) > 0 {
		users := frslist.GetUsersFromHeaders(headersToSendTo, allHeader, msgsToSend)
		for _, user := range users {
			messages.QueueMessage(&messages.Message{
				User:  user,
				Type:  requester.RequestType(),
				Title: requester.PageTitle(),
				RFCID: rfcid,
			})
			log.Println("Queued a message for", user.Username, "to give feedback on", requester.PageTitle(), "in", user.Header)
		}
		return nil
	} else {
		log.Println("WARNING: Headers to send to returned as less than one for page", requester.PageTitle(), "so ignoring for now, but this could be a bug")
		return errors.New("Did not find a header for " + requester.PageTitle() + " no messages were sent for that page")
	}
}
