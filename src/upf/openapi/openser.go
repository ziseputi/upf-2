package openapi

import (
	"github.com/vishvananda/netlink"
	"log"
	"net/http"
)

func Start() {
	http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("upf open api"))
	})

	http.HandleFunc("/link", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("upf print link"))
		func() float64 {
			log.Printf("link-print gtp")
			tunnels, err := netlink.GTPPDPList()
			if err != nil {
				log.Printf("openser: could not get tunnels: %s", err)
				return 0
			}
			return float64(len(tunnels))
		}()
	})

	http.ListenAndServe("127.0.0.1:8080", nil)
}
