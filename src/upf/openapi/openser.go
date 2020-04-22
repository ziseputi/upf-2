package openapi

import (
	"net/http"
	"upf/src/upf/service"
)

func Start(node *service.Node) {
	http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("upf open api"))
	})
	http.HandleFunc("/link", linkshow)
	http.HandleFunc("/session/create", create)
	http.HandleFunc("/session/modify", modify)
	http.HandleFunc("/session/delete", delete)
	http.HandleFunc("/session/report", report)
	opNode = node
	http.ListenAndServe("0.0.0.0:8080", nil)
}
