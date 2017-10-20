package main

import (
	"fmt"
	"log"

	"github.com/ironzhang/coap"
)

func main() {
	coap.Verbose = 0
	for i := 0; i < 10; i++ {
		var client coap.Client
		req, err := coap.NewRequest(true, coap.POST, "coap://localhost/ping", []byte("ping"))
		if err != nil {
			log.Printf("new request: %v", err)
			return
		}
		req.Options.Add(coap.URIPort, 5683)
		req.Options.Add(coap.URIPort, 5684)

		resp, err := client.SendRequest(req)
		if err != nil {
			log.Printf("send coap request: %v", err)
			return
		}
		fmt.Printf("%s\n", resp.Payload)
	}
}
