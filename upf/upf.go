package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"upf/gtp/v1"
	"upf/upf/service"
)

func main() {

	initService("/Users/wuhao/data/code/github/go/src/upf/upf/upf.yml")

	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGUSR1, syscall.SIGUSR2)
	s := <-c

	log.Println("upf stoped", s)
}

func initService(file string) {
	var pfcpAddress = "127.0.0.1:2152"
	var gtpAddress = "127.0.0.1:2152"
	cfg, err := service.LoadConfig(file)
	if err != nil {
		log.Fatal(err)
	} else {
		pfcpAddress = cfg.LocalAddrs.PFCPADDR
		gtpAddress = cfg.LocalAddrs.GTPUADDR
	}
	go initPFCP(pfcpAddress)
	go initGTPU(gtpAddress)
}

func initPFCP(address string) {

}

func initGTPU(address string) (err error) {
	log.Println("upf starting")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if ctx == nil {
		log.Println("context.WithCancel ip error:", address)
		return err
	}
	srvAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Println("ResolveUDPAddr ip error:", address)
		return err
	}
	//start server

	func() {
		log.Println("upf start run server:", address)
		srvConn := v1.NewUPlaneConn(srvAddr)
		if err := srvConn.ListenAndServe(ctx); err != nil {
			log.Println("upf start ListenAndServe error:", address)
			return
		} else {
			log.Println("upf bind ip ok:", address)
		}

	}()

	log.Println("upf start ok")

	return nil
}
