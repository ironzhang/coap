package base

import (
	"io"
	"reflect"
	"testing"
)

func TestBaseLayerError(t *testing.T) {
	l := BaseLayer{Name: "base"}
	got, want := l.NewError(io.EOF), Error{Layer: l.Name, Cause: io.EOF}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewError: %#v != %#v", got, want)
	}
	got, want = l.Errorf(io.EOF, "read a.txt"), Error{Layer: l.Name, Cause: io.EOF, Details: "read a.txt"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Errorf: %v != %v", got, want)
	}
}

func TestBaseLayer(t *testing.T) {
	r := CountRecver{}
	s := CountSender{}
	l := BaseLayer{
		Name:   "base",
		Recver: &r,
		Sender: &s,
	}

	l.Recv(Message{})
	if got, want := r.Count, 1; got != want {
		t.Errorf("recv count: %v != %v", got, want)
	}

	l.Send(Message{})
	l.SendRST(1)
	if got, want := s.Count, 2; got != want {
		t.Errorf("send count: %v != %v", got, want)
	}
}
