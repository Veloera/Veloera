// Copyright (c) 2025 Tethys Plex
//
// This file is part of Veloera.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package common

import (
	"time"
)

func SafeSendBool(ch chan bool, value bool) (closed bool) {
	defer func() {
		// Recover from panic if one occurred. A panic would mean the channel was closed.
		if recover() != nil {
			closed = true
		}
	}()

	// This will panic if the channel is closed.
	ch <- value

	// If the code reaches here, then the channel was not closed.
	return false
}

func SafeSendString(ch chan string, value string) (closed bool) {
	defer func() {
		// Recover from panic if one occurred. A panic would mean the channel was closed.
		if recover() != nil {
			closed = true
		}
	}()

	// This will panic if the channel is closed.
	ch <- value

	// If the code reaches here, then the channel was not closed.
	return false
}

// SafeSendStringTimeout send, return true, else return false
func SafeSendStringTimeout(ch chan string, value string, timeout int) (closed bool) {
	defer func() {
		// Recover from panic if one occurred. A panic would mean the channel was closed.
		if recover() != nil {
			closed = false
		}
	}()

	// This will panic if the channel is closed.
	select {
	case ch <- value:
		return true
	case <-time.After(time.Duration(timeout) * time.Second):
		return false
	}
}
