package coap

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/ironzhang/coap/internal/message"
	"github.com/ironzhang/coap/internal/stack"
	"github.com/ironzhang/coap/internal/stack/base"
)

var Verbose = true

const defaultResponseTimeout = 20 * time.Second

var (
	ErrReset   = errors.New("wait response reset by peer")
	ErrTimeout = errors.New("wait response timeout")
)

// Handler 响应COAP请求的接口
type Handler interface {
	ServeCOAP(ResponseWriter, *Request)
}

// Observer 观察者
type Observer interface {
	ServeObserve(*Response)
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
	code        message.Code
	options     Options
	buffer      bytes.Buffer
	acked       bool
	needAck     bool
}

func (r *response) Ack(code message.Code) {
	if r.needAck {
		r.acked = true
		m := message.Message{
			Type:      ACK,
			Code:      code,
			MessageID: r.messageID,
		}
		r.session.postMessage(m)
	}
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

type session struct {
	writer     io.Writer
	handler    Handler
	observer   Observer
	localAddr  net.Addr
	remoteAddr net.Addr
	host       string
	port       uint16

	lastRecvMutex sync.RWMutex
	lastRecvTime  time.Time

	donec    chan struct{}
	servingc chan func()
	runningc chan func()

	// 以下字段只能在running协程中访问
	seq         uint16
	stack       stack.Stack
	ackWaiters  map[uint16]*ackWaiter
	respWaiters map[string]*responseWaiter
}

func newSession(w io.Writer, h Handler, o Observer, la, ra net.Addr) *session {
	return new(session).init(w, h, o, la, ra)
}

func (s *session) init(w io.Writer, h Handler, o Observer, la, ra net.Addr) *session {
	s.writer = w
	s.handler = h
	s.observer = o
	s.localAddr = la
	s.remoteAddr = ra
	host, port, err := net.SplitHostPort(la.String())
	if err == nil {
		s.host = host
		if n, err := strconv.ParseUint(port, 10, 16); err == nil {
			s.port = uint16(n)
		}
	}

	s.donec = make(chan struct{})
	s.servingc = make(chan func(), 8)
	s.runningc = make(chan func(), 8)

	s.stack.Init(s, s, s.ackTimeout)
	s.ackWaiters = make(map[uint16]*ackWaiter)
	s.respWaiters = make(map[string]*responseWaiter)

	go s.serving()
	go s.running()

	return s
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
	t := time.NewTicker(base.ACK_TIMEOUT)
	defer t.Stop()
	for {
		select {
		case <-s.donec:
			close(s.runningc)
			return
		case f := <-s.runningc:
			f()
		case <-t.C:
			s.update()
		}
	}
}

func (s *session) update() {
	s.stack.Update()
	for k, w := range s.respWaiters {
		if w.Timeout() {
			delete(s.respWaiters, k)
			w.Done(message.Message{}, ErrTimeout)
		}
	}
}

func (s *session) Key() string {
	return s.remoteAddr.String()
}

func (s *session) CanGC() bool {
	return s.lastRecvTimeExpired()
}

func (s *session) ExecuteGC() {
	s.Close()
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
		log.Printf("handler is nil")
		if err := s.sendRST(m.MessageID); err != nil {
			log.Printf("send rst: %v", err)
		}
		return
	}
	url, err := s.parseURLFromOptions(m.Options)
	if err != nil {
		log.Printf("parse url from options: %v", err)
		if err := s.sendRST(m.MessageID); err != nil {
			log.Printf("send rst: %v", err)
		}
		return
	}

	// 由serving协程调用上层handler处理请求
	s.servingc <- func() {
		req := &Request{
			Confirmable: m.Type == CON,
			Method:      m.Code,
			Options:     m.Options,
			URL:         url,
			Token:       m.Token,
			Payload:     m.Payload,
			RemoteAddr:  s.remoteAddr,
		}
		resp := &response{
			session:     s,
			confirmable: req.Confirmable,
			messageID:   m.MessageID,
			token:       m.Token,
			code:        Content,
			needAck:     req.Confirmable,
		}
		s.handler.ServeCOAP(resp, req)
		s.postResponse(resp)
	}
}

func (s *session) handleResponse(m message.Message) {
	options := Options(m.Options)
	if _, ok := options.GetOption(Observe); ok {
		s.handleObserveResponse(m)
	} else {
		s.handleNormalResponse(m)
	}
}

func (s *session) handleObserveResponse(m message.Message) {
	if s.observer == nil {
		log.Printf("observer is nil")
		if err := s.sendRST(m.MessageID); err != nil {
			log.Printf("send rst: %v", err)
		}
		return
	}

	// 由serving协程调用上层观察者程序处理订阅响应
	s.servingc <- func() {
		resp := &Response{
			Ack:     m.Type == ACK,
			Status:  m.Code,
			Options: m.Options,
			Token:   m.Token,
			Payload: m.Payload,
		}
		s.observer.ServeObserve(resp)
	}

	// 回复ACK
	if m.Type == CON {
		if err := s.sendACK(m.MessageID); err != nil {
			log.Printf("send ack: %v", err)
		}
	}
}

func (s *session) handleNormalResponse(m message.Message) {
	// 结束响应等待
	s.finishResponseWait(m, nil)

	// 回复ACK
	if m.Type == CON {
		if err := s.sendACK(m.MessageID); err != nil {
			log.Printf("send ack: %v", err)
		}
	}
}

func (s *session) handleACK(m message.Message) {
	s.finishAckWait(m, nil)
	if len(m.Token) > 0 {
		s.finishResponseWait(m, nil)
	}
}

func (s *session) handleRST(m message.Message) {
	s.finishAckWait(m, ErrReset)
	for k, w := range s.respWaiters {
		if w.messageID == m.MessageID {
			delete(s.respWaiters, k)
			w.Done(message.Message{}, ErrReset)
			break
		}
	}
}

func (s *session) Send(m message.Message) error {
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	_, err = s.writer.Write(data)
	return err
}

func (s *session) ackTimeout(m message.Message) {
	s.finishAckWait(m, ErrTimeout)
	if len(m.Token) > 0 {
		s.finishResponseWait(m, ErrTimeout)
	}
}

func (s *session) recvData(data []byte) {
	s.lastRecvTimeUpdate()
	s.runningc <- func() {
		var m message.Message
		if err := m.Unmarshal(data); err != nil {
			log.Printf("message unmarshal: %v", err)
			return
		}
		s.recvMessage(m)
	}
}

func (s *session) recvMessage(m message.Message) {
	if Verbose {
		log.Printf("recv: %s\n", m.String())
	}
	if err := s.stack.Recv(m); err != nil {
		log.Printf("stack recv: %v", err)
	}
}

func (s *session) postMessage(m message.Message) {
	s.runningc <- func() {
		if err := s.sendMessage(m); err != nil {
			log.Printf("send message: %v", err)
		}
	}
}

func (s *session) sendMessage(m message.Message) error {
	if Verbose {
		log.Printf("send: %s\n", m.String())
	}
	return s.stack.Send(m)
}

func (s *session) postResponse(r *response) {
	s.runningc <- func() {
		if err := s.sendResponse(r); err != nil {
			log.Printf("send response: %v", err)
		}
	}
}

func (s *session) sendResponse(r *response) error {
	if !r.needAck {
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

	if r.acked {
		// 单独响应
		if r.code != Content || len(r.options) > 0 || r.buffer.Len() > 0 {
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
	} else {
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

	return nil
}

func (s *session) postRequest(r *Request) {
	s.runningc <- func() {
		if err := s.sendRequest(r, nil, nil); err != nil {
			log.Printf("send request: %v", err)
		}
	}
}

func (s *session) postRequestAndWaitAck(r *Request) error {
	w := newAckWaiter()
	s.runningc <- func() {
		if err := s.sendRequest(r, w, nil); err != nil {
			log.Printf("send request: %v", err)
		}
	}
	return w.Wait()
}

func (s *session) postRequestAndWaitResponse(r *Request) (*Response, error) {
	w := newResponseWaiter()
	if r.Timeout > 0 {
		w.timeout = r.Timeout
	}
	if r.Confirmable && w.timeout < base.EXCHANGE_LIFETIME {
		w.timeout = base.EXCHANGE_LIFETIME
	}
	s.runningc <- func() {
		if err := s.sendRequest(r, nil, w); err != nil {
			log.Printf("send request: %v", err)
		}
	}
	return w.Wait()
}

func (s *session) sendRequest(r *Request, aw *ackWaiter, rw *responseWaiter) error {
	done := func(err error) {
		if aw != nil {
			aw.Done(err)
		}
		if rw != nil {
			rw.Done(message.Message{}, err)
		}
	}

	// 发送消息
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

	// 设置应答等待
	if aw != nil {
		if r.Confirmable {
			s.ackWaiters[m.MessageID] = aw
		} else {
			aw.Done(nil)
		}
	}

	// 设置响应等待
	if rw != nil {
		rw.messageID = m.MessageID
		s.respWaiters[m.Token] = rw
	}

	return nil
}

func (s *session) sendACK(messageID uint16) error {
	m := message.Message{
		Type:      message.ACK,
		MessageID: messageID,
	}
	return s.sendMessage(m)
}

func (s *session) sendRST(messageID uint16) error {
	m := message.Message{
		Type:      message.RST,
		MessageID: messageID,
	}
	return s.sendMessage(m)
}

func (s *session) finishAckWait(m message.Message, err error) {
	if w, ok := s.ackWaiters[m.MessageID]; ok {
		delete(s.ackWaiters, m.MessageID)
		w.Done(err)
	}
}

func (s *session) finishResponseWait(m message.Message, err error) {
	if w, ok := s.respWaiters[m.Token]; ok {
		delete(s.respWaiters, m.Token)
		w.Done(m, err)
	}
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

func (s *session) lastRecvTimeUpdate() {
	//s.lastRecvMutex.Lock()
	s.lastRecvTime = time.Now()
	//s.lastRecvMutex.Unlock()
}

func (s *session) lastRecvTimeExpired() bool {
	//s.lastRecvMutex.Lock()
	//defer s.lastRecvMutex.Unlock()
	if time.Since(s.lastRecvTime) > time.Hour {
		return true
	}
	return false
}

func (s *session) parseURLFromOptions(options Options) (*url.URL, error) {
	scheme := "coap"
	host, ok := options.Get(URIHost).(string)
	if !ok {
		host = s.host
	}
	port, ok := options.Get(URIPort).(uint32)
	if !ok {
		port = uint32(s.port)
	}
	path := options.GetPath()
	query := options.GetQuery()
	return url.Parse(fmt.Sprintf("%s://%s:%d/%s?%s", scheme, host, port, path, query))
}
