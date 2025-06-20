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

// NoRfCIDYetError is an error used when an RfC detected does not yet have an ID assigned.
type NoRfCIDYetError struct{}

const noRfCIDYetErrorText string = "An identified RfC does not yet have an assigned RfC ID from Legobot."

func (e NoRfCIDYetError) Error() string {
	return noRfCIDYetErrorText
}
