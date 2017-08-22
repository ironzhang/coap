package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ironzhang/coap"
)

type Handler struct {
}

func (h *Handler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	log.Printf("%s\n", r.Payload)
	fmt.Fprintf(w, "client echo %s", r.Payload)
}

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
	c.Handler = &Handler{}
	if err = c.SendRequest(req); err != nil {
		log.Printf("send coap request: %v", err)
		return
	}

	time.Sleep(100 * time.Millisecond)
}
