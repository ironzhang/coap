package coap

import (
	"reflect"
	"testing"

	"github.com/ironzhang/coap/internal/stack/base"
)

func TestRequestKey(t *testing.T) {
	tests := []struct {
		method Code
		url    string
		key    string
	}{
		{
			method: GET,
			url:    "coap://localhost/",
			key:    "GET coap://localhost:5683/",
		},
		{
			method: GET,
			url:    "coaps://localhost/",
			key:    "GET coaps://localhost:5684/",
		},
		{
			method: PUT,
			url:    "coap://localhost/hello",
			key:    "PUT coap://localhost:5683/hello",
		},
	}
	for i, tt := range tests {
		req, err := NewRequest(true, tt.method, tt.url, nil)
		if err != nil {
			t.Fatalf("case%d: new request: %v", i, err)
		}
		if got, want := requestKey(req), tt.key; got != want {
			t.Errorf("case%d: %v != %v", i, got, want)
		}
	}
}

func TestIsCacheStatus(t *testing.T) {
	tests := []struct {
		status Code
		yes    bool
	}{
		{Created, false},
		{Changed, false},
		{Content, true},
		{BadRequest, true},
		{BadOption, true},
		{InternalServerError, true},
		{ServiceUnavailable, true},
	}
	for i, tt := range tests {
		if got, want := isCacheStatus(tt.status), tt.yes; got != want {
			t.Errorf("case%d: %s: %v != %v", i, tt.status, got, want)
		}
	}
}

func TestCloneOptionsExclude(t *testing.T) {
	tests := []struct {
		src Options
		dst Options
	}{
		{
			src: Options{},
			dst: Options{},
		},
		{
			src: Options{
				{IfMatch, []byte{1, 2}},
			},
			dst: Options{
				{IfMatch, []byte{1, 2}},
			},
		},
		{
			src: Options{
				{IfMatch, []byte{1, 2}},
				{URIHost, "localhost"},
			},
			dst: Options{
				{IfMatch, []byte{1, 2}},
				{URIHost, "localhost"},
			},
		},
		{
			src: Options{
				{IfMatch, []byte{1, 2}},
				{URIHost, "localhost"},
				{Size1, 1024},
			},
			dst: Options{
				{IfMatch, []byte{1, 2}},
				{URIHost, "localhost"},
			},
		},
	}
	for i, tt := range tests {
		if got, want := cloneOptionsExclude(tt.src, func(o base.Option) bool { return base.NoCacheKey(o.ID) }), tt.dst; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: %v != %v", i, got, want)
		}
	}
}

func TestOptionsEqual(t *testing.T) {
	tests := []struct {
		src Options
		dst Options
	}{
		{
			src: Options{},
			dst: Options{},
		},
		{
			src: Options{
				{IfMatch, []byte{1, 2}},
			},
			dst: Options{
				{IfMatch, []byte{1, 2}},
			},
		},
		{
			src: Options{
				{IfMatch, []byte{1, 2}},
				{URIHost, "localhost"},
			},
			dst: Options{
				{IfMatch, []byte{1, 2}},
				{URIHost, "localhost"},
			},
		},
		{
			src: Options{
				{IfMatch, []byte{1, 2}},
				{URIHost, "localhost"},
				{Size1, 1024},
			},
			dst: Options{
				{IfMatch, []byte{1, 2}},
				{URIHost, "localhost"},
				{Size2, 1024},
			},
		},
	}
	for i, tt := range tests {
		if !optionsEqual(tt.src, tt.dst) {
			t.Errorf("case%d: %v != %v", i, tt.src, tt.dst)
		}
	}
}

func TestCache(t *testing.T) {
	req0, _ := NewRequest(true, POST, "coap://localhost/test", nil)
	req1, _ := NewRequest(true, DELETE, "coap://localhost/test", nil)
	req2, _ := NewRequest(true, GET, "coap://localhost/test", nil)
	req3, _ := NewRequest(true, POST, "coap://localhost/post", nil)
	tests := []struct {
		req    *Request
		resp   *Response
		cached bool
	}{
		{
			req:    req0,
			resp:   &Response{Status: Created},
			cached: false,
		},
		{
			req:    req1,
			resp:   &Response{Status: Deleted},
			cached: false,
		},
		{
			req:    req2,
			resp:   &Response{Status: Content, Payload: []byte("test")},
			cached: true,
		},
		{
			req:    req3,
			resp:   &Response{Status: BadRequest, Payload: []byte("bad request")},
			cached: true,
		},
	}

	var c cache
	for _, tt := range tests {
		c.Add(tt.req, tt.resp)
	}
	for i, tt := range tests {
		resp, ok := c.Get(tt.req)
		if got, want := ok, tt.cached; got != want {
			t.Errorf("case%d: cached: %v != %v", i, got, want)
		}
		if ok {
			if got, want := resp, tt.resp; !reflect.DeepEqual(got, want) {
				t.Errorf("case%d: response: %v != %v", i, got, want)
			}
		}
	}
}
