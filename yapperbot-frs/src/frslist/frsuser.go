package frslist

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

// An FRSUser is a struct representing a user who has signed up for the FRS.
// A single username may have multiple FRSUser objects; each corresponds to an
// individual subscription.
type FRSUser struct {
	Username string
	Header   string
	Limit    uint16
	Limited  bool
}

// GetCount takes a header and gets the number of messages sent for that header this month.
func (f FRSUser) GetCount() uint16 {
	sentCountMux.Lock()
	defer sentCountMux.Unlock()
	return sentCount[f.Header][f.Username]
}

// ExceedsLimit is a simple helper function for checking if a user is limited,
// and if they are, whether they can be messaged according to their limits.
func (f FRSUser) ExceedsLimit() bool {
	if f.Limited {
		return (f.GetCount() >= f.Limit)
	}
	return false
}

// MarkMessageSent increases the number of messages sent for the user by one. It's
// intended for use at the point of queueing a message.
func (f FRSUser) MarkMessageSent() {
	sentCountMux.Lock()
	defer sentCountMux.Unlock()

	// prevent nil map errors
	if sentCount[f.Header] == nil {
		sentCount[f.Header] = map[string]uint16{}
	}

	sentCount[f.Header][f.Username]++
}

// MarkMessageUnsent decreases the number of messages sent for the user by one. It
// should only be used if something goes wrong while we're sending a message to the user.
func (f FRSUser) MarkMessageUnsent() {
	sentCountMux.Lock()
	defer sentCountMux.Unlock()

	// prevent nil map errors
	if sentCount[f.Header] == nil {
		return
	}

	sentCount[f.Header][f.Username]--
}
