package base

import (
	"fmt"
	"io"
	"testing"

	"github.com/ironzhang/coap/internal/message"
)

func TestBaseLayer(t *testing.T) {
	r := CountRecver{}
	s := CountSender{}
	l := BaseLayer{
		Name:   "base",
		Recver: &r,
		Sender: &s,
	}

	l.Recv(message.Message{})
	if got, want := r.Count, 1; got != want {
		t.Errorf("recv count: %v != %v", got, want)
	}

	l.Send(message.Message{})
	l.SendRST(1)
	if got, want := s.Count, 2; got != want {
		t.Errorf("send count: %v != %v", got, want)
	}
}

func TestBaseLayerError(t *testing.T) {
	l := BaseLayer{Name: "base"}
	fmt.Println(l.NewError(io.EOF))
	fmt.Println(l.Errorf(io.EOF, "read a.txt"))
}
