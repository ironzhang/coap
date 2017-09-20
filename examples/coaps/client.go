package main

import (
	"log"

	"github.com/ironzhang/coap"
	"github.com/ironzhang/dtls"
)

func main() {
	ca, err := dtls.LoadX509Certificate("./cert/ca.crt")
	if err != nil {
		log.Fatalf("load x509 certificate: %v", err)
	}
	defer ca.Close()

	var c = coap.Client{
		DTLSConfig: &dtls.Config{
			CA:       ca,
			Authmode: dtls.VERIFY_REQUIRED,
		},
	}

	req, err := coap.NewRequest(true, coap.POST, "coaps://localhost", []byte("hello, world"))
	if err != nil {
		log.Fatalf("new request: %v", err)
	}
	resp, err := c.SendRequest(req)
	if err != nil {
		log.Fatalf("send request: %v", err)
	}
	log.Printf("%s", resp.Payload)
}
