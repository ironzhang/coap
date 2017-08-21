package main

import (
	"log"

	"github.com/ironzhang/coap"
)

func main() {
	req, err := coap.NewRequest(true, coap.GET, "coap://localhost:5683/hello", []byte("hello, world"))
	if err != nil {
		log.Printf("new coap request: %v", err)
		return
	}
	req.Callback = func(resp *coap.Response) {
		log.Printf("%s\n", resp.Payload)
	}

	var c coap.Client
	if err = c.SendRequest(req); err != nil {
		log.Printf("send coap request: %v", err)
		return
	}
}
