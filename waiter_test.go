package coap

import (
	"io"
	"testing"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

func TestAckWaiter(t *testing.T) {
	w0 := newAckWaiter()
	w0.Done(nil)
	if got, want := w0.Wait(), error(nil); got != want {
		t.Errorf("case0: %v != %v", got, want)
	}

	w1 := newAckWaiter()
	time.AfterFunc(10*time.Millisecond, func() { w1.Done(io.EOF) })
	if got, want := w1.Wait(), io.EOF; got != want {
		t.Errorf("case1: %v != %v", got, want)
	}
}

func TestResponseWaiterReturnNil(t *testing.T) {
	w := newResponseWaiter()
	time.AfterFunc(10*time.Millisecond, func() { w.Done(base.Message{Code: base.Created, Token: "1"}, nil) })
	resp, err := w.Wait()
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if resp.Status != Created {
		t.Errorf("status: %v", resp.Token)
	}
	if resp.Token != "1" {
		t.Errorf("token: %v", resp.Token)
	}
}

func TestResponseWaiterReturnErr(t *testing.T) {
	w := newResponseWaiter()
	time.AfterFunc(10*time.Millisecond, func() { w.Done(base.Message{}, io.EOF) })
	_, err := w.Wait()
	if err != io.EOF {
		t.Errorf("wait: %v != %v", err, io.EOF)
	}
}

func TestResponseWaiterTimeout(t *testing.T) {
	w := newResponseWaiter()
	w.timeout = 200 * time.Millisecond
	if got, want := w.Timeout(), false; got != want {
		t.Errorf("first: %v != %v", got, want)
	}
	time.Sleep(100 * time.Millisecond)
	if got, want := w.Timeout(), false; got != want {
		t.Errorf("second: %v != %v", got, want)
	}
	time.Sleep(100 * time.Millisecond)
	if got, want := w.Timeout(), true; got != want {
		t.Errorf("third: %v != %v", got, want)
	}
}
