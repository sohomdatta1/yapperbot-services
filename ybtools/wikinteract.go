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
	"log"

	"cgt.name/pkg/go-mwclient"
	"cgt.name/pkg/go-mwclient/params"
)

// NoMaxlagFunction is the definition of a function accepted by NoMaxlagDo;
// it's just a function with no parameters that returns an error.
type NoMaxlagFunction func() error

// PageInQueryCallback is a function used as a callback for ForPageInQuery.
type PageInQueryCallback func(pageTitle, pageContent, pageContentModel, revTS, curTS string)

var w *mwclient.Client

// DefaultMaxlag is a Maxlag representing sensible defaults for any non-urgent
// task being run by a Yapperbot.
var DefaultMaxlag mwclient.Maxlag = mwclient.Maxlag{
	On:      true,
	Timeout: "5",
	Retries: 3,
}

// NoMaxlag is a Maxlag for urgent tasks, where maxlag must be disabled.
var NoMaxlag mwclient.Maxlag = mwclient.Maxlag{
	On:      true,
	Timeout: "5",
	Retries: 3,
}

// CreateAndAuthenticateClient uses the details already passed into ybtools
// in setup.go to return a fully-authenticated mwclient
func CreateAndAuthenticateClient(maxlag mwclient.Maxlag) *mwclient.Client {
	if settings.TaskName == "" || settings.BotUser == "" {
		PanicErr("Call ybtools.SetupBot first!")
	}

	var err error

	w, err = mwclient.New(config.APIEndpoint, "Yapperbot-"+settings.TaskName+" on User:"+settings.BotUser+" - Golang, licensed GNU GPL")
	if err != nil {
		PanicErr("Failed to create MediaWiki client with error ", err)
	}

	// This is necessary because maxlag.sleep is unexported,
	// and is only configured correctly for production within
	// mwclient.New. Annoying, but this is a bit of an edge case /shrug
	w.Maxlag.On = maxlag.On
	w.Maxlag.Retries = maxlag.Retries
	w.Maxlag.Timeout = maxlag.Timeout

	err = w.Login(config.BotUsername, botPassword)
	if err != nil {
		PanicErr("Failed to authenticate with MediaWiki with username ", config.BotUsername, " - error was ", err)
	}

	// runs here to make sure we have a client authenticated when we run it
	killTaskIfNeeded()

	return w
}

// NoMaxlagDo takes a function which returns an error (or nil),
// and an mwclient pointer, than executes that function with no maxlag.
// It returns the same return as the NoMaxlagFunction it's passed.
func NoMaxlagDo(f NoMaxlagFunction, w *mwclient.Client) error {
	w.Maxlag.On = false
	err := f()
	w.Maxlag.On = true
	return err
}

// FetchWikitext takes a pageId and gets the wikitext of that page.
// The default functionality in the library does not work for this in
// my experience; it just returns an empty string for some reason. So we're rolling our own!
func FetchWikitext(pageID string) (content string, err error) {
	content, _, _, err = fetchWikitextFrom("pageids", pageID)
	return
}

// FetchWikitextWithTimestamps takes a pageId and gets the wikitext of that page,
// also returning the revision timestamp and the current timestamp.
func FetchWikitextWithTimestamps(pageID string) (content string, revtimestamp string, curtimestamp string, err error) {
	return fetchWikitextFrom("pageids", pageID)
}

// FetchWikitextFromTitle takes a title and gets the wikitext of that page.
func FetchWikitextFromTitle(pageTitle string) (content string, err error) {
	content, _, _, err = fetchWikitextFrom("titles", pageTitle)
	return
}

// FetchWikitextFromTitleWithTimestamps takes a title and gets the wikitext of that page,
// also returning the revision timestamp and the current timestamp.
func FetchWikitextFromTitleWithTimestamps(pageTitle string) (content string, revtimestamp string, curtimestamp string, err error) {
	return fetchWikitextFrom("titles", pageTitle)
}

// ForPageInQuery takes parameters and a callback function. It then queries using the parameters it is given,
// and calls the callback function for every page in the query response.
func ForPageInQuery(parameters params.Values, callback PageInQueryCallback) {
	query := w.NewQuery(parameters)
	for query.Next() {
		pages := GetPagesFromQuery(query.Resp())

		curTS, err := query.Resp().GetString("curtimestamp")
		if err != nil {
			PanicErr("Failed to get current timestamp! Error was", err)
		}

		if len(pages) > 0 {
			for _, page := range pages {
				pageTitle, err := page.GetString("title")
				if err != nil {
					log.Println("Failed to get title from page, so skipping it. Error was", err)
					continue
				}

				if _, err := page.GetValue("missing"); err == nil {
					log.Printf("Page `%s` is missing, so skipping it: probably deleted. Error was %s\n", pageTitle, err)
					continue
				}

				pageRevisions, err := page.GetObjectArray("revisions")
				if err != nil {
					log.Printf("Failed to get revisions array from page `%s`, so skipping it. Error was %s\n", pageTitle, err)
					continue
				}

				pageContent, err := GetMainSlotFromRevision(pageRevisions[0])
				if err != nil {
					log.Printf("Failed to get content from page `%s`, so skipping it. Error was %s\n", pageTitle, err)
					continue
				}

				lastTimestamp, err := pageRevisions[0].GetString("timestamp")
				if err != nil {
					log.Printf("Failed to get timestamp from revision on page `%s`, so skipping it. Error was %s\n", pageTitle, err)
					continue
				}

				pageContentModel, err := pageRevisions[0].GetString("slots", "main", "contentmodel")
				if err != nil {
					log.Printf("Failed to get contentmodel from revision on page `%s` so skipping it. Error was %s\n", pageTitle, err)
				}

				callback(pageTitle, pageContent, pageContentModel, lastTimestamp, curTS)
			}
		}
	}
}

// fetchWikitextFrom takes an identifier name (i.e. pageids or titles), and one of those identifiers,
// and then returns the wikitext, the revision timestamp, the current timestamp, and an error.
func fetchWikitextFrom(identifierName string, identifier string) (string, string, string, error) {
	queryResult, err := w.Get(params.Values{
		"action":       "query",
		identifierName: identifier,
		"prop":         "revisions",
		"curtimestamp": "1",
		"rvprop":       "timestamp|content",
		"rvslots":      "main",
	})
	if err != nil {
		return "", "", "", err
	}

	curtimestamp, err := queryResult.GetString("curtimestamp")
	if err != nil {
		return "", "", "", err
	}

	pages := GetPagesFromQuery(queryResult)
	if len(pages) < 1 {
		return "", "", "", mwclient.ErrPageNotFound
	}

	rev, err := pages[0].GetObjectArray("revisions")
	if err != nil {
		return "", "", "", err
	}

	revtimestamp, err := rev[0].GetString("timestamp")
	if err != nil {
		return "", "", "", err
	}

	text, err := GetMainSlotFromRevision(rev[0])
	return text, revtimestamp, curtimestamp, err
}
