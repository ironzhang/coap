package coap

import (
	"io"
	"testing"
)

func TestDone(t *testing.T) {
	d := done{ch: make(chan struct{})}
	go d.Done(io.EOF)
	if err := d.Wait(0); err != io.EOF {
		t.Error("done wait result unexpect")
	}
}

func TestSession(t *testing.T) {
}
