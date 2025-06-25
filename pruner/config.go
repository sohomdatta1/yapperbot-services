package main

//
// Yapperbot-Pruner, the user pruning bot for Wikipedia
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

// Config holds the configuration pulled from the standard
// ybtools task-specific config file.
type Config struct {
	DSN                       string
	ConfigTemplate            string
	FormatsJSONPageID         string
	DefaultExpiredMsgTemplate string
	DefaultTalkMsgHeader      string
}

var config Config
