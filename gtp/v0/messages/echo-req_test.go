// Copyright 2019-2020 upf authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package messages_test

import (
	"testing"

	"upf/gtp/v0/messages"
	"upf/gtp/v0/testutils"
)

func TestEchoRequest(t *testing.T) {
	cases := []testutils.TestCase{
		{
			Description: "normal",
			Structured: messages.NewEchoRequest(
				testutils.TestFlow.Seq, testutils.TestFlow.Label, testutils.TestFlow.TID,
			),
			Serialized: []byte{
				// Header
				0x1e, 0x01, 0x00, 0x00,
				// SequenceNumber
				0x00, 0x01, 0x00, 0x00,
				// Sndpd
				0xff, 0xff, 0xff, 0xff,
				// TID
				0x21, 0x43, 0x65, 0x87, 0x09, 0x21, 0x43, 0x55,
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
