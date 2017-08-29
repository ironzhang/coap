package coap

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Request COAP请求
type Request struct {
	// 是否为可靠消息
	Confirmable bool

	// 请求方法
	Method Code

	// COAP选项
	Options Options

	// 目标url
	URL *url.URL

	// 消息令牌, 消息接收端使用, 发送端不应该使用该字段
	Token string

	// 消息负载
	Payload []byte

	// 远端地址, 消息接收端使用, 发送段不应该使用该字段
	RemoteAddr net.Addr

	// 请求超时时间, 消息发送端使用
	Timeout time.Duration

	// 若设置该字段，发送请求时使用Request中的Token字段，否则消息的token自动生成
	useToken bool
}

// NewRequest 构造COAP请求.
func NewRequest(confirmable bool, method Code, urlstr string, payload []byte) (*Request, error) {
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

func splitHostPort(hostport string) (string, uint32, error) {
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
	return host, uint32(n), nil
}
