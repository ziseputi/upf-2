// Copyright 2019-2020 upf authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package service

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config is a configurations loaded from yaml.
type Config struct {
	LocalAddrs struct {
		GTPUADDR string `yaml:"gtpu_addr"`
		PFCPADDR string `yaml:"pfcp_addr"`
	} `yaml:"local_addresses"`

	MCC string `yaml:"mcc"`
	MNC string `yaml:"mnc"`
	APN string `yaml:"apn"`

	PeerAddrs struct {
		SMFADDR string `yaml:"smf_addr"`
	} `yaml:"peer_addresses"`
	RouteSubnet string `yaml:"route_subnet"`
}

func LoadConfig(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	if err := yaml.Unmarshal(buf, c); err != nil {
		return nil, err
	}

	return c, nil
}
