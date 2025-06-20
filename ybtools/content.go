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

	"github.com/antonholmquist/jason"
)

// GetContentFromPage turns a *jason.Object for a Page into the main slot content,
// and/or an applicable error.
//
// Yes, there is a function to do this already in the library (GetPagesByName).
// No, I don't want to use it. Why? We've already got the page content here -
// making another request to get it again is wasteful when we could just locally
// parse what we already have.
func GetContentFromPage(page *jason.Object) (content string, err error) {
	rev, err := page.GetObjectArray("revisions")
	if err != nil {
		log.Println("Failed to get revisions from page, erroring GetContentFromPage. Error was ", err)
		return "", err
	}
	return GetMainSlotFromRevision(rev[0])
}

// GetPagesFromQuery takes a query and returns an array of Pages.
// Convenience wrapper for GetThingFromQuery.
func GetPagesFromQuery(resp *jason.Object) []*jason.Object {
	pages, err := GetThingFromQuery(resp, "pages")
	if err != nil {
		panic(err)
	}
	return pages
}

// GetThingFromQuery takes a query and a key that's being looked for,
// and returns the inner thing array.
func GetThingFromQuery(resp *jason.Object, thing string) ([]*jason.Object, error) {
	query, err := resp.GetObject("query")
	if err != nil {
		switch err.(type) {
		case jason.KeyNotFoundError:
			// no query means no results
			return []*jason.Object{}, nil
		default:
			return nil, err
		}
	}
	pages, err := query.GetObjectArray(thing)
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// GetMainSlotFromRevision fetches the main slot content, given a revision object
// formatted per the MediaWiki Action API JSON.
func GetMainSlotFromRevision(revision *jason.Object) (string, error) {
	content, err := revision.GetString("slots", "main", "content")
	if err != nil {
		log.Println("Failed to get main slot content from page, erroring GetMainSlotFromRevision. Error was", err)
		return "", err
	}
	return content, nil
}

// GetCategorisationTimestampFromPage takes a page,
// and gets the timestamp at which the page was categorised.
// All the errors in this function are Fatal, because frankly,
// if something's gone wrong with the timestamp reading, we're not really
// going to be able to run the algorithm correctly anyway.
func GetCategorisationTimestampFromPage(page *jason.Object, category string) (timestamp string) {
	itemCategories, err := page.GetObjectArray("categories")
	if err != nil {
		PanicErr("Failed to get categories with error message ", err)
	}
	relevantCategory := itemCategories[0]

	timestamp, err = relevantCategory.GetString("timestamp")
	if err != nil {
		PanicErr("Failed to get categorisation timestamp with error message ", err)
	}
	return
}
