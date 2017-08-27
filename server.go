package coap

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ironzhang/coap/internal/gctable"
)

func ListenAndServe(network, address string, h Handler) error {
	return (&Server{Handler: h}).ListenAndServe(network, address)
}

// Server 定义了运行一个COAP Server的参数
type Server struct {
	Handler  Handler  // 请求响应接口
	Observer Observer // 观察者接口
	sessions gctable.Table
}

// ListenAndServe 监听在指定地址并提供COAP服务
func (s *Server) ListenAndServe(network, address string) error {
	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return err
	}
	l, err := net.ListenUDP(network, addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
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
		s.addSession(l, addr).recvData(data)
	}
}

// SendRequest 发送COAP请求
func (s *Server) SendRequest(req *Request) (*Response, error) {
	addr, err := net.ResolveUDPAddr("udp", req.URL.Host)
	if err != nil {
		return nil, err
	}
	sess, ok := s.getSession(addr)
	if !ok {
		return nil, fmt.Errorf("session(%s) not found", addr)
	}
	return sess.postRequestAndWaitResponse(req)
}

func (s *Server) SendObserveRequest(req *Request) error {
	addr, err := net.ResolveUDPAddr("udp", req.URL.Host)
	if err != nil {
		return err
	}
	sess, ok := s.getSession(addr)
	if !ok {
		return fmt.Errorf("session(%s) not found", addr)
	}
	sess.postRequest(req)
	return nil
}

func (s *Server) addSession(conn net.PacketConn, addr net.Addr) *session {
	if obj, ok := s.sessions.Get(addr.String()); ok {
		return obj.(*session)
	}
	sess := newSession(&serverConn{conn: conn, addr: addr}, s.Handler, s.Observer, conn.LocalAddr(), addr)
	s.sessions.Add(sess)
	return sess
}

func (s *Server) getSession(addr net.Addr) (*session, bool) {
	if obj, ok := s.sessions.Get(addr.String()); ok {
		return obj.(*session), true
	}
	return nil, false
}

type serverConn struct {
	conn net.PacketConn
	addr net.Addr
}

func (c *serverConn) Write(p []byte) (int, error) {
	return c.conn.WriteTo(p, c.addr)
}
