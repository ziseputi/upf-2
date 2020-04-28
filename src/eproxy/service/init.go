package service

import (
	"fmt"
	"golang.org/x/net/ipv4"
	"os"
	"syscall"
	"upf/gtp/utils"
)

type Node struct {
	routeAddr string

	ErrCh chan error
}

func ListenTcp(config Config) {
	fd, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd %d", fd))
	for {
		buf := make([]byte, 1500)
		f.Read(buf)
		ip4header, _ := ipv4.ParseHeader(buf[:20])
		npi := ip4header.Src.To4()
		cpi := utils.ToIp(config.RouteAddrs.Addr)
		if npi.String() == cpi.String() {
			fmt.Println("ipheader:", ip4header)

		}

	}
}

func ListenUdp(config Config) {
	fd, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd %d", fd))
	for {
		buf := make([]byte, 1500)
		f.Read(buf)
		ip4header, _ := ipv4.ParseHeader(buf[:20])
		npi := ip4header.Src.To4()
		cpi := utils.ToIp(config.RouteAddrs.Addr)
		if npi.String() == cpi.String() {
			fmt.Println("ipheader:", ip4header)
			SendGtp(buf)

		}
		//tcpheader := util.NewTCPHeader(buf[20:40])
		//fmt.Println("tcpheader:", tcpheader)
	}
}
