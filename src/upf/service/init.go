package service

import (
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/context"
	"log"
	"net"
	"upf/gtp/v1"
	"upf/gtp/v2"
	"upf/gtp/v2/messages"
)

type Node struct {
	cConn *v2.Conn
	uConn *v1.UPlaneConn

	n3c, n3u string
	ngiIF    string

	routeSubnet *net.IPNet
	addedRoutes []*netlink.Route
	addedRules  []*netlink.Rule

	promAddr string

	ErrCh chan error
}

func NewUPF(cfg *Config) (*Node, error) {
	node := &Node{
		n3c:   cfg.LocalAddrs.N3CADDR,
		n3u:   cfg.LocalAddrs.N3UADDR,
		ngiIF: cfg.NgIfName,

		ErrCh: make(chan error, 1),
	}

	var err error
	_, node.routeSubnet, err = net.ParseCIDR(cfg.RouteSubnet)
	if err != nil {
		return nil, err
	}

	if cfg.PromAddr != "" {
		// validate if the address is valid or not.
		if _, err = net.ResolveTCPAddr("tcp", cfg.PromAddr); err != nil {
			return nil, err
		}
		node.promAddr = cfg.PromAddr
	}

	return node, nil
}

func (node *Node) Run(ctx context.Context) error {
	cAddr, err := net.ResolveUDPAddr("udp", node.n3c)
	if err != nil {
		return err
	}
	node.cConn = v2.NewConn(cAddr, v2.IFTypeS5S8PGWGTPC, 0)
	go func() {
		if err := node.cConn.ListenAndServe(ctx); err != nil {
			log.Println(err)
			return
		}
	}()
	log.Printf("Started serving GTP-C on %s", cAddr)

	// register handlers for ALL the messages you expect remote endpoint to send.
	node.cConn.AddHandlers(map[uint8]v2.HandlerFunc{
		messages.MsgTypeCreateSessionRequest: node.handleCreateSessionRequest,
		messages.MsgTypeDeleteSessionRequest: node.handleDeleteSessionRequest,
	})

	uAddr, err := net.ResolveUDPAddr("udp", node.n3u)
	if err != nil {
		return err
	}
	node.uConn = v1.NewUPlaneConn(uAddr)
	if err := node.uConn.EnableKernelGTP("gtp-upf", v1.RoleGGSN); err != nil {
		return err
	}
	go func() {
		if err = node.uConn.ListenAndServe(ctx); err != nil {
			log.Println(err)
			return
		}
		log.Println("uConn.ListenAndServe exitted")
	}()
	log.Printf("Started serving GTP-U on %s", uAddr)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-node.ErrCh:
			log.Printf("Warning: %s", err)
		}
	}
}

func (node *Node) Close() error {
	var errs []error
	for _, r := range node.addedRoutes {
		if err := netlink.RouteDel(r); err != nil {
			errs = append(errs, err)
		}
	}
	for _, r := range node.addedRules {
		if err := netlink.RuleDel(r); err != nil {
			errs = append(errs, err)
		}
	}

	if node.uConn != nil {
		if err := netlink.LinkDel(node.uConn.GTPLink); err != nil {
			errs = append(errs, err)
		}
		if err := node.uConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if err := node.cConn.Close(); err != nil {
		errs = append(errs, err)
	}

	close(node.ErrCh)

	if len(errs) > 0 {
		return errors.Errorf("errors while closing S-GW: %v", errs)
	}
	return nil
}

func (node *Node) setupUPlane(peerIP, msIP net.IP, otei, itei uint32) error {
	if err := node.uConn.AddTunnelOverride(peerIP, msIP, otei, itei); err != nil {
		return err
	}

	ms32 := &net.IPNet{IP: msIP, Mask: net.CIDRMask(32, 32)}
	dlroute := &netlink.Route{ // ip route replace
		Dst:       ms32,                             // UE's IP
		LinkIndex: node.uConn.GTPLink.Attrs().Index, // dev gtp-pgw
		Scope:     netlink.SCOPE_LINK,               // scope link
		Protocol:  4,                                // proto static
		Priority:  1,                                // metric 1
		Table:     3001,                             // table 3001
	}
	if err := netlink.RouteReplace(dlroute); err != nil {
		return err
	}
	node.addedRoutes = append(node.addedRoutes, dlroute)

	link, err := netlink.LinkByName(node.ngiIF)
	if err != nil {
		return err
	}

	ulroute := &netlink.Route{ // ip route replace
		Dst:       node.routeSubnet,   // dst network via SGi
		LinkIndex: link.Attrs().Index, // SGi I/F name
		Scope:     netlink.SCOPE_LINK, // scope link
		Protocol:  4,                  // proto static
		Priority:  1,                  // metric 1
	}
	if err := netlink.RouteReplace(ulroute); err != nil {
		return err
	}
	node.addedRoutes = append(node.addedRoutes, ulroute)

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
	node.addedRules = append(node.addedRules, rule)

	return nil
}
