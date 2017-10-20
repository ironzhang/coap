package main

import (
	"fmt"
	"log"

	"github.com/ironzhang/coap"
)

type Server struct {
	svr coap.Server
}

func (s *Server) ListenAndServe(address string) error {
	s.svr.Handler = s
	return s.svr.ListenAndServe(address)
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
	resp, err := s.svr.SendRequest(req)
	if err != nil {
		log.Printf("send request: %v", err)
		return
	}
	log.Printf("%s\n", resp.Payload)
}

func main() {
	var s Server
	if err := s.ListenAndServe(":5683"); err != nil {
		log.Fatal(err)
	}
}
