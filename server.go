package coap

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func ListenAndServe(network, address string, h Handler) error {
	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return err
	}
	l, err := net.ListenUDP(network, addr)
	if err != nil {
		return err
	}
	return (&Server{Handler: h}).Serve(l)
}

// Server 定义了运行一个COAP Server的参数
type Server struct {
	Handler Handler // 请求响应接口

	mu       sync.RWMutex
	sessions map[string]*session
}

// Serve 提供COAP服务
func (s *Server) Serve(l net.PacketConn) error {
	buf := make([]byte, 1500)
	for {
		n, addr, err := l.ReadFrom(buf)
		if err != nil {
			log.Printf("listener(%s) read from: %v", l.LocalAddr(), err)
			if e, ok := err.(net.Error); ok {
				if e.Temporary() || e.Timeout() {
					time.Sleep(5 * time.Millisecond)
					continue
				}
			}
			return err
		}
		data := make([]byte, n)
		copy(data, buf)
		s.addSession(l, addr).Recv(data)
	}
	return nil
}

// SendRequest 发送COAP请求
func (s *Server) SendRequest(req *Request) error {
	addr, err := net.ResolveUDPAddr("udp", req.URL.Host)
	if err != nil {
		return err
	}
	sess, ok := s.getSession(addr)
	if !ok {
		return fmt.Errorf("session(%s) not found", addr)
	}
	return sess.SendRequest(req)
}

func (s *Server) addSession(conn net.PacketConn, addr net.Addr) *session {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessions == nil {
		s.sessions = make(map[string]*session)
	}
	sess, ok := s.sessions[addr.String()]
	if !ok {
		sess = newSession(&serverConn{conn: conn, addr: addr}, s.Handler)
		s.sessions[addr.String()] = sess
	}
	return sess
}

func (s *Server) getSession(addr net.Addr) (*session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[addr.String()]
	return sess, ok
}

type serverConn struct {
	conn net.PacketConn
	addr net.Addr
}

func (c *serverConn) Write(p []byte) (int, error) {
	return c.conn.WriteTo(p, c.addr)
}
