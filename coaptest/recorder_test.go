package coaptest

import (
	"fmt"
	"testing"

	"github.com/ironzhang/coap"
)

func TestRecorder(t *testing.T) {
	f := func(w coap.ResponseWriter, r *coap.Request) {
		w.SetConfirmable()
		w.WriteCode(coap.Changed)
		fmt.Fprintf(w, "hello, world")
	}
	h := coap.HandlerFunc(f)
	rec := NewRecorder()
	req, _ := coap.NewRequest(false, coap.PUT, "coap://foo.com/", nil)
	h.ServeCOAP(rec, req)
	if got, want := rec.Confirmable, true; got != want {
		t.Errorf("Confirmable: %v != %v", got, want)
	}
	if got, want := rec.Code, coap.Changed; got != want {
		t.Errorf("Code: %v != %v", got, want)
	}
	if got, want := rec.Body.String(), "hello, world"; got != want {
		t.Errorf("Body: %v != %v", got, want)
	}
}
