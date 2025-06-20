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

// frsRequesting is an interface covering all objects that could request FRS.
// At the moment, that's only ga.Nom and rfc.RfC
type frsRequesting interface {
	// IncludeHeader returns a bool indicating if the header is applicable for the
	// requesting instance, and also a bool indicating if the header is the catch-all
	// for the requester.
	IncludeHeader(string) (headerShouldBeIncluded bool, headerIsAllHeader bool)

	PageTitle() string
	RequestType() string
}
