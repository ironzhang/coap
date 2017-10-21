package coap_test

import (
	"log"
	"sync"
	"testing"

	"github.com/ironzhang/coap"
)

type TestCOAPHandler struct{}

func (h TestCOAPHandler) ServeCOAP(w coap.ResponseWriter, r *coap.Request) {
	w.Write(r.Payload)
}

func ListenAndServeTestCOAP(addr string) {
	if err := coap.ListenAndServe(addr, TestCOAPHandler{}, nil); err != nil {
		log.Fatalf("coap listen and serve: %v", err)
	}
}

func init() {
	coap.Verbose = 0
	go ListenAndServeTestCOAP(":5683")
}

func TestCOAP(t *testing.T) {
	tests := []struct {
		confirmable bool
		method      coap.Code
		urlstr      string
		payload     []byte
	}{
		{
			confirmable: true,
			method:      coap.PUT,
			urlstr:      "coap://localhost",
			payload:     []byte("hello"),
		},
		{
			confirmable: false,
			method:      coap.POST,
			urlstr:      "coap://127.0.0.1",
			payload:     []byte("hello"),
		},
	}
	for i, tt := range tests {
		req, err := coap.NewRequest(tt.confirmable, tt.method, tt.urlstr, tt.payload)
		if err != nil {
			t.Fatalf("case%d: coap new request: %v", i, err)
		}
		resp, err := coap.DefaultClient.SendRequest(req)
		if err != nil {
			t.Fatalf("case%d: coap send request: %v", i, err)
		}
		if got, want := resp.Status, coap.Content; got != want {
			t.Errorf("case%d: response status: %v != %v", i, got, want)
		}
		if got, want := string(resp.Payload), string(tt.payload); got != want {
			t.Errorf("case%d: response payload: %v != %v", i, got, want)
		}
	}
}

func BenchmarkSerialSendRequest(b *testing.B) {
	payload := []byte("hello")
	for i := 0; i < b.N; i++ {
		req, err := coap.NewRequest(true, coap.POST, "coap://localhost", payload)
		if err != nil {
			b.Fatalf("coap new request: %v", err)
		}
		_, err = coap.DefaultClient.SendRequest(req)
		if err != nil {
			b.Fatalf("coap send request: %v", err)
		}
	}
}

func BenchmarkParallelSendRequest(b *testing.B) {
	payload := []byte("hello")
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		req, err := coap.NewRequest(true, coap.POST, "coap://localhost", payload)
		if err != nil {
			b.Fatalf("coap new request: %v", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := coap.DefaultClient.SendRequest(req)
			if err != nil {
				b.Fatalf("coap send request: %v", err)
			}
		}()
	}
	wg.Wait()
}
