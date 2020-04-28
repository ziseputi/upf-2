package main

import (
	"fmt"
	"github.com/mushroomsir/blog/examples/util"
	"golang.org/x/net/ipv4"
	"os"
	"syscall"
	"upf/gtp/utils"
)

func main() {
	fd, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	f := os.NewFile(uintptr(fd), fmt.Sprintf("fd %d", fd))
	for {
		buf := make([]byte, 1500)
		f.Read(buf)
		ip4header, _ := ipv4.ParseHeader(buf[:20])
		npi := ip4header.Src.To4()
		cpi := utils.ToIp("10.10.12.77")
		if npi.String() == cpi.String() {
			fmt.Println("ipheader:", ip4header)
			tcpheader := util.NewTCPHeader(buf[20:40])
			fmt.Println("tcpheader:", tcpheader)

		}

		//tcpheader := util.NewTCPHeader(buf[20:40])
		//fmt.Println("tcpheader:", tcpheader)
	}
}
