package main

import (
	"log"

	"github.com/ironzhang/coap"
)

type Handler struct{}

func (h Handler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	w.Write(r.Payload)
}

func main() {
	if err := coap.ListenAndServeDTLS("udp", ":5684", "./cert/svr.crt", "./cert/svr.key", Handler{}, nil); err != nil {
		log.Fatalf("listen and serve DTLS: %v", err)
	}
}
