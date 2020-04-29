package service

import (
	"context"
	"log"
	"net"
	"time"
	"upf/gtp/v1"
)

var CliConn *v1.UPlaneConn
var SrvAddr *net.UDPAddr

type testVal struct {
	teidIn, teidOut uint32
	seq             uint16
	payload         []byte
}

func SetUp() (cliConn *v1.UPlaneConn, add *net.UDPAddr, err error) {
	SrvAddr, err := net.ResolveUDPAddr("udp", "10.10.12.96:2152")
	if CliConn != nil {
		return CliConn, SrvAddr, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cliAddr, err := net.ResolveUDPAddr("udp", "10.10.12.96:2162")
	if err != nil {
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, err
	}

	// XXX - waiting for server to be well-prepared, should consider better way.
	time.Sleep(1 * time.Second)
	CliConn, err = v1.DialUPlane(ctx, cliAddr, SrvAddr)
	if err != nil {
		return nil, nil, err
	}

	return cliConn, SrvAddr, nil
}

func SendGtp(buffer []byte) {
	var (
		okCh  = make(chan struct{})
		errCh = make(chan error)
		//buf   = make([]byte, 2048)
		tv = &testVal{
			0x11111111, 0x22222222, 0x3333,
			buffer,
		}
	)

	cliConn, srvAddr, err := SetUp()
	if err != nil {
		log.Printf("set up error", err)
	}

	go func(tv *testVal) {
		//
		okCh <- struct{}{}
	}(tv)

	if _, err := cliConn.WriteToGTP(tv.teidOut, buffer, srvAddr); err != nil {
		log.Printf("send  get error", err)
	}

	select {
	case <-okCh:
		return
	case err := <-errCh:
		log.Printf("wait  error", err)
	case <-time.After(10 * time.Second):
		log.Printf("wait time out", err)
	}
}
