package service

import (
	"golang.org/x/net/context"
	"log"
	"net"
	"time"
	"upf/gtp/v1"
)

var CliConn *v1.UPlaneConn
var SrvAddr *net.UDPAddr

func SetUp() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cliAddr, err := net.ResolveUDPAddr("udp", "10.10.12.96:2562")
	if err != nil {
		log.Printf("net sdp gtp error")
		return
	}
	SrvAddr, err := net.ResolveUDPAddr("udp", "10.10.12.96:2152")
	if err != nil {
		log.Printf("net sdp gtp error")
		return
	}

	// XXX - waiting for server to be well-prepared, should consider better way.
	time.Sleep(1 * time.Second)
	CliConn, err = v1.DialUPlane(ctx, cliAddr, SrvAddr)
	if err != nil {
		return
	}

}
func SendGtp(buffer []byte) {

	if _, err := CliConn.WriteToGTP(111, buffer, SrvAddr); err != nil {
		log.Printf("send gtp error")
	}
}
