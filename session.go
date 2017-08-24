package coap

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/ironzhang/coap/internal/message"
	"github.com/ironzhang/coap/internal/stack"
)

var Verbose = true

var ErrRST = errors.New("reset by peer")

// Handler 响应COAP请求的接口
type Handler interface {
	ServeCOAP(ResponseWriter, *Request)
}

// ResponseWriter 用于构造COAP响应
type ResponseWriter interface {
	// Ack 回复空ACK，服务器无法立即响应，可先调用该方法返回一个空的ACK
	Ack(message.Code)

	// SetConfirmable 设置响应为可靠消息，作为单独响应时生效
	SetConfirmable()

	// Options 返回Options
	Options() Options

	// WriteCode 设置响应状态码，默认为Content
	WriteCode(message.Code)

	// Write 写入payload
	Write([]byte) (int, error)
}

// response 实现了ResponseWriter接口
type response struct {
	session     *session
	confirmable bool
	messageID   uint16
	token       string
	acked       bool
	code        message.Code
	options     Options
	buffer      bytes.Buffer
	needAck     bool
}

func (r *response) Ack(code message.Code) {
	r.acked = true
	m := message.Message{
		Type:      ACK,
		Code:      code,
		MessageID: r.messageID,
	}
	r.session.SendMessage(m)
}

func (r *response) SetConfirmable() {
	r.confirmable = true
}

func (r *response) Options() Options {
	return r.options
}

func (r *response) WriteCode(code message.Code) {
	r.code = code
}

func (r *response) Write(p []byte) (int, error) {
	return r.buffer.Write(p)
}

type done struct {
	err error
	ch  chan struct{}
}

func (d *done) Done(err error) {
	d.err = err
	close(d.ch)
}

func (d *done) Wait() error {
	<-d.ch
	return d.err
}

type callback struct {
	ts time.Time
	cb func(*Response)
}

type session struct {
	writer  io.Writer
	addr    net.Addr
	handler Handler

	donec    chan struct{}
	servingc chan func()
	runningc chan func()

	// 以下字段只能在running协程中访问
	seq       uint16
	stack     stack.Stack
	dones     map[uint16]func(error)
	callbacks map[string]callback
}

func newSession(w io.Writer, a net.Addr, h Handler) *session {
	return new(session).Init(w, a, h)
}

func (s *session) Init(w io.Writer, a net.Addr, h Handler) *session {
	s.writer = w
	s.addr = a
	s.handler = h
	s.donec = make(chan struct{})
	s.servingc = make(chan func(), 8)
	s.runningc = make(chan func(), 8)
	s.dones = make(map[uint16]func(error))
	s.callbacks = make(map[string]callback)
	s.stack.Init(s, s)
	go s.serving()
	go s.running()
	return s
}

func (s *session) Close() error {
	close(s.donec)
	return nil
}

func (s *session) Recv(m message.Message) error {
	switch m.Type {
	case CON, NON:
		s.handleMSG(m)
	case ACK:
		s.handleACK(m)
	case RST:
		s.handleRST(m)
	default:
	}
	return nil
}

func (s *session) Send(m message.Message) error {
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	_, err = s.writer.Write(data)
	return err
}

func (s *session) RecvData(data []byte) {
	s.runningc <- func() { s.doRecvData(data) }
}

func (s *session) SendMessage(m message.Message) {
	s.runningc <- func() { s.doSendMessage(m) }
}

func (s *session) SendResponse(r *response) {
	s.runningc <- func() { s.doSendResponse(r) }
}

func (s *session) SendRequest(r *Request) error {
	d := &done{ch: make(chan struct{})}
	s.runningc <- func() { s.doSendRequest(r, d.Done) }
	return d.Wait()
}

func (s *session) serving() {
	for {
		select {
		case <-s.donec:
			close(s.servingc)
			return
		case f := <-s.servingc:
			f()
		}
	}
}

func (s *session) running() {
	for {
		select {
		case <-s.donec:
			close(s.runningc)
			return
		case f := <-s.runningc:
			f()
		}
	}
}

func (s *session) doRecvData(data []byte) {
	var m message.Message
	if err := m.Unmarshal(data); err != nil {
		log.Printf("message unmarshal: %v", err)
		return
	}
	if Verbose {
		log.Printf("recv: %s\n", m.String())
	}
	if err := s.stack.Recv(m); err != nil {
		log.Printf("stack recv: %v", err)
	}
}

func (s *session) handleMSG(m message.Message) {
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

func (s *session) handleRequest(m message.Message) {
	if s.handler == nil {
		resp := message.Message{
			Type:      NON,
			Code:      NotFound,
			MessageID: m.MessageID,
			Token:     m.Token,
		}
		if m.Type == CON {
			resp.Type = ACK
		}
		if err := s.sendMessage(resp); err != nil {
			log.Printf("send message: %v", err)
			return
		}
		return
	}

	// 由serving协程调用上层handler处理请求
	s.servingc <- func() {
		req := &Request{
			Confirmable: m.Type == CON,
			Method:      m.Code,
			Options:     m.Options,
			Token:       m.Token,
			Payload:     m.Payload,
			RemoteAddr:  s.addr,
		}
		resp := &response{
			session:     s,
			confirmable: req.Confirmable,
			messageID:   m.MessageID,
			token:       m.Token,
			acked:       false,
			code:        Content,
			needAck:     req.Confirmable,
		}
		s.handler.ServeCOAP(resp, req)
		s.SendResponse(resp)
	}
}

func (s *session) handleResponse(m message.Message) {
	// 回调上层响应的Response处理函数
	s.callback(&Response{
		Ack:        false,
		Status:     m.Code,
		Options:    m.Options,
		Token:      m.Token,
		Payload:    m.Payload,
		RemoteAddr: s.addr,
	})

	// 对可靠消息响应回复一个空ACK
	if m.Type == CON {
		ack := message.Message{
			Type:      ACK,
			Code:      Content,
			MessageID: m.MessageID,
		}
		if err := s.sendMessage(ack); err != nil {
			log.Printf("send message: %v", err)
		}
	}
}

func (s *session) handleACK(m message.Message) {
	s.done(m.MessageID, nil)
	if m.Code != Content || len(m.Payload) > 0 {
		s.callback(&Response{
			Ack:        true,
			Status:     m.Code,
			Options:    m.Options,
			Token:      m.Token,
			Payload:    m.Payload,
			RemoteAddr: s.addr,
		})
	}
}

func (s *session) handleRST(m message.Message) {
	s.done(m.MessageID, ErrRST)
}

func (s *session) done(id uint16, err error) {
	if done, ok := s.dones[id]; ok {
		delete(s.dones, id)
		done(err)
	}
}

func (s *session) callback(r *Response) {
	if cb, ok := s.callbacks[r.Token]; ok {
		delete(s.callbacks, r.Token)
		s.servingc <- func() { cb.cb(r) }
	}
}

func (s *session) doSendMessage(m message.Message) {
	if err := s.sendMessage(m); err != nil {
		log.Printf("send message: %v", err)
	}
}

func (s *session) sendMessage(m message.Message) error {
	if Verbose {
		log.Printf("send: %s\n", m.String())
	}
	return s.stack.Send(m)
}

func (s *session) doSendResponse(r *response) {
	if err := s.sendResponse(r); err != nil {
		log.Printf("send response: %v", err)
	}
}

func (s *session) sendResponse(r *response) error {
	if !r.acked && r.needAck {
		// 附带响应
		m := message.Message{
			Type:      ACK,
			Code:      r.code,
			MessageID: r.messageID,
			Token:     r.token,
			Options:   r.options,
			Payload:   r.buffer.Bytes(),
		}
		return s.sendMessage(m)
	}

	if r.code != Content || r.buffer.Len() > 0 {
		// 单独响应
		m := message.Message{
			Type:      NON,
			Code:      r.code,
			MessageID: s.genMessageID(),
			Token:     r.token,
			Options:   r.options,
			Payload:   r.buffer.Bytes(),
		}
		if r.confirmable {
			m.Type = CON
		}
		return s.sendMessage(m)
	}
	return nil
}

func (s *session) doSendRequest(r *Request, done func(error)) {
	if err := s.sendRequest(r, done); err != nil {
		log.Printf("send request: %v", err)
	}
}

func (s *session) sendRequest(r *Request, done func(error)) error {
	// 发送消息
	r.Options.SetPath(r.URL.Path)
	m := message.Message{
		Type:      NON,
		Code:      r.Method,
		MessageID: s.genMessageID(),
		Token:     s.genToken(),
		Options:   r.Options,
		Payload:   r.Payload,
	}
	if r.Confirmable {
		m.Type = CON
	}
	if err := s.sendMessage(m); err != nil {
		done(err)
		return err
	}

	// 设置Response回调
	if r.Callback != nil {
		s.callbacks[m.Token] = callback{ts: time.Now(), cb: r.Callback}
	}

	if r.Confirmable {
		// 可靠消息待ACK返回后再通知上层发送结果
		s.dones[m.MessageID] = done
	} else {
		// 非可靠消息直接通知上层请求发送成功
		done(nil)
	}
	return nil
}

func (s *session) genMessageID() uint16 {
	s.seq++
	return s.seq
}

func (s *session) genToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return string(b)
}
