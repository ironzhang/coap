package main

import (
	"log"

	"github.com/ironzhang/coap"
)

func responseHandler(c *coap.Client, resp *coap.Response) {
	log.Printf("%s\n", resp.Payload)
}

func main() {
	req, err := coap.NewRequest(true, coap.GET, "coap://localhost:5683/hello", []byte("hello, world"))
	if err != nil {
		log.Printf("new coap request: %v", err)
		return
	}

	var c coap.Client
	if err = c.SendRequest(req, responseHandler); err != nil {
		log.Printf("send coap request: %v", err)
		return
	}
}
