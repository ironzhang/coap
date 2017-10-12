package main

import (
	"log"

	"github.com/ironzhang/coap"
)

type Handler struct {
}

func (h Handler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	log.Println(r.URL.String())
	w.Write(r.Payload)
}

func main() {
	coap.Verbose = 2
	if err := coap.ListenAndServe("udp", ":5683", Handler{}, nil); err != nil {
		log.Fatal(err)
	}
}
