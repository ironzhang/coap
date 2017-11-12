package coap

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync/atomic"
)

// Conn COAP链接
type Conn struct {
	url    *url.URL
	conn   net.Conn
	sess   *session
	closed int64
}

func newConn(url *url.URL, conn net.Conn, sess *session) *Conn {
	c := &Conn{url: url, conn: conn, sess: sess}
	go c.reading()
	return c
}

func (c *Conn) reading() {
	var buf [1500]byte
	for atomic.LoadInt64(&c.closed) == 0 {
		n, err := c.conn.Read(buf[:])
		if err != nil {
			continue
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		c.sess.recvData(data)
	}
}

// Close 关闭COAP链接
func (c *Conn) Close() error {
	if atomic.CompareAndSwapInt64(&c.closed, 0, 1) {
		c.sess.Close()
		return c.conn.Close()
	}
	return nil
}

// SendRequest 发送COAP请求
func (c *Conn) SendRequest(req *Request) (*Response, error) {
	if atomic.LoadInt64(&c.closed) != 0 {
		return nil, errors.New("conn closed")
	}
	if c.url.Host != req.URL.Host {
		return nil, fmt.Errorf("%q is unacceptable, correct url host is %q", req.URL.Host, c.url.Host)
	}
	return c.sess.postRequestWithCache(req)
}

// Client 定义了运行一个COAP Client的参数
type Client struct {
}

var DefaultClient = &Client{}

// SendRequest 发送COAP请求
func (c *Client) SendRequest(req *Request) (*Response, error) {
	if req.URL == nil {
		return nil, errors.New("coap: nil Request.URL")
	}
	if len(req.URL.Host) <= 0 {
		return nil, errors.New("coap: invalid Request.URL.Host")
	}

	conn, err := c.dialUDP(req.URL)
	if err != nil {
		return nil, err
	}
	sess := newSession(conn, nil, nil, conn.LocalAddr(), conn.RemoteAddr(), req.URL.Scheme)

	var closed int64
	defer func() {
		atomic.StoreInt64(&closed, 1)
		conn.Close()
		sess.Close()
	}()

	go func() {
		var buf [1500]byte
		for atomic.LoadInt64(&closed) == 0 {
			n, err := conn.Read(buf[:])
			if err != nil {
				continue
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			sess.recvData(data)
		}
	}()

	return sess.postRequestAndWaitResponse(req)
}

func (c *Client) dialUDP(u *url.URL) (net.Conn, error) {
	return net.Dial("udp", u.Host)
}

// Dial 建立COAP链接
func (c *Client) Dial(urlstr string, handler Handler, observer Observer) (*Conn, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	_, port, err := splitHostPort(u.Host)
	if err != nil {
		return nil, err
	}
	if port == 0 {
		if u.Scheme == "coaps" {
			u.Host += ":5684"
		} else {
			u.Host += ":5683"
		}
	}
	nc, err := c.dialUDP(u)
	if err != nil {
		return nil, err
	}
	return newConn(u, nc, newSession(nc, handler, observer, nc.LocalAddr(), nc.RemoteAddr(), u.Scheme)), nil
}
