package coap

import (
	"errors"
	"net"
	"net/url"
	"sync/atomic"
)

type Client struct {
	Handler Handler
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

	var closed int64
	defer func() {
		atomic.StoreInt64(&closed, 1)
		conn.Close()
	}()

	sess := newSession(conn, c.Handler, nil, conn.LocalAddr(), conn.RemoteAddr())
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

func (c *Client) dialUDP(url *url.URL) (net.Conn, error) {
	conn, err := net.Dial("udp", url.Host)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
