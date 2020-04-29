package main

import (
	"flag"
	"log"
	"upf/src/eproxy/service"
)

func main() {
	var configPath = flag.String("config", "/Users/wuhao/data/code/go/src/upf/src/eproxy/eproxy.yml", "Path to the configuration file.")
	flag.Parse()
	log.SetPrefix("[EPROXY] ")

	cfg, err := service.LoadConfig(*configPath)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Printf("load config complete")
	}
	service.SetUp(*cfg)
	service.ListenTcp(*cfg)
	//service.ListenTcp(cfg)

}
