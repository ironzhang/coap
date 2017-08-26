package coap

import (
	"reflect"
	"testing"

	"github.com/ironzhang/coap/internal/message"
)

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		hostport string
		host     string
		port     uint16
	}{
		{"localhost", "localhost", 0},
		{"localhost:8000", "localhost", 8000},
	}
	for i, tt := range tests {
		host, port, err := splitHostPort(tt.hostport)
		if err != nil {
			t.Fatalf("case%d: split host port: %v", i, err)
		}
		if host != tt.host {
			t.Errorf("case%d: host: %q != %q", i, host, tt.host)
		}
		if port != tt.port {
			t.Errorf("case%d: port: %d != %d", i, port, tt.port)
		}
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		confirmable bool
		method      message.Code
		urlstr      string
		options     Options
	}{
		{
			confirmable: true,
			method:      GET,
			urlstr:      "coap://localhost/1/2/3?a=1&b=2&c=3",
			options: Options{
				{URIHost, []interface{}{"localhost"}},
				{URIPath, []interface{}{"1", "2", "3"}},
				{URIQuery, []interface{}{"a=1", "b=2", "c=3"}},
			},
		},
		{
			confirmable: false,
			method:      POST,
			urlstr:      "coap://127.0.0.1:8000/a/b",
			options: Options{
				{URIPort, []interface{}{uint16(8000)}},
				{URIPath, []interface{}{"a", "b"}},
			},
		},
		{
			confirmable: false,
			method:      POST,
			urlstr:      "coaps://127.0.0.1/",
			options:     Options{},
		},
	}
	for i, tt := range tests {
		req, err := NewRequest(tt.confirmable, tt.method, tt.urlstr, nil)
		if err != nil {
			t.Fatalf("case%d: new request: %v", i, err)
		}
		if got, want := req.Confirmable, tt.confirmable; got != want {
			t.Errorf("case%d: Confirmable: %v != %v", i, got, want)
		}
		if got, want := req.Method, tt.method; got != want {
			t.Errorf("case%d: Method: %v != %v", i, got, want)
		}
		if got, want := req.Options, tt.options; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: Options:\ngot:\n%s\nwant:\n%s\n", i, OptionsString(got), OptionsString(want))
		}
	}
}
