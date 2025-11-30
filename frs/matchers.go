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
	"fmt"
	"regexp"
	"strings"

	"github.com/sohomdatta1/yapperbot-services/frs/src/ga"
	"github.com/sohomdatta1/yapperbot-services/frs/src/rfc"
	"github.com/sohomdatta1/yapperbot-services/ybtools"
)

// rfcMatcher is a regex that matches {{rfc}} templates on pages.
// Its contents are documented in matchers.go:init().
var rfcMatcher *regexp.Regexp

// gaMatcher is a regex that matches {{GA nominee}} templates on pages.
// Its contents are documented in matchers.go:init().
var gaMatcher *regexp.Regexp

// namedParamMatcher is a regex that matches against named parameters in
// a template parameter list; e.g. {{template|name=param}}, matching name=param.
// Its contents are documented in matchers.go:init().
var namedParamMatcher *regexp.Regexp

const rfcIDParam string = "rfcid="

func init() {
	// RfC matching regex.
	// First capture group is all the params of the rfc template
	// (all non-named params are to be treated as categories)
	// second capture group is the actual content of the RfC
	rfcMatcher = regexp.MustCompile(`(?i){{rfc\|(.*?)}}(.|\n)*?\(UTC\)`)

	// GA nom matching regex.
	// First capture group is the topic. Second capture group is the subtopic.
	// IMPORTANT: If the subtopic is empty, the first capture group will be empty string, and likewise for the topic.
	// This regex will always prefer a topic to a subtopic, but will settle for a subtopic if no topic is available.
	// Great thanks go to Ouims from #regex on Freenode for the help with debugging and correcting this regex!
	gaMatcher = regexp.MustCompile(`(?i){{GA nominee(?:\|(?:[^|}]*?\|)*(?:[\t\f\v ]*?(?:subtopic=([^|}]+).*?)|topic=([^|}]+))|.*?)*}}`)

	// Matches against named parameters in the parameter list.
	// Ensures the equals is after a named param specifically.
	// The [\w\d\s] set means it won't trigger on {{=}} and the like.
	namedParamMatcher = regexp.MustCompile(`(?m)^[\w\d\s]*?=`)
}

// extractRfcs takes a string of content containing rfcs, and the page title,
// and returns a slice of rfcs. It can optionally be passed excludeDone, which prevents
// already-done RfCs from being included in the generated list.
// extractRfcs output should be checked for RfCs with no ID string, as those haven't
// yet been assigned an ID by Legobot.
func extractRfcs(content string, title string, excludeDone bool) (rfcs []rfc.RfC, err error) {
	matchedRfcTags := rfcMatcher.FindAllStringSubmatch(content, -1)
	for _, tag := range matchedRfcTags {
		// group 1 contains all the parameters; split on | to find the individual params
		params := strings.Split(tag[1], "|")

		rfcID, categories, feedbackDone := func(pr []string) (id string, cats map[string]bool, done bool) {
			cats = make(map[string]bool)
			for _, p := range pr {
				if strings.Contains(p, rfcIDParam) {
					id = strings.TrimPrefix(p, rfcIDParam)
				} else if !namedParamMatcher.MatchString(p) {
					// it's not a named param, we can assume it's a category
					// for more on why this is like this, see frsRequesting
					cats[p] = true
				}
			}
			if id != "" && rfc.AlreadyDone(id) {
				done = true
			}
			return
		}(params)

		if feedbackDone && excludeDone {
			continue
		} else {
			rfcs = append(rfcs, rfc.RfC{ID: rfcID, Categories: categories, FeedbackDone: feedbackDone, PageHolding: title})
		}
	}
	return
}

// extractGANom takes a page name and content that's been nominated for GA,
// and returns the GA nom object.
func extractGANom(content string, title string) (nom ga.Nom) {
	matchedGaTag := gaMatcher.FindStringSubmatch(content)
	defer func() {
		if r := recover(); r != nil {
			ybtools.PanicErr(fmt.Sprintf("Substring matching for %s failed in matchers.extractGANom", title))
		}
	}()
	// first capture group is name of topic, if applicable
	// second capture group is name of subtopic
	nom = ga.Nom{Topic: matchedGaTag[2], Subtopic: matchedGaTag[1], Article: title}
	return
}
