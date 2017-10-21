package coap

import (
	"bytes"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/ironzhang/coap/internal/stack/base"
)

func NewErrorHandlerSession(w io.Writer) *session {
	la, err := net.ResolveUDPAddr("udp", "localhost:5683")
	if err != nil {
		panic(err)
	}
	ra, err := net.ResolveUDPAddr("udp", "localhost:5684")
	if err != nil {
		panic(err)
	}
	return newSession(w, nil, nil, la, ra, "coap")
}

func TestMessageErrorHandler(t *testing.T) {
	tests := []struct {
		ignore bool
		in     base.Message
		out    base.Message
	}{
		{
			ignore: false,
			in: base.Message{
				Type:      base.CON,
				Code:      base.GET,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{
				Type:      base.RST,
				MessageID: 1,
			},
		},
		{
			ignore: false,
			in: base.Message{
				Type:      base.CON,
				Code:      base.Created,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{
				Type:      base.RST,
				MessageID: 1,
			},
		},
		{
			ignore: false,
			in: base.Message{
				Type:      base.NON,
				Code:      base.GET,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{
				Type:      base.RST,
				MessageID: 1,
			},
		},
		{
			ignore: true,
			in: base.Message{
				Type:      base.NON,
				Code:      base.Created,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{},
		},
		{
			ignore: true,
			in: base.Message{
				Type:      base.ACK,
				Code:      base.Created,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{},
		},
		{
			ignore: true,
			in: base.Message{
				Type:      base.RST,
				MessageID: 1,
			},
			out: base.Message{},
		},
	}
	for i, tt := range tests {
		var b bytes.Buffer
		s := NewErrorHandlerSession(&b)
		messageFormatErrorHandler.handle(s, tt.in, errors.New("message format error"))
		if tt.ignore {
			if b.Len() != 0 {
				t.Fatalf("case%d: not ignore message", i)
			}
		} else {
			var got base.Message
			if err := got.Unmarshal(b.Bytes()); err != nil {
				t.Fatalf("case%d: message unmarshal: %v", i, err)
			}
			if want := tt.out; !reflect.DeepEqual(got, want) {
				t.Fatalf("case%d: %v != %v", i, got, want)
			}
		}
	}
}

func TestBadOptionsErrorHandler(t *testing.T) {
	tests := []struct {
		ignore bool
		in     base.Message
		out    base.Message
	}{
		{
			ignore: false,
			in: base.Message{
				Type:      base.CON,
				Code:      base.GET,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{
				Type:      base.ACK,
				Code:      base.BadOption,
				MessageID: 1,
				Token:     "1",
				Payload:   []byte(`Unrecognized options of class "critical" that occur in a Confirmable request`),
			},
		},
		{
			ignore: false,
			in: base.Message{
				Type:      base.CON,
				Code:      base.Created,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{
				Type:      base.RST,
				MessageID: 1,
			},
		},
		{
			ignore: false,
			in: base.Message{
				Type:      base.NON,
				Code:      base.GET,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{
				Type:      base.RST,
				MessageID: 1,
			},
		},
		{
			ignore: true,
			in: base.Message{
				Type:      base.NON,
				Code:      base.Created,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{},
		},
		{
			ignore: true,
			in: base.Message{
				Type:      base.ACK,
				Code:      base.Created,
				MessageID: 1,
				Token:     "1",
			},
			out: base.Message{},
		},
		{
			ignore: true,
			in: base.Message{
				Type:      base.RST,
				MessageID: 1,
			},
			out: base.Message{},
		},
	}
	for i, tt := range tests {
		var b bytes.Buffer
		s := NewErrorHandlerSession(&b)
		badOptionsErrorHandler.handle(s, tt.in, errors.New("bad options error"))
		if tt.ignore {
			if b.Len() != 0 {
				t.Fatalf("case%d: not ignore message", i)
			}
		} else {
			var got base.Message
			if err := got.Unmarshal(b.Bytes()); err != nil {
				t.Fatalf("case%d: message unmarshal: %v", i, err)
			}
			if want := tt.out; !reflect.DeepEqual(got, want) {
				t.Fatalf("case%d: %v != %v", i, got, want)
			}
		}
	}
}
