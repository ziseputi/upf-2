package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
	"upf/src/upf/service"
)

func TestClientWrite(t *testing.T) {
	message := &service.SessionMessage{
		Teid:   111,
		PeerIp: "10.10.12.96",
		UeIp:   "10.10.12.77",
		Imsi:   "111111",
	}

	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	jsonStr, _ := json.Marshal(message)
	resp, err := client.Post("http://127.0.0.1:8099/session/create", "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 超时时间：5秒
	clientGet := &http.Client{Timeout: 5 * time.Second}
	clientGet.Get("http://127.0.0.1:8099/link")

}
