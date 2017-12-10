package main

import (
	"fmt"
	"log"

	"github.com/ironzhang/coap"
)

type Server struct {
	coap.Server
}

func (s *Server) ListenAndServe(address string) error {
	s.Server.Handler = s
	return s.Server.ListenAndServe(address)
}

func (s *Server) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	switch r.URL.Path {
	case "/TestConRequest":
		s.TestConRequest(w, r)
	case "/TestNonRequest":
	case "/TestDeduplication":
	case "/TestBlock1":
	case "/TestBlock2":
	case "/TestCache":
	default:
		w.WriteCode(coap.NotFound)
		fmt.Fprintf(w, "%q path not found", r.URL.Path)
	}
}

func (s *Server) TestConRequest(w coap.ResponseWriter, r *coap.Request) {
}

func main() {
	addr := ":5683"
	log.Printf("listen and serve on %q", addr)

	var s Server
	if err := s.ListenAndServe(addr); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
