package service

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config is a configurations loaded from yaml.
type Config struct {
	LocalAddrs struct {
		N3CADDR string `yaml:"n3c_addr"`
		N3UADDR string `yaml:"n3u_addr"`
		N4ADDR  string `yaml:"n4_addr"`
	} `yaml:"local_addresses"`

	MCC       string `yaml:"mcc"`
	MNC       string `yaml:"mnc"`
	APN       string `yaml:"apn"`
	NgIfName  string `yaml:"ngif_name"`
	PromAddr  string `yaml:"prom_addr"`
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
