package coap

import (
	"bytes"
	"io"
	"log"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

func TestParseURLFromOptions(t *testing.T) {
	s := session{scheme: "coap", host: "localhost", port: 5683}
	tests := []struct {
		options Options
		scheme  string
		host    string
		path    string
		query   string
	}{
		{
			options: Options{
				{URIHost, "www.ablecloud.com"},
				{URIPort, uint32(8000)},
				{URIPath, "1"},
				{URIPath, "2"},
				{URIQuery, "a=1"},
				{URIQuery, "b=2"},
			},
			scheme: "coap",
			host:   "www.ablecloud.com:8000",
			path:   "/1/2",
			query:  "a=1&b=2",
		},
		{
			options: Options{
				{URIPort, uint32(8000)},
				{URIPath, "1"},
				{URIPath, "2"},
				{URIQuery, "a=1"},
				{URIQuery, "b=2"},
			},
			scheme: "coap",
			host:   "localhost:8000",
			path:   "/1/2",
			query:  "a=1&b=2",
		},
		{
			options: Options{
				{URIPath, "1"},
				{URIPath, "2"},
				{URIQuery, "a=1"},
				{URIQuery, "b=2"},
			},
			scheme: "coap",
			host:   "localhost:5683",
			path:   "/1/2",
			query:  "a=1&b=2",
		},
		{
			options: Options{
				{URIQuery, "a=1"},
			},
			scheme: "coap",
			host:   "localhost:5683",
			path:   "/",
			query:  "a=1",
		},
		{
			options: Options{},
			scheme:  "coap",
			host:    "localhost:5683",
			path:    "/",
			query:   "",
		},
	}
	for i, tt := range tests {
		u, err := s.parseURLFromOptions(tt.options)
		if err != nil {
			t.Fatalf("case%d: parse url from options: %v", i, err)
		}
		if got, want := u.Scheme, tt.scheme; got != want {
			t.Errorf("case%d: Scheme: %q != %q", i, got, want)
		}
		if got, want := u.Host, tt.host; got != want {
			t.Errorf("case%d: Host: %q != %q", i, got, want)
		}
		if got, want := u.Path, tt.path; got != want {
			t.Errorf("case%d: Path: %q != %q", i, got, want)
		}
		if got, want := u.RawQuery, tt.query; got != want {
			t.Errorf("case%d: RawQuery: %q != %q", i, got, want)
		}
	}
}

type TestSessionHandler struct{}

func (h TestSessionHandler) ServeCOAP(w ResponseWriter, r *Request) {
	log.Printf("%s", r.Payload)
	w.Write(r.Payload)
}

func NewTestSession(w io.Writer) *session {
	la, err := net.ResolveUDPAddr("udp", "localhost:5683")
	if err != nil {
		panic(err)
	}
	ra, err := net.ResolveUDPAddr("udp", "localhost:5684")
	if err != nil {
		panic(err)
	}
	return newSession(w, TestSessionHandler{}, nil, la, ra, "coap")
}

func TestSessionRecvRequest(t *testing.T) {
	tests := []struct {
		in  base.Message
		out base.Message
	}{
		{
			in: base.Message{
				Type:      base.CON,
				Code:      base.PUT,
				MessageID: 1,
				Token:     "1",
				Payload:   []byte("hello, world"),
			},
			out: base.Message{
				Type:      base.ACK,
				Code:      base.Content,
				MessageID: 1,
				Token:     "1",
				Payload:   []byte("hello, world"),
			},
		},
	}
	for i, tt := range tests {
		var b bytes.Buffer
		var m base.Message
		s := NewTestSession(&b)
		s.Recv(tt.in)
		time.Sleep(10 * time.Millisecond)
		if err := m.Unmarshal(b.Bytes()); err != nil {
			t.Fatalf("case%d: message unmarshal: %v", i, err)
		}
		if got, want := m, tt.out; !reflect.DeepEqual(got, want) {
			t.Fatalf("case%d: %v != %v", i, got, want)
		}
	}
}

func TestSessionSendRequest(t *testing.T) {
}
