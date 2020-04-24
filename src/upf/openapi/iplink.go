package openapi

import (
	"github.com/vishvananda/netlink"
	"log"
	"net/http"
)

func linkshow(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("upf print link"))
	func() float64 {
		log.Printf("link-print gtp")
		tunnels, err := netlink.GTPPDPList()
		if err != nil {
			log.Printf("openser: could not get tunnels: %s", err)
			return 0
		}
		log.Printf(" get gtp tunnels: %s", tunnels)
		return float64(len(tunnels))
	}()

}
