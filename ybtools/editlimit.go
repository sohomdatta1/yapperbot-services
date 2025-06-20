package ybtools

import (
	"encoding/binary"
	"io/ioutil"
	"log"
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

var currentUsedEditLimit int64
var editLimit int64

// EditLimit can be called to increment the current edit count
// Returns true if allowed to edit or false if not
// Remember to call SaveEditLimit at the end of the program if using this!
func EditLimit() bool {
	if editLimit > 0 {
		if currentUsedEditLimit >= editLimit {
			log.Println("edit limited, not performing edit - limit was", editLimit, "and this is", currentUsedEditLimit)
			return false
		}

		currentUsedEditLimit++
		return true
	}
	return true
}

// SaveEditLimit saves the current edit limit to the edit limit file,
// assuming that there is an edit limit usage to save
// This function must be called at the end of the program for edit limiting to work
func SaveEditLimit() {
	if currentUsedEditLimit > 0 {
		buf := make([]byte, binary.MaxVarintLen16)
		binary.PutVarint(buf, currentUsedEditLimit)
		err := ioutil.WriteFile("editlimit", buf, 0644)
		if err != nil {
			PanicErr("Failed to write edit limit file with err ", err)
		}
	}
}

// SetupEditLimit takes in a limit as an int64
// and sets that as the limit for edits for the bot,
// as well as enabling the edit limiting functionality
func setupEditLimit(limit int64) {
	editLimit = limit

	editLimitFileContents, err := ioutil.ReadFile("editlimit")
	if err != nil {
		// the edit limit file doesn't exist probably, try creating it
		err := ioutil.WriteFile("editlimit", []uint8{0x00, 0x00, 0x00}, 0644)
		if err != nil {
			PanicErr("Failed to create edit limit file with error ", err)
		}
		editLimitFileContents = []uint8{0x00, 0x00, 0x00}
	}
	var bytesRead int
	currentUsedEditLimit, bytesRead = binary.Varint(editLimitFileContents)
	if bytesRead < 0 {
		PanicErr("editlimit file is corrupt, failed to convert with bytesRead ", bytesRead)
	}
}
