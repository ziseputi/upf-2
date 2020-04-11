// Copyright 2019-2020 go-pfcp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ie

import (
	"encoding/binary"
	"io"
	"time"
)

// NewEthernetInactivityTimer creates a new EthernetInactivityTimer IE.
//
// the period should be within the range of uint32, otherwise it overflows.
func NewEthernetInactivityTimer(period time.Duration) *IE {
	return newUint32ValIE(EthernetInactivityTimer, uint32(period.Seconds()))
}

// EthernetInactivityTimer returns EthernetInactivityTimer in time.Duration if the type of IE matches.
func (i *IE) EthernetInactivityTimer() (time.Duration, error) {
	if i.Type != EthernetInactivityTimer {
		return 0, &InvalidTypeError{Type: i.Type}
	}
	if len(i.Payload) < 4 {
		return 0, io.ErrUnexpectedEOF
	}

	return time.Duration(binary.BigEndian.Uint32(i.Payload[0:4])) * time.Second, nil
}
