// Copyright 2019-2020 upf authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package messages_test

import (
	"testing"

	v2 "upf/gtp/v2"
	"upf/gtp/v2/ies"
	"upf/gtp/v2/messages"
	"upf/gtp/v2/testutils"
)

func TestReleaseAccessBearersResponse(t *testing.T) {
	cases := []testutils.TestCase{
		{
			Description: "Normal/CauseOnly",
			Structured: messages.NewReleaseAccessBearersResponse(
				testutils.TestBearerInfo.TEID, testutils.TestBearerInfo.Seq,
				ies.NewCause(v2.CauseRequestAccepted, 0, 0, 0, nil),
			),
			Serialized: []byte{
				// Header
				0x48, 0xab, 0x00, 0x0e, 0x11, 0x22, 0x33, 0x44, 0x00, 0x00, 0x01, 0x00,
				0x02, 0x00, 0x02, 0x00, 0x10, 0x00,
			},
		},
	}

	testutils.Run(t, cases, func(b []byte) (testutils.Serializable, error) {
		v, err := messages.ParseReleaseAccessBearersResponse(b)
		if err != nil {
			return nil, err
		}
		v.Payload = nil
		return v, nil
	})
}
