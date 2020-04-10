// Copyright 2019-2020 go-gtp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package messages_test

import (
	"testing"

	"upf/v2/ies"
	"upf/v2/messages"
	"upf/v2/testutils"
)

func TestEchoRequest(t *testing.T) {
	cases := []testutils.TestCase{
		{
			Description: "Normal",
			Structured:  messages.NewEchoRequest(0, ies.NewRecovery(0x80)),
			Serialized: []byte{
				0x40, 0x01, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00,
				0x03, 0x00, 0x01, 0x00, 0x80,
			},
		},
	}

	testutils.Run(t, cases, func(b []byte) (testutils.Serializable, error) {
		v, err := messages.ParseEchoRequest(b)
		if err != nil {
			return nil, err
		}
		v.Payload = nil
		return v, nil
	})
}
