package main

import (
	"context"
	"github.com/vishvananda/netlink"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"upf/gtp/v1"
	"upf/gtp/v2"
	"upf/upf/service"
)

type upf struct {
	cConn *v2.Conn
	uConn *v1.UPlaneConn

	s5c, s5u string
	sgiIF    string

	routeSubnet *net.IPNet
	addedRoutes []*netlink.Route
	addedRules  []*netlink.Rule

	errCh chan error
}

func main() {

	initService("/home/wuhao/data/code/go/src/upf/upf/upf.yml")

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
	ui := &upf{
		errCh: make(chan error, 1),
	}

	//if err := ui.setupUPlane(net.ParseIP("10.10.12.96"), net.ParseIP("10.10.12.96"), 11111,1111); err != nil {
	//	return
	//}
	_, ui.routeSubnet, err = net.ParseCIDR(cfg.RouteSubnet)

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
	log.Println("upf start run server:", address)
	srvConn := v1.NewUPlaneConn(srvAddr)
	if err := srvConn.EnableKernelGTP("gtp-upf", v1.RoleSGSN); err != nil {
		log.Println("gtp-upf EnableKernelGTP error:", address)

	}
	func() {

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

func (ui *upf) setupUPlane(peerIP, msIP net.IP, otei, itei uint32) error {
	if err := ui.uConn.AddTunnelOverride(peerIP, msIP, otei, itei); err != nil {
		return err
	}

	ms32 := &net.IPNet{IP: msIP, Mask: net.CIDRMask(32, 32)}
	dlroute := &netlink.Route{ // ip route replace
		Dst:       ms32,                           // UE's IP
		LinkIndex: ui.uConn.GTPLink.Attrs().Index, // dev gtp-pgw
		Scope:     netlink.SCOPE_LINK,             // scope link
		Protocol:  4,                              // proto static
		Priority:  1,                              // metric 1
		Table:     3001,                           // table 3001
	}
	if err := netlink.RouteReplace(dlroute); err != nil {
		return err
	}
	ui.addedRoutes = append(ui.addedRoutes, dlroute)

	link, err := netlink.LinkByName(ui.sgiIF)
	if err != nil {
		return err
	}

	ulroute := &netlink.Route{ // ip route replace
		Dst:       ui.routeSubnet,     // dst network via SGi
		LinkIndex: link.Attrs().Index, // SGi I/F name
		Scope:     netlink.SCOPE_LINK, // scope link
		Protocol:  4,                  // proto static
		Priority:  1,                  // metric 1
	}
	if err := netlink.RouteReplace(ulroute); err != nil {
		return err
	}
	ui.addedRoutes = append(ui.addedRoutes, ulroute)

	rules, err := netlink.RuleList(0)
	if err != nil {
		return err
	}
	for _, r := range rules {
		if r.IifName == link.Attrs().Name && r.Dst == ms32 {
			return nil
		}
	}

	rule := netlink.NewRule()
	rule.IifName = link.Attrs().Name
	rule.Dst = ms32
	rule.Table = 3001
	if err := netlink.RuleAdd(rule); err != nil {
		return err
	}
	ui.addedRules = append(ui.addedRules, rule)

	return nil
}
