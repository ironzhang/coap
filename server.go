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
func ListenAndServe(address string, h Handler, o Observer) error {
	return (&Server{
		Handler:  h,
		Observer: o,
	}).ListenAndServe(address)
}

// Server 定义了运行一个COAP Server的参数
type Server struct {
	Handler    Handler  // 请求响应接口
	Observer   Observer // 观察者接口
	ReadBytes  int      // 读缓冲大小
	WriteBytes int      // 写缓冲大小

	sessions gctable.Table
}

// ListenAndServe 在指定地址端口监听并提供COAP服务.
func (s *Server) ListenAndServe(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	if s.ReadBytes > 0 {
		ln.SetReadBuffer(s.ReadBytes)
	}
	if s.WriteBytes > 0 {
		ln.SetWriteBuffer(s.WriteBytes)
	}

	return s.Serve("coap", ln)
}

// Serve 提供COAP服务.
func (s *Server) Serve(scheme string, l net.PacketConn) error {
	if scheme == "" {
		scheme = "coap"
	}
	if scheme != "coap" && scheme != "coaps" {
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
		s.addSession(scheme, l, addr).recvData(data)
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
	return sess.postRequestWithCache(req)
}

// Observe 订阅.
//
// token长度不能大于8个字节, 且需要保证token永不重复.
func (s *Server) Observe(token Token, urlstr string, accept uint32) (*Response, error) {
	if len(token) > 8 {
		return nil, errors.New("invalid token")
	}
	req, err := NewRequest(true, GET, urlstr, nil)
	if err != nil {
		return nil, err
	}
	req.useToken = true
	req.Token = token
	req.Options.Set(Observe, 0)
	req.Options.Set(Accept, accept)
	return s.postRequestAndWaitResponse(req)
}

// CancelObserve 取消订阅.
func (s *Server) CancelObserve(urlstr string, accept uint32) (*Response, error) {
	req, err := NewRequest(true, GET, urlstr, nil)
	if err != nil {
		return nil, err
	}
	req.Options.Set(Observe, 1)
	req.Options.Set(Accept, accept)
	return s.postRequestAndWaitResponse(req)
}

func (s *Server) postRequestAndWaitResponse(req *Request) (*Response, error) {
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

func (s *Server) addSession(scheme string, conn net.PacketConn, addr net.Addr) *session {
	obj := s.sessions.Add(addr.String(), func() gctable.Object {
		return newSession(&serverConn{conn: conn, addr: addr}, s.Handler, s.Observer, conn.LocalAddr(), addr, scheme)
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
