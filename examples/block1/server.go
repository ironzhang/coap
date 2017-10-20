package main

import (
	"io/ioutil"
	"log"

	"github.com/ironzhang/coap"
)

type Handler struct {
}

func (h Handler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	log.Printf("payload length: %d", len(r.Payload))
	if err := ioutil.WriteFile("output.html", r.Payload, 0664); err != nil {
		log.Printf("write file: %v", err)
	}
}

func main() {
	coap.ListenAndServe(":5683", Handler{}, nil)
}
