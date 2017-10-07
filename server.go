package coap

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/ironzhang/coap/internal/gctable"
)

var ErrSessionNotFound = errors.New("session not found")

// ListenAndServe 在指定地址端口监听并提供COAP服务.
func ListenAndServe(network, address string, h Handler, o Observer) error {
	return (&Server{
		Handler:  h,
		Observer: o,
	}).ListenAndServe(network, address)
}

// Server 定义了运行一个COAP Server的参数
type Server struct {
	Handler  Handler  // 请求响应接口
	Observer Observer // 观察者接口
	Scheme   string

	sessions gctable.Table
}

// ListenAndServe 在指定地址端口监听并提供COAP服务.
func (s *Server) ListenAndServe(network, address string) error {
	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return err
	}

	ln, err := net.ListenUDP(network, addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.Scheme = "coap"
	return s.Serve(ln)
}

// Serve 提供COAP服务.
func (s *Server) Serve(l net.PacketConn) error {
	if s.Scheme == "" {
		s.Scheme = "coap"
	}
	if s.Scheme != "coap" && s.Scheme != "coaps" {
		return errors.New("invalid scheme")
	}

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

// SendRequest 发送COAP请求.
func (s *Server) SendRequest(req *Request) (*Response, error) {
	addr, err := net.ResolveUDPAddr("udp", req.URL.Host)
	if err != nil {
		return nil, err
	}
	sess, ok := s.getSession(addr)
	if !ok {
		return nil, ErrSessionNotFound
	}
	return sess.postRequestAndWaitResponse(req)
}

// Observe 订阅.
//
// token长度不能大于8个字节.
func (s *Server) Observe(token, urlstr string, accept uint32) error {
	if len(token) > 8 {
		return errors.New("invalid token")
	}
	req, err := NewRequest(true, GET, urlstr, nil)
	if err != nil {
		return err
	}
	req.useToken = true
	req.Token = token
	req.Options.Set(Observe, 0)
	req.Options.Set(Accept, accept)
	return s.postRequestAndWaitAck(req)
}

// CancelObserve 取消订阅.
func (s *Server) CancelObserve(urlstr string) error {
	req, err := NewRequest(true, GET, urlstr, nil)
	if err != nil {
		return err
	}
	req.Options.Set(Observe, 1)
	return s.postRequestAndWaitAck(req)
}

func (s *Server) postRequestAndWaitAck(req *Request) error {
	addr, err := net.ResolveUDPAddr("udp", req.URL.Host)
	if err != nil {
		return err
	}
	sess, ok := s.getSession(addr)
	if !ok {
		return ErrSessionNotFound
	}
	return sess.postRequestAndWaitAck(req)
}

func (s *Server) addSession(conn net.PacketConn, addr net.Addr) *session {
	obj := s.sessions.Add(addr.String(), func() gctable.Object {
		return newSession(&serverConn{conn: conn, addr: addr}, s.Handler, s.Observer, conn.LocalAddr(), addr, s.Scheme)
	})
	return obj.(*session)
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
