package main

import (
	"fmt"
	"log"

	"github.com/ironzhang/coap"
)

type Handler struct {
}

func (h *Handler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	log.Printf("%s\n", r.Payload)
	fmt.Fprintf(w, "client echo %s", r.Payload)
}

func main() {
	client := coap.Client{}
	conn, err := client.Dial("coap://localhost:5683", &Handler{}, nil)
	if err != nil {
		log.Printf("dial: %v", err)
		return
	}
	req, err := coap.NewRequest(true, coap.GET, "coap://localhost:5683/hello", []byte("hello, world"))
	if err != nil {
		log.Printf("new coap request: %v", err)
		return
	}
	resp, err := conn.SendRequest(req)
	if err != nil {
		log.Printf("send coap request: %v", err)
		return
	}
	log.Printf("%s\n", resp.Payload)
}
