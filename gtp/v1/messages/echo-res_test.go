// Copyright 2019-2020 upf authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package messages_test

import (
	"testing"

	"upf/gtp/v1/ies"
	"upf/gtp/v1/messages"
	"upf/gtp/v1/testutils"
)

func TestEchoResponse(t *testing.T) {
	cases := []testutils.TestCase{
		{
			Description: "Normal",
			Structured:  messages.NewEchoResponse(0, ies.NewRecovery(0x80)),
			Serialized: []byte{
				0x32, 0x02, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x0e, 0x80,
			},
		},
	}

	testutils.Run(t, cases, func(b []byte) (testutils.Serializable, error) {
		v, err := messages.ParseEchoResponse(b)
		if err != nil {
			return nil, err
		}
		v.Payload = nil
		return v, nil
	})
}
