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
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

// PanicErr panics the program with a specified message,
// also sending a message to the tool inbox on Toolforge explaining the issue.
func PanicErr(v ...interface{}) {
	strerr := fmt.Sprint(v...)
	toolemail := "tools." + strings.ToLower(settings.BotUser) + "@tools.wmflabs.org"

	m := gomail.NewMessage()
	m.SetHeader("From", toolemail)
	m.SetHeader("To", toolemail)
	m.SetHeader("Subject", settings.BotUser+" errored in "+settings.TaskName)
	m.SetBody("text/plain", strerr)

	d := gomail.Dialer{Host: "mail.tools.wmflabs.org", Port: 25}
	if err := d.DialAndSend(m); err != nil {
		strerr = "FAILED TO EMAIL ERROR (ERR " + err.Error() + "): " + strerr
	}
	panic(strerr)
}
