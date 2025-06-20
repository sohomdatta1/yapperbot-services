package ybtools

import (
	"encoding/json"

	"github.com/antonholmquist/jason"
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

// SerializeToJSON takes in any serializable object and returns the serialized JSON string
func SerializeToJSON(serializable interface{}) string {
	serialized, err := json.Marshal(serializable)
	if err != nil {
		PanicErr("Failed to serialize object, dumping what I was trying to serialize: ", serializable)
	}
	return string(serialized)
}

// LoadJSONFromPageID takes a pageID, then loads and deserializes the contained JSON.
// It returns the deserialised JSON in a jason.Object pointer.
func LoadJSONFromPageID(pageID string) *jason.Object {
	storedJSON, err := FetchWikitext(pageID)
	if err != nil {
		PanicErr("Failed to fetch JSON page with ID ", pageID, " with error ", err)
	}
	return parseJSON(storedJSON, "Failed to parse JSON on page ID "+pageID+" with error ")
}

// LoadJSONFromPageTitle takes a title string, then loads and deserializes the contained JSON.
// It returns the deserialised JSON in a jason.Object pointer.
func LoadJSONFromPageTitle(pageTitle string) *jason.Object {
	storedJSON, err := FetchWikitextFromTitle(pageTitle)
	if err != nil {
		PanicErr("Failed to fetch JSON page with title ", pageTitle, " with error ", err)
	}
	return parseJSON(storedJSON, "Failed to parse JSON on page "+pageTitle+" with error ")
}

func parseJSON(contentToParse string, errorMsg string) *jason.Object {
	parsedJSON, err := jason.NewObjectFromBytes([]byte(contentToParse))
	if err != nil {
		PanicErr(errorMsg, err)
	}
	return parsedJSON
}
