package service

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Config is a configurations loaded from yaml.
type Config struct {
	FilterAddr string `yaml:"filter_addr"`
	UpfAddr    string `yaml:"upf_addr"`
	ProxyAddr  string `yaml:"proxy_addr"`
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
