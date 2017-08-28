package coap

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ironzhang/coap/internal/message"
)

type Request struct {
	Confirmable bool
	Method      message.Code
	Options     Options
	URL         *url.URL
	Token       string
	Payload     []byte
	RemoteAddr  net.Addr
	Timeout     time.Duration

	// 若设置该字段，发送请求时使用Request中的Token字段，否则消息的token自动生成
	useToken bool
}

func NewRequest(confirmable bool, method message.Code, urlstr string, payload []byte) (*Request, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "coap" && u.Scheme != "coaps" {
		return nil, errors.New("invalid scheme")
	}
	if u.Fragment != "" {
		return nil, errors.New("unsupport fragment")
	}
	host, port, err := splitHostPort(u.Host)
	if err != nil {
		return nil, err
	}

	options := Options{}
	if net.ParseIP(host) == nil {
		options.Set(URIHost, host)
	}
	if port == 0 {
		if u.Scheme == "coaps" {
			u.Host += ":5684"
		} else {
			u.Host += ":5683"
		}
	} else {
		options.Set(URIPort, port)
	}
	options.SetPath(u.Path)
	options.SetQuery(u.RawQuery)
	r := &Request{
		Confirmable: confirmable,
		Method:      method,
		Options:     options,
		URL:         u,
		Payload:     payload,
	}
	return r, nil
}

func splitHostPort(hostport string) (string, uint16, error) {
	if !strings.Contains(hostport, ":") {
		return hostport, 0, nil
	}
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", 0, err
	}
	if len(host) <= 0 {
		return "", 0, errors.New("invalid host")
	}
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", 0, err
	}
	return host, uint16(n), nil
}
