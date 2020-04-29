package main

import (
	"flag"
	"log"
	"upf/src/eproxy/service"
)

func main() {
	var configPath = flag.String("config", "/home/wuhao/data/code/go/src/upf/src/eproxy/eproxy.yml", "Path to the configuration file.")
	flag.Parse()
	log.SetPrefix("[EPROXY] ")

	cfg, err := service.LoadConfig(*configPath)
	if err != nil {
		log.Println(err)
		return
	}
	service.SetUp(*cfg)
	service.ListenTcp(*cfg)
	//service.ListenTcp(cfg)

}
