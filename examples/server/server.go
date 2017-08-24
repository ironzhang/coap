package main

import (
	"fmt"
	"log"

	"github.com/ironzhang/coap"
)

type Server struct {
	svr coap.Server
}

func (s *Server) ListenAndServe(network, address string) error {
	s.svr.Handler = s
	return s.svr.ListenAndServe(network, address)
}

func (s *Server) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	w.Ack(coap.Changed)

	log.Printf("%s\n", r.Payload)
	fmt.Fprintf(w, "server echo %s", r.Payload)

	req, err := coap.NewRequest(false, coap.GET, fmt.Sprintf("coap://%s/hi", r.RemoteAddr.String()), []byte("hi"))
	if err != nil {
		log.Printf("new request: %v", err)
		return
	}
	req.Callback = func(resp *coap.Response) {
		if resp.Status == coap.Content {
			log.Printf("%s\n", resp.Payload)
		}
	}

	if err = s.svr.SendRequest(req); err != nil {
		log.Printf("send request: %v", err)
		return
	}
}

func main() {
	var s Server
	if err := s.ListenAndServe("udp", ":5683"); err != nil {
		log.Fatal(err)
	}
}
