package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"upf/src/upf/openapi"
	"upf/src/upf/service"
)

func main() {
	var configPath = flag.String("config", "/home/wuhao/data/code/go/src/upf/src/upf/upf.yml", "Path to the configuration file.")
	flag.Parse()
	log.SetPrefix("[UPF] ")

	cfg, err := service.LoadConfig(*configPath)
	if err != nil {
		log.Println(err)
		return
	}

	node, err := service.NewUPF(cfg)
	if err != nil {
		log.Printf("failed to initialize UPF: %s", err)
		return
	}
	defer node.Close()

	go openapi.Start(node)

	nigCh := make(chan os.Signal, 1)
	signal.Notify(nigCh, syscall.SIGINT, syscall.SIGHUP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fatalCh := make(chan error)
	go func() {
		if err := node.Run(ctx); err != nil {
			fatalCh <- err
		}
	}()

	for {
		select {
		case sig := <-nigCh:
			// TODO: reload config on receiving NIGHUP
			log.Println(sig)
			return
		case err := <-node.ErrCh:
			log.Printf("WARN: %s", err)
		case err := <-fatalCh:
			log.Printf("FATAL: %s", err)
			return
		}
	}
}
