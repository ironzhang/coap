package main

import (
	"log"

	"github.com/ironzhang/coap"
)

type handler struct {
}

func (h *handler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	log.Printf("%s\n", r.Payload)
	w.Write(r.Payload)
}

func main() {
	if err := coap.ListenAndServe("udp", ":5683", &handler{}); err != nil {
		log.Fatal(err)
	}
}
