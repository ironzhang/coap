package coap

import (
	"errors"
	"net"
	"sync"
)

type Client struct {
	Handler Handler

	mu    sync.RWMutex
	conns map[string]*clientConn
}

func (c *Client) SendRequest(req *Request) error {
	if req.URL == nil {
		return errors.New("coap: nil Request.URL")
	}
	if len(req.URL.Host) <= 0 {
		return errors.New("coap: invalid Request.URL.Host")
	}

	addr, err := net.ResolveUDPAddr("udp", req.URL.Host)
	if err != nil {
		return err
	}
	conn, err := c.addConn(addr)
	if err != nil {
		return err
	}
	return conn.sess.SendRequest(req)
}

func (c *Client) addConn(addr *net.UDPAddr) (*clientConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conns == nil {
		c.conns = make(map[string]*clientConn)
	}
	conn, ok := c.conns[addr.String()]
	if !ok {
		conn = &clientConn{}
		if err := conn.init(addr, c.Handler); err != nil {
			return nil, err
		}
		c.conns[addr.String()] = conn
	}
	return conn, nil
}

type clientConn struct {
	conn *net.UDPConn
	addr *net.UDPAddr
	sess session
}

func (c *clientConn) init(addr *net.UDPAddr, h Handler) error {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	c.conn = conn
	c.addr = addr
	c.sess.Init(conn, h)
	go c.reading()
	return nil
}

func (c *clientConn) reading() {
	buf := make([]byte, 1500)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			continue
		}
		data := make([]byte, n)
		copy(data, buf)
		c.sess.Recv(data)
	}
}
