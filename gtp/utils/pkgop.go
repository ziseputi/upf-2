package utils

import (
	"github.com/mushroomsir/blog/examples/util"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"strconv"
	"strings"
	"syscall"
)

//up data
func WriteMock() {
	local := "10.10.12.96"
	remote := "10.10.12.77"
	dport := uint16(8099)
	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: To4byte(local),
	}
	yipHeader := ipv4.Header{
		Version:  4,
		Len:      20,
		TotalLen: 20, // 20 bytes for IP, 10 for ICMP
		TTL:      64,
		Flags:    0x4000,
		Protocol: 6, // TCP
		Dst:      ToIp(remote),
		Src:      ToIp(local),
	}
	payload, _ := yipHeader.Marshal()
	ycpHeader := TCPHeader{
		Source:      17663, // Random ephemeral port
		Destination: dport,
		Reserved:    0,      // 3 bits
		ECN:         0,      // 3 bits
		Ctrl:        2,      // 6 bits (000010, SYN bit set)
		Window:      0xaaaa, // size of your receive window
		Checksum:    0,      // Kernel will set this if it's 0
		Urgent:      99,
	}
	data := ycpHeader.Marshal()
	ycpHeader.Checksum = util.Csum(data, To4byte(local), To4byte(remote))
	data = ycpHeader.Marshal()
	payload = append(payload, data...)
	//write data
	WritePayLoad(payload, addr)
}

func IpWritePayLoad(payload []byte, ip string) {
	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: To4byte(ip),
	}
	WritePayLoad(payload, addr)
}
func WritePayLoad(payload []byte, addr syscall.SockaddrInet4) {
	fd, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	syscall.Sendto(fd, payload, 0, &addr)
}

func To4byte(addr string) [4]byte {
	parts := strings.Split(addr, ".")
	b0, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Fatalf("to4byte: %s (latency works with IPv4 addresses only, but not IPv6!)\n", err)
	}
	b1, _ := strconv.Atoi(parts[1])
	b2, _ := strconv.Atoi(parts[2])
	b3, _ := strconv.Atoi(parts[3])
	return [4]byte{byte(b0), byte(b1), byte(b2), byte(b3)}
}
func ToIp(addr string) net.IP {
	b4addr := To4byte(addr)
	dip := net.IPv4(byte(b4addr[0]), byte(b4addr[1]), byte(b4addr[2]), byte(b4addr[3]))
	return dip
}
