package main

import (
	"golang.org/x/net/context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"upf/gtp/v1"
)

func main() {
	log.Println("upf starting")
	var address = "10.10.12.77:2152"

	start(address)

	log.Println("upf started")

	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGUSR1, syscall.SIGUSR2)
	s := <-c

	log.Println("upf stoped", s)
}

func start(address string) (srvConn *v1.UPlaneConn, err error) {
	log.Println("upf gtp bind ip:", address)
	ctx, cancel := context.WithCancel(context.Background())
	if cancel != nil {
		return nil, err
	}
	srvAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Println("ResolveUDPAddr ip error:", address)
		return nil, err
	}
	srvConn = v1.NewUPlaneConn(srvAddr)
	if err := srvConn.ListenAndServe(ctx); err != nil {
		log.Println("upf start bind ip error:", address)
		return nil, err
	}
	return srvConn, nil
}
