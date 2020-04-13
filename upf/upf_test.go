// Copyright 2019-2020 upf authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main_test

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"net"
	"testing"
	"time"

	"upf/gtp/v1"
)

type testVal struct {
	teidIn, teidOut uint32
	seq             uint16
	payload         []byte
}

func setup(ctx context.Context) (cliConn *v1.UPlaneConn, add *net.UDPAddr, err error) {
	cliAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2162")
	if err != nil {
		return nil, nil, err
	}
	srvAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2152")
	if err != nil {
		return nil, nil, err
	}

	// XXX - waiting for server to be well-prepared, should consider better way.
	time.Sleep(1 * time.Second)
	cliConn, err = v1.DialUPlane(ctx, cliAddr, srvAddr)
	if err != nil {
		return nil, nil, err
	}

	return cliConn, srvAddr, nil
}

func TestClientWrite(t *testing.T) {
	var (
		okCh  = make(chan struct{})
		errCh = make(chan error)
		buf   = make([]byte, 2048)
		tv    = &testVal{
			0x11111111, 0x22222222, 0x3333,
			[]byte{0xde, 0xad, 0xbe, 0xef},
		}
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cliConn, srvAddr, err := setup(ctx)
	if err != nil {
		t.Fatal(err)
	}

	go func(tv *testVal) {
		n, addr, teid, err := cliConn.ReadFromGTP(buf)
		if err != nil {
			errCh <- err
			return
		}

		if diff := cmp.Diff(n, len(tv.payload)); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(addr, cliConn.LocalAddr()); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(teid, tv.teidOut); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(buf[:n], tv.payload); diff != "" {
			t.Error(diff)
		}
		okCh <- struct{}{}
	}(tv)

	if _, err := cliConn.WriteToGTP(tv.teidOut, tv.payload, srvAddr); err != nil {
		t.Fatal(err)
	}

	select {
	case <-okCh:
		return
	case err := <-errCh:
		t.Fatal(err)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out while waiting for response to come")
	}
}
