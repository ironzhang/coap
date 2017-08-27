package coap

import "testing"

func TestParseURLFromOptions(t *testing.T) {
	s := session{host: "localhost", port: 5683}
	tests := []struct {
		options Options
		scheme  string
		host    string
		path    string
		query   string
	}{
		{
			options: Options{
				{URIHost, []interface{}{"www.ablecloud.com"}},
				{URIPort, []interface{}{uint32(8000)}},
				{URIPath, []interface{}{"1", "2"}},
				{URIQuery, []interface{}{"a=1", "b=2"}},
			},
			scheme: "coap",
			host:   "www.ablecloud.com:8000",
			path:   "/1/2",
			query:  "a=1&b=2",
		},
		{
			options: Options{
				{URIPort, []interface{}{uint32(8000)}},
				{URIPath, []interface{}{"1", "2"}},
				{URIQuery, []interface{}{"a=1", "b=2"}},
			},
			scheme: "coap",
			host:   "localhost:8000",
			path:   "/1/2",
			query:  "a=1&b=2",
		},
		{
			options: Options{
				{URIPath, []interface{}{"1", "2"}},
				{URIQuery, []interface{}{"a=1", "b=2"}},
			},
			scheme: "coap",
			host:   "localhost:5683",
			path:   "/1/2",
			query:  "a=1&b=2",
		},
		{
			options: Options{
				{URIQuery, []interface{}{"a=1"}},
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
