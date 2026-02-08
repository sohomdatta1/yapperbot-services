package yapperconfig

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

// configObject is the local implementation of ybtools' configuration.
// each of these keys are just pulled straight from the config-frs yml file
// in the application directory by ybtools.
// this doesn't include EditLimit, which is handled by ybtools directly
type configObject struct {
	FRSPageID                string
	SentCountPageID          string
	GAGuidelinesHeaderPageID string
	RFCsDonePageID           string
	ErrorsPageID             string
}

// Config is the global configuration object. This should only really ever be read from.
var Config configObject
