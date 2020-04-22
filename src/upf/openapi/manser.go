package openapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"upf/src/upf/service"
)

var opNode *service.Node

func create(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read body err, %v\n", err)
		return
	}
	println("json:", string(body))

	var message service.SessionMessage
	if err = json.Unmarshal(body, &message); err != nil {
		fmt.Printf("Unmarshal err, %v\n", err)
		return
	}
	fmt.Printf("%+v", message)
	w.Write([]byte("create ok"))
	opNode.CreateSessionRequest(message)

}

func modify(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read body err, %v\n", err)
		return
	}
	println("json:", string(body))

	var message service.SessionMessage
	if err = json.Unmarshal(body, &message); err != nil {
		fmt.Printf("Unmarshal err, %v\n", err)
		return
	}
	fmt.Printf("%+v", message)

}

func delete(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read body err, %v\n", err)
		return
	}
	println("json:", string(body))

	var message service.SessionMessage
	if err = json.Unmarshal(body, &message); err != nil {
		fmt.Printf("Unmarshal err, %v\n", err)
		return
	}
	fmt.Printf("%+v", message)

	w.Write([]byte("delete ok"))
	opNode.CreateSessionRequest(message)
}

func report(w http.ResponseWriter, r *http.Request) {

	fmt.Println("method:", r.Method)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read body err, %v\n", err)
		return
	}
	println("json:", string(body))

	var message service.SessionMessage
	if err = json.Unmarshal(body, &message); err != nil {
		fmt.Printf("Unmarshal err, %v\n", err)
		return
	}
	fmt.Printf("%+v", message)

}
