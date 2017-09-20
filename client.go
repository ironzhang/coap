package coap

import (
	"errors"
	"net"
	"net/url"
	"sync/atomic"

	"github.com/ironzhang/dtls"
)

type Client struct {
	Handler    Handler
	DTLSConfig *dtls.Config
}

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
	sess := newSession(conn, c.Handler, nil, conn.LocalAddr(), conn.RemoteAddr())

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
	if u.Scheme == "coaps" {
		return dtls.Dial("udp", u.Host, c.DTLSConfig)
	}
	return net.Dial("udp", u.Host)
}
