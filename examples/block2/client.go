package main

import (
	"io/ioutil"
	"log"

	"github.com/ironzhang/coap"
)

func main() {
	var c coap.Client
	req, err := coap.NewRequest(true, coap.GET, "coap://localhost:5683", nil)
	if err != nil {
		log.Printf("new request: %v", err)
		return
	}
	resp, err := c.SendRequest(req)
	if err != nil {
		log.Printf("send request: %v", err)
		return
	}
	if err := ioutil.WriteFile("output.html", resp.Payload, 0664); err != nil {
		log.Printf("write file: %v", err)
		return
	}
}
