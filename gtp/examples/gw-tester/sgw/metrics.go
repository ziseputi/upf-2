// Copyright 2019-2020 upf authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/vishvananda/netlink"
)

type metricsCollector struct {
	activeSessions   prometheus.GaugeFunc
	activeBearers    prometheus.GaugeFunc
	messagesSent     *prometheus.CounterVec
	messagesReceived *prometheus.CounterVec
}

func (s *sgw) runMetricsCollector() error {
	mc := &metricsCollector{}
	mc.activeSessions = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "sgw_active_sessions",
			Help: "number of session established currently",
		},
		func() float64 {
			return float64(s.s11Conn.SessionCount())
		},
	)

	mc.activeBearers = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "sgw_active_bearers",
			Help: "number of GTP-U tunnels established currently",
		},
		func() float64 {
			tunnels, err := netlink.GTPPDPList()
			if err != nil {
				log.Printf("metrics: could not get tunnels: %s", err)
				return 0
			}
			return float64(len(tunnels))
		},
	)

	mc.messagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sgw_messages_sent_total",
			Help: "number of messages sent by messagge type",
		},
		[]string{"dst", "type"},
	)

	mc.messagesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sgw_messages_received_total",
			Help: "number of messages received by messagge type",
		},
		[]string{"src", "type"},
	)

	s.mc = mc
	return nil
}
