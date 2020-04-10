// Copyright 2019-2020 go-gtp authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vishvananda/netlink"

	v1 "upf/v1"
	v2 "upf/v2"
	"upf/v2/messages"
)

type sgw struct {
	// C-Plane
	s11Addr, s5cAddr net.Addr
	s11Conn, s5cConn *v2.Conn

	// U-Plane
	s1uAddr, s5uAddr net.Addr
	s1uConn, s5uConn *v1.UPlaneConn

	s11IP, s5cIP, s1uIP, s5uIP string

	addedRoutes []*netlink.Route
	addedRules  []*netlink.Rule

	promAddr string
	mc       *metricsCollector

	errCh chan error
}

func newSGW(cfg *Config) (*sgw, error) {
	s := &sgw{
		errCh: make(chan error, 1),
	}

	var err error
	s.s11Addr, err = net.ResolveUDPAddr("udp", cfg.LocalAddrs.S11IP+v2.GTPCPort)
	if err != nil {
		return nil, err
	}
	s.s11IP, _, err = net.SplitHostPort(s.s11Addr.String())
	if err != nil {
		return nil, err
	}

	s.s5cAddr, err = net.ResolveUDPAddr("udp", cfg.LocalAddrs.S5CIP+v2.GTPCPort)
	if err != nil {
		return nil, err
	}
	s.s5cIP, _, err = net.SplitHostPort(s.s5cAddr.String())
	if err != nil {
		return nil, err
	}

	s.s1uAddr, err = net.ResolveUDPAddr("udp", cfg.LocalAddrs.S1UIP+v2.GTPUPort)
	if err != nil {
		return nil, err
	}
	s.s1uIP, _, err = net.SplitHostPort(s.s1uAddr.String())
	if err != nil {
		return nil, err
	}

	s.s5uAddr, err = net.ResolveUDPAddr("udp", cfg.LocalAddrs.S5UIP+v2.GTPUPort)
	if err != nil {
		return nil, err
	}
	s.s5uIP, _, err = net.SplitHostPort(s.s5uAddr.String())
	if err != nil {
		return nil, err
	}

	if cfg.PromAddr != "" {
		// validate if the address is valid or not.
		if _, err = net.ResolveTCPAddr("tcp", cfg.PromAddr); err != nil {
			return nil, err
		}
		s.promAddr = cfg.PromAddr
	}

	return s, nil
}

func (s *sgw) run(ctx context.Context) error {
	s.s11Conn = v2.NewConn(s.s11Addr, v2.IFTypeS11S4SGWGTPC, 0)
	go func() {
		if err := s.s11Conn.ListenAndServe(ctx); err != nil {
			log.Println(err)
			return
		}
	}()
	log.Printf("Started serving S11 on %s", s.s11Addr)

	s.s5cConn = v2.NewConn(s.s5cAddr, v2.IFTypeS5S8SGWGTPC, 0)
	go func() {
		if err := s.s5cConn.ListenAndServe(ctx); err != nil {
			log.Println(err)
			return
		}
	}()
	log.Printf("Started serving S5-C on %s", s.s5cAddr)

	// register handlers for ALL the messages you expect remote endpoint to send.
	s.s11Conn.AddHandlers(map[uint8]v2.HandlerFunc{
		messages.MsgTypeCreateSessionRequest: s.handleCreateSessionRequest,
		messages.MsgTypeModifyBearerRequest:  s.handleModifyBearerRequest,
		messages.MsgTypeDeleteSessionRequest: s.handleDeleteSessionRequest,
		messages.MsgTypeDeleteBearerResponse: s.handleDeleteBearerResponse,
	})
	s.s5cConn.AddHandlers(map[uint8]v2.HandlerFunc{
		messages.MsgTypeCreateSessionResponse: s.handleCreateSessionResponse,
		messages.MsgTypeDeleteSessionResponse: s.handleDeleteSessionResponse,
		messages.MsgTypeDeleteBearerRequest:   s.handleDeleteBearerRequest,
	})

	s.s1uConn = v1.NewUPlaneConn(s.s1uAddr)
	if err := s.s1uConn.EnableKernelGTP("gtp-sgw-s1", v1.RoleGGSN); err != nil {
		return err
	}
	go func() {
		if err := s.s1uConn.ListenAndServe(ctx); err != nil {
			log.Println(err)
			return
		}
		log.Println("uConn.ListenAndServe exitted")
	}()
	log.Printf("Started serving S1-U on %s", s.s1uAddr)

	s.s5uConn = v1.NewUPlaneConn(s.s5uAddr)
	if err := s.s5uConn.EnableKernelGTP("gtp-sgw-s5", v1.RoleSGSN); err != nil {
		return err
	}
	go func() {
		if err := s.s5uConn.ListenAndServe(ctx); err != nil {
			log.Println(err)
			return
		}
		log.Println("uConn.ListenAndServe exitted")
	}()
	log.Printf("Started serving S5-U on %s", s.s5uAddr)

	if err := s.addRoutes(); err != nil {
		return err
	}

	// start serving Prometheus, if address is given
	if s.promAddr != "" {
		if err := s.runMetricsCollector(); err != nil {
			return err
		}

		http.Handle("/metrics", promhttp.Handler())
		go func() {
			if err := http.ListenAndServe(s.promAddr, nil); err != nil {
				log.Println(err)
			}
		}()
		log.Printf("Started serving Prometheus on %s", s.promAddr)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-s.errCh:
			log.Printf("Warning: %s", errors.WithStack(err))
		}
	}
}

func (s *sgw) close() error {
	var errs []error
	for _, r := range s.addedRoutes {
		if err := netlink.RouteDel(r); err != nil {
			errs = append(errs, err)
		}
	}
	for _, r := range s.addedRules {
		if err := netlink.RuleDel(r); err != nil {
			errs = append(errs, err)
		}
	}

	if s.s1uConn != nil {
		if err := s.s1uConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.s5uConn != nil {
		if err := s.s5uConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.s11Conn != nil {
		if err := s.s11Conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.s5cConn != nil {
		if err := s.s5cConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	close(s.errCh)

	if len(errs) > 0 {
		return errors.Errorf("errors while closing S-GW: %v", errs)
	}
	return nil
}

func (s *sgw) addRoutes() error {
	defnet := &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}
	s1route := &netlink.Route{ // ip route replace
		Dst:       defnet,                  // default
		LinkIndex: s.s5uConn.GTPLink.Index, // dev gtp-s5
		Scope:     netlink.SCOPE_LINK,      // scope link
		Protocol:  4,                       // proto static
		Priority:  1,                       // metric 1
		Table:     2001,                    // table 2001
	}

	if err := netlink.RouteReplace(s1route); err != nil {
		return err
	}
	s.addedRoutes = append(s.addedRoutes, s1route)

	s5route := &netlink.Route{ // ip route replace
		Dst:       defnet,                          // default
		LinkIndex: s.s1uConn.GTPLink.Attrs().Index, // dev gtp-s1
		Scope:     netlink.SCOPE_LINK,              // scope link
		Protocol:  4,                               // proto static
		Priority:  1,                               // metric 1
		Table:     2005,                            // table 2005
	}

	if err := netlink.RouteReplace(s5route); err != nil {
		return err
	}
	s.addedRoutes = append(s.addedRoutes, s1route)

	rules, err := netlink.RuleList(0)
	if err != nil {
		return err
	}

	var s1found, s5found bool
	for _, r := range rules {
		if s1found && s5found {
			break
		}

		if r.IifName == s.s1uConn.GTPLink.Name && r.Table == 2001 {
			s1found = true
		}
		if r.IifName == s.s5uConn.GTPLink.Name && r.Table == 2005 {
			s5found = true
		}
	}

	if !s1found {
		rule := netlink.NewRule()
		rule.IifName = s.s1uConn.GTPLink.Name
		rule.Table = 2001

		if err := netlink.RuleAdd(rule); err != nil {
			return err
		}
		s.addedRules = append(s.addedRules, rule)
	}

	if !s5found {
		rule := netlink.NewRule()
		rule.IifName = s.s5uConn.GTPLink.Name
		rule.Table = 2005

		if err := netlink.RuleAdd(rule); err != nil {
			return err
		}
		s.addedRules = append(s.addedRules, rule)
	}

	return nil
}
