package coap

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Handler 响应COAP请求的接口
type Handler interface {
	ServeCOAP(ResponseWriter, *Request)
}

// ResponseWriter 用于构造COAP响应
type ResponseWriter interface {
	// Ack 回复空ACK，服务器无法立即响应，可先调用该方法返回一个空的ACK
	Ack(Code) error

	// SetConfirmable 设置响应为可靠消息，作为单独响应时生效
	SetConfirmable()

	// Options 返回Options
	Options() Options

	// WriteCode 设置响应状态码，默认为Content
	WriteCode(Code)

	// Write 写入payload
	Write([]byte) (int, error)
}

// response 实现了ResponseWriter接口
type response struct {
	session     *session
	confirmable bool
	messageID   uint16
	token       []byte
	acked       bool
	code        Code
	options     Options
	buffer      bytes.Buffer
}

func (r *response) Ack(code Code) error {
	r.acked = true
	m := message{
		Type:      ACK,
		Code:      code,
		MessageID: r.messageID,
	}
	return r.session.send(m)
}

func (r *response) SetConfirmable() {
	r.confirmable = true
}

func (r *response) Options() Options {
	return r.options
}

func (r *response) WriteCode(code Code) {
	r.code = code
}

func (r *response) Write(p []byte) (int, error) {
	return r.buffer.Write(p)
}

type session struct {
	svr   *Server
	conn  net.PacketConn
	addr  net.Addr
	recvc chan []byte

	seq       uint16
	pending   map[uint16]chan struct{}
	callbacks map[string]func(s *Server, r *Response)

	//outgoingMessages map[uint16]rawMessage
}

func newSession(svr *Server, conn net.PacketConn, addr net.Addr) *session {
	return new(session).init(svr, conn, addr)
}

func (s *session) init(svr *Server, conn net.PacketConn, addr net.Addr) *session {
	s.svr = svr
	s.conn = conn
	s.addr = addr
	s.recvc = make(chan []byte, 16)
	go s.inputing()
	return s
}

func (s *session) close() {
	close(s.recvc)
}

func (s *session) inputing() {
	for data := range s.recvc {
		m, err := parseMessage(data)
		if err != nil {
			log.Printf("parse message: %v", err)
			continue
		}
		s.serve(m)
	}
}

func parseMessage(data []byte) (m message, err error) {
	err = m.Unmarshal(data)
	return m, err
}

func (s *session) serve(m message) {
	switch m.Type {
	case CON, NON:
		s.handleMSG(m)
	case ACK:
		s.handleACK(m)
	case RST:
		s.handleRST(m)
	default:
	}
}

func (s *session) handleMSG(m message) {
	if m.Code == 0 {
		// 空消息
		return
	}

	c := m.Code >> 5
	switch {
	case c == 0:
		// 请求
		s.handleRequest(m)
	case c >= 2 && c <= 5:
		// 响应
		s.handleResponse(m)
	default:
		// 保留
		log.Printf("reserved code: %d.%d", c, m.Code&0x1f)
	}
}

func (s *session) handleACK(m message) {
	if done, ok := s.pending[m.MessageID]; ok {
		delete(s.pending, m.MessageID)
		close(done)
	}

	if len(m.Token) > 0 {
		token := string(m.Token)
		cb, ok := s.callbacks[token]
		if ok {
			delete(s.callbacks, token)
			resp := &Response{
				Ack:     true,
				Status:  m.Code,
				Options: m.Options,
				Token:   m.Token,
				Payload: m.Payload,
			}
			cb(s.svr, resp)
		}
	}
}

func (s *session) handleRST(m message) {
	//delete(s.outgoingMessages, m.MessageID)
}

func (s *session) handleRequest(m message) {
	// 去重处理

	// 处理请求
	req := &Request{
		Confirmable: m.Type == CON,
		Method:      m.Code,
		Options:     m.Options,
		Token:       m.Token,
		Payload:     m.Payload,
	}
	resp := &response{
		session:     s,
		confirmable: req.Confirmable,
		messageID:   m.MessageID,
		token:       m.Token,
		acked:       false,
		code:        Content,
	}
	s.svr.Handler.ServeCOAP(resp, req)
	if err := s.sendResponse(resp); err != nil {
		log.Printf("send response: %v", err)
	}
}

func (s *session) handleResponse(m message) {
	if len(m.Token) > 0 {
		token := string(m.Token)
		cb, ok := s.callbacks[token]
		if ok {
			delete(s.callbacks, token)
			resp := &Response{
				Ack:     false,
				Status:  m.Code,
				Options: m.Options,
				Token:   m.Token,
				Payload: m.Payload,
			}
			cb(s.svr, resp)
		}
	}
}

func (s *session) genToken() string {
	return ""
}

func (s *session) genMessageID() uint16 {
	return 1
}

func (s *session) sendRequest(req *Request, cb func(*Server, *Response)) error {
	token := s.genToken()
	req.Options.SetPath(req.URL.Path)
	m := message{
		Type:      NON,
		Code:      req.Method,
		MessageID: s.genMessageID(),
		Token:     []byte(token),
		Options:   req.Options,
		Payload:   req.Payload,
	}
	if req.Confirmable {
		m.Type = CON
	}

	s.pending[m.MessageID] = make(chan struct{})
	s.callbacks[token] = cb

	return s.send(m)
}

func (s *session) sendResponse(resp *response) error {
	// 附带响应
	if !resp.acked {
		m := message{
			Type:      ACK,
			Code:      resp.code,
			MessageID: resp.messageID,
			Token:     resp.token,
			Options:   resp.options,
			Payload:   resp.buffer.Bytes(),
		}
		return s.send(m)
	}

	// 单独响应
	if resp.code != Content || resp.buffer.Len() > 0 {
		m := message{
			Type:      NON,
			Code:      resp.code,
			MessageID: s.genMessageID(),
			Token:     resp.token,
			Options:   resp.options,
			Payload:   resp.buffer.Bytes(),
		}
		if resp.confirmable {
			m.Type = CON
		}
		return s.send(m)
	}

	return nil
}

// send 发送消息
func (s *session) send(m message) error {
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	_, err = s.conn.WriteTo(data, s.addr)
	return err
}

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
		sess := s.addSession(l, addr)
		sess.recvc <- data
	}
}

// SendRequest 发送COAP请求
func (s *Server) SendRequest(req *Request, cb func(*Server, *Response)) error {
	addr, err := net.ResolveUDPAddr("udp", req.URL.Host)
	if err != nil {
		return err
	}
	sess, ok := s.getSession(addr)
	if !ok {
		return fmt.Errorf("session(%s) not found", addr)
	}
	return sess.sendRequest(req, cb)
}

func (s *Server) addSession(conn net.PacketConn, addr net.Addr) *session {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessions == nil {
		s.sessions = make(map[string]*session)
	}
	sess, ok := s.sessions[addr.String()]
	if !ok {
		sess = newSession(s, conn, addr)
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
