package coap

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
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

type TestEchoHandler struct{}

func (h TestEchoHandler) ServeCOAP(w ResponseWriter, r *Request) {
	//log.Printf("%s", r.Payload)
	w.Write(r.Payload)
}

type TestAckHandler struct{}

func (h TestAckHandler) ServeCOAP(w ResponseWriter, r *Request) {
	w.Ack(Changed)
}

func NewTestSession(w io.Writer, h Handler) *session {
	la, err := net.ResolveUDPAddr("udp", "localhost:5683")
	if err != nil {
		panic(err)
	}
	ra, err := net.ResolveUDPAddr("udp", "localhost:5684")
	if err != nil {
		panic(err)
	}
	return newSession(w, h, nil, la, ra, "coap")
}

func SessionRecvData(t *testing.T, s *session, n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		m := base.Message{
			Type:      base.CON,
			Code:      base.PUT,
			MessageID: uint16(i),
			Token:     fmt.Sprint(i),
			Payload:   []byte("hello"),
		}
		data, err := m.Marshal()
		if err != nil {
			t.Fatalf("message marshal: %v", err)
		}

		wg.Add(1)
		go func(data []byte) {
			defer wg.Done()
			s.recvData(data)
		}(data)
	}
	wg.Wait()
}

func TestSessionRecvData0(t *testing.T) {
	s := NewTestSession(&bytes.Buffer{}, TestEchoHandler{})
	SessionRecvData(t, s, 65535)
}

func TestSessionRecvData1(t *testing.T) {
	s := NewTestSession(&bytes.Buffer{}, TestAckHandler{})
	SessionRecvData(t, s, 65535)
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
		{
			in: base.Message{
				Type:      base.NON,
				Code:      base.PUT,
				MessageID: 1,
				Token:     "1",
				Payload:   []byte("hello, world"),
			},
			out: base.Message{
				Type:      base.NON,
				Code:      base.Content,
				MessageID: 101,
				Token:     "1",
				Payload:   []byte("hello, world"),
			},
		},
	}
	for i, tt := range tests {
		var b bytes.Buffer
		var m base.Message
		s := NewTestSession(&b, TestEchoHandler{})
		s.seq = 100
		s.Recv(tt.in)
		time.Sleep(1 * time.Millisecond)
		if err := m.Unmarshal(b.Bytes()); err != nil {
			t.Fatalf("case%d: message unmarshal: %v", i, err)
		}
		if got, want := m, tt.out; !reflect.DeepEqual(got, want) {
			t.Fatalf("case%d: %v != %v", i, got, want)
		}
	}
}

type TestSessionWriter struct {
	s *session
}

func (w *TestSessionWriter) Write(p []byte) (n int, err error) {
	var m base.Message
	if err = m.Unmarshal(p); err != nil {
		return 0, err
	}

	if m.Type == base.CON || m.Type == base.NON {
		m.Code = base.Content
		if m.Type == base.CON {
			m.Type = base.ACK
		}
		data, err := m.Marshal()
		if err != nil {
			return 0, err
		}
		go w.s.recvData(data)
	}
	return len(p), nil
}

func TestSessionSendRequest(t *testing.T) {
	var w TestSessionWriter
	s := NewTestSession(&w, TestEchoHandler{})
	w.s = s

	tests := []struct {
		req  Request
		resp Response
	}{
		{
			req: Request{
				Confirmable: true,
				Method:      PUT,
				Token:       "1",
				Payload:     []byte("hello, world"),
				useToken:    true,
			},
			resp: Response{
				Ack:     true,
				Status:  Content,
				Token:   "1",
				Payload: []byte("hello, world"),
			},
		},
		{
			req: Request{
				Confirmable: false,
				Method:      PUT,
				Token:       "1",
				Payload:     []byte("hello, world"),
				useToken:    true,
			},
			resp: Response{
				Ack:     false,
				Status:  Content,
				Token:   "1",
				Payload: []byte("hello, world"),
			},
		},
	}
	for i, tt := range tests {
		resp, err := s.postRequestAndWaitResponse(&tt.req)
		if err != nil {
			t.Fatalf("case%d: post request and wait response: %v", i, err)
		}
		if got, want := *resp, tt.resp; !reflect.DeepEqual(got, want) {
			t.Fatalf("case%d: %v != %v", i, got, want)
		}
	}
}

func TestSessionPostRequestAndWaitAck(t *testing.T) {
	s := NewTestSession(&bytes.Buffer{}, TestEchoHandler{})

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.seq = 0
			r := &Request{
				Confirmable: true,
				Method:      PUT,
				Token:       "1",
				Payload:     []byte("hello, world"),
				useToken:    true,
			}
			if _, err := s.postRequestAndWaitAck(r); err != nil {
				log.Printf("post request and wait ack: %v", err)
			}
		}()
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
	m := base.Message{
		Type:      base.ACK,
		MessageID: s.seq,
	}
	data, err := m.Marshal()
	if err != nil {
		t.Fatalf("message marshal: %v", err)
	}
	s.recvData(data)
	wg.Wait()
}
