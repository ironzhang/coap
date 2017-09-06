package main

import "github.com/ironzhang/coap"

type Handler struct {
}

func (h Handler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	w.Write(r.Payload)
}

func main() {
	coap.ListenAndServe("udp", ":5683", Handler{}, nil)
}
