package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/ironzhang/coap"
)

type Server struct {
	coap.Server
	cacheCount         int64
	deduplicationCount int64
}

func (s *Server) ListenAndServe(address string) error {
	s.Server.Handler = s
	return s.Server.ListenAndServe(address)
}

func (s *Server) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	switch r.URL.Path {
	case "/TestConRequest", "/TestNonRequest":
		s.TestConOrNonRequest(w, r)
	case "/TestBlock":
		s.TestBlock(w, r)
	case "/TestCache":
		s.TestCache(w, r)
	case "/TestDeduplication":
		s.TestDeduplication(w, r)
	default:
		w.WriteCode(coap.NotFound)
		fmt.Fprintf(w, "%q path not found", r.URL.Path)
	}
}

func (s *Server) TestConOrNonRequest(w coap.ResponseWriter, r *coap.Request) {
	coap.PrintRequest(os.Stdout, r, true)
	w.Write(r.Payload)
}

func (s *Server) TestBlock(w coap.ResponseWriter, r *coap.Request) {
	w.Write(r.Payload)
}

func (s *Server) TestCache(w coap.ResponseWriter, r *coap.Request) {
	n := atomic.AddInt64(&s.cacheCount, 1)
	log.Printf("[TestCache] count=%d", n)
	w.Write(r.Payload)
}

func (s *Server) TestDeduplication(w coap.ResponseWriter, r *coap.Request) {
	n := atomic.AddInt64(&s.deduplicationCount, 1)
	log.Printf("[TestDeduplication] start: count=%d", n)
	defer func() {
		log.Printf("[TestDeduplication] end: count=%d", n)
	}()

	d, err := time.ParseDuration(string(r.Payload))
	if err != nil {
		w.WriteCode(coap.BadRequest)
		fmt.Fprintf(w, "parse duration: %v", err)
		return
	}

	time.Sleep(d)
	fmt.Fprintf(w, "count=%d", n)
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":5683", "address")
	flag.IntVar(&coap.Verbose, "verbose", 0, "verbose")
	flag.Parse()

	var s Server
	log.Printf("listen and serve on %q", addr)
	if err := s.ListenAndServe(addr); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
