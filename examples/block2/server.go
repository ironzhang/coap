package main

import (
	"io/ioutil"
	"log"

	"github.com/ironzhang/coap"
)

type Handler struct {
}

func (h Handler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	payload, err := ioutil.ReadFile("./ietf-block.html")
	if err != nil {
		log.Printf("read file: %v", err)
		return
	}
	w.Write(payload)
}

func main() {
	coap.ListenAndServe(":5683", Handler{}, nil)
}
