package main

import (
	"io/ioutil"
	"log"

	"github.com/ironzhang/coap"
)

func main() {
	var c coap.Client

	payload, err := ioutil.ReadFile("./ietf-block.html")
	if err != nil {
		log.Printf("read file: %v", err)
		return
	}

	req, err := coap.NewRequest(true, coap.POST, "coap://localhost:5683", payload)
	if err != nil {
		log.Printf("new request: %v", err)
		return
	}

	_, err = c.SendRequest(req)
	if err != nil {
		log.Printf("send request: %v", err)
		return
	}
}
