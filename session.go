package coap

import (
	"bytes"
	"errors"
	"io"
	"log"
	"time"
)

var ErrRST = errors.New("reset by peer")

// Handler 响应COAP请求的接口
type Handler interface {
	ServeCOAP(ResponseWriter, *Request)
}

// ResponseWriter 用于构造COAP响应
type ResponseWriter interface {
	// Ack 回复空ACK，服务器无法立即响应，可先调用该方法返回一个空的ACK
	Ack(Code)

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
	token       string
	acked       bool
	code        Code
	options     Options
	buffer      bytes.Buffer
}

func (r *response) Ack(code Code) {
	r.acked = true
	r.session.messagec <- message{
		Type:      ACK,
		Code:      code,
		MessageID: r.messageID,
	}
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

type callback struct {
	ts time.Time
	cb func(*Response)
}

type session struct {
	writer    io.Writer
	handler   Handler
	dones     map[uint16]func(error)
	callbacks map[string]callback

	donec     chan struct{}
	taskc     chan func()
	recvc     chan []byte    // 待处理的数据
	messagec  chan message   // 待发送的消息
	requestc  chan *Request  // 待发送的请求
	responsec chan *response // 待发送的响应
}

func newSession(w io.Writer, h Handler) *session {
	return new(session).init(w, h)
}

func (s *session) init(w io.Writer, h Handler) *session {
	s.writer = w
	s.handler = h
	s.dones = make(map[uint16]func(error))
	s.callbacks = make(map[string]callback)
	s.donec = make(chan struct{})
	s.taskc = make(chan func(), 8)
	s.recvc = make(chan []byte, 8)
	s.messagec = make(chan message, 8)
	s.requestc = make(chan *Request, 8)
	s.responsec = make(chan *response, 8)
	go s.serving()
	go s.running()
	return s
}

func (s *session) close() {
	close(s.donec)
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
		s.taskc <- func() {
			cb.cb(r)
		}
	}
}

func (s *session) serving() {
	for {
		select {
		case <-s.donec:
			return
		case f := <-s.taskc:
			f()
		}
	}
}

func (s *session) running() {
	for {
		select {
		case <-s.donec:
			return
		case data := <-s.recvc:
			s.handle(data)
		case msg := <-s.messagec:
			s.sendMessage(msg)
		case req := <-s.requestc:
			s.sendRequest(req)
		case resp := <-s.responsec:
			s.sendResponse(resp)
		}
	}
}

func (s *session) handle(data []byte) {
	var m message
	if err := m.Unmarshal(data); err != nil {
		log.Printf("message unmarshal: %v", err)
		return
	}
	log.Printf("recv: %s\n", m.String())

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
	s.done(m.MessageID, nil)
	if m.Code != Content || len(m.Payload) > 0 {
		s.callback(&Response{
			Ack:     true,
			Status:  m.Code,
			Options: m.Options,
			Token:   m.Token,
			Payload: m.Payload,
		})
	}
}

func (s *session) handleRST(m message) {
	s.done(m.MessageID, ErrRST)
}

func (s *session) handleRequest(m message) {
	if s.handler == nil {
		m := message{
			Type:      ACK,
			Code:      NotFound,
			MessageID: m.MessageID,
			Token:     m.Token,
		}
		if err := s.sendMessage(m); err != nil {
			log.Printf("send message: %v", err)
			return
		}
		return
	}

	s.taskc <- func() {
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
		s.handler.ServeCOAP(resp, req)
		s.responsec <- resp
	}
}

func (s *session) handleResponse(m message) {
	s.callback(&Response{
		Ack:     false,
		Status:  m.Code,
		Options: m.Options,
		Token:   m.Token,
		Payload: m.Payload,
	})
}

func (s *session) sendMessage(m message) error {
	log.Printf("send: %s\n", m.String())

	data, err := m.Marshal()
	if err != nil {
		return err
	}
	_, err = s.writer.Write(data)
	return err
}

func (s *session) sendRequest(r *Request) {
	// 发送消息
	r.Options.SetPath(r.URL.Path)
	m := message{
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
		r.done(err)
		return
	}

	// 设置Response回调
	if r.Callback != nil {
		s.callbacks[m.Token] = callback{ts: time.Now(), cb: r.Callback}
	}

	if r.Confirmable {
		// 可靠消息待ACK返回后再通知上层发送结果
		s.dones[m.MessageID] = r.done
	} else {
		// 非可靠消息直接通知上层请求发送成功
		r.done(nil)
	}
}

func (s *session) sendResponse(r *response) {
	if !r.acked {
		// 附带响应
		m := message{
			Type:      ACK,
			Code:      r.code,
			MessageID: r.messageID,
			Token:     r.token,
			Options:   r.options,
			Payload:   r.buffer.Bytes(),
		}
		s.sendMessage(m)
		return
	}

	if r.code != Content || r.buffer.Len() > 0 {
		// 单独响应
		m := message{
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
		s.sendMessage(m)
		return
	}
}

func (s *session) genMessageID() uint16 {
	return 1
}

func (s *session) genToken() string {
	return ""
}
