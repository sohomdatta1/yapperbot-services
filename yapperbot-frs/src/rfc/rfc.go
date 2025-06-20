package rfc

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

import "regexp"

// rfcPrefixRegex is a regex that matches the comments at the start of each RfC line;
// these are in the form <!--rfc:categoryname-->, and mean the bot can match it.
var rfcPrefixRegex *regexp.Regexp

const requestType string = "request for comment"

// An RfC has an id, a categories map and a setting for whether feedback has been given for it.
// The map should be map[string]bool, with bool as true for every element
// This is so membership verification is o(1) rather than o(n)
type RfC struct {
	ID           string
	Categories   map[string]bool
	FeedbackDone bool
	PageHolding  string
}

func init() {
	rfcPrefixRegex = regexp.MustCompile(`<!--rfc:(\w*?)-->`)
}

// IncludeHeader determines if a given FRS header corresponds to this item correctly
// Takes a string of the entire header (minus the === bits) and returns a bool for
// if the header is included, and separately a bool indicating whether the header is the all
// header or not
func (r RfC) IncludeHeader(header string) (bool, bool) {
	matches := rfcPrefixRegex.FindStringSubmatch(header)
	if matches == nil {
		// no matches means it's not an RfC
		return false, false
	}

	// check for special keyword "all"
	if matches[1] == "all" {
		return true, true
	}
	// check if in categories
	_, exists := r.Categories[matches[1]]
	return exists, false
}

// PageTitle is a simple getter for the HoldingPage in order to make the interface work
func (r RfC) PageTitle() string {
	return r.PageHolding
}

// RequestType returns the type this is - an RfC - so that it can be used in a template
func (r RfC) RequestType() string {
	return requestType
}
