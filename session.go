package coap

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	mrand "math/rand"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/ironzhang/coap/internal/stack"
	"github.com/ironzhang/coap/internal/stack/base"
)

var (
	Verbose     = 1
	EnableCache = true
)

var (
	ErrReset      = errors.New("wait response reset by peer")
	ErrTimeout    = errors.New("wait response timeout")
	ErrAckTimeout = errors.New("wait ack timeout")
)

// Handler 响应COAP请求的接口
type Handler interface {
	ServeCOAP(ResponseWriter, *Request)
}

// Observer 观察者接口
type Observer interface {
	ServeObserve(*Response)
}

// ResponseWriter 用于构造COAP响应
type ResponseWriter interface {
	// Ack 回复空ACK, 服务器无法立即响应的情况下, 可先调用该方法返回一个空的ACK
	Ack(Code)

	// SetConfirmable 设置响应为可靠消息, 作为单独响应或处理非可靠消息时生效
	SetConfirmable()

	// Options 返回Options
	Options() *Options

	// WriteCode 写入响应状态码, 默认为Content
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
	code        Code
	options     Options
	buffer      bytes.Buffer
	acked       bool
	needAck     bool
}

func (r *response) Ack(code Code) {
	if r.needAck {
		r.acked = true
		m := base.Message{
			Type:      base.ACK,
			Code:      uint8(code),
			MessageID: r.messageID,
		}
		r.session.postMessage(m)
	}
}

func (r *response) SetConfirmable() {
	r.confirmable = true
}

func (r *response) Options() *Options {
	return &r.options
}

func (r *response) WriteCode(code Code) {
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
	scheme     string
	host       string
	port       uint32

	lastRecvMutex sync.RWMutex
	lastRecvTime  time.Time
	cache         cache

	donec    chan struct{}
	servingc chan func()
	runningc chan func()

	// 以下字段只能在running协程中访问
	seq         uint16
	stack       stack.Stack
	ackWaiters  map[uint16]*ackWaiter
	respWaiters map[string]*responseWaiter
}

func newSession(w io.Writer, h Handler, o Observer, la, ra net.Addr, scheme string) *session {
	return new(session).init(w, h, o, la, ra, scheme)
}

func (s *session) init(w io.Writer, h Handler, o Observer, la, ra net.Addr, scheme string) *session {
	s.writer = w
	s.handler = h
	s.observer = o
	s.localAddr = la
	s.remoteAddr = ra
	s.scheme = scheme
	host, port, err := net.SplitHostPort(la.String())
	if err == nil {
		s.host = host
		if n, err := strconv.ParseUint(port, 10, 16); err == nil {
			s.port = uint32(n)
		}
	}

	s.donec = make(chan struct{})
	s.servingc = make(chan func(), 8)
	s.runningc = make(chan func(), 8)

	s.seq = uint16(mrand.Uint32() % math.MaxUint16)
	s.stack.Init(s, s, s.genMessageID)
	s.ackWaiters = make(map[uint16]*ackWaiter)
	s.respWaiters = make(map[string]*responseWaiter)

	go s.serving() // 调用上层回调接口协程
	go s.running() // 主逻辑协程

	return s
}

func (s *session) serving() {
	for {
		select {
		case <-s.donec:
			close(s.runningc)
			return
		case f, ok := <-s.servingc:
			if ok {
				f()
			}
		}
	}
}

func (s *session) running() {
	t := time.NewTicker(base.ACK_TIMEOUT)
	defer t.Stop()
	for {
		select {
		case <-s.donec:
			close(s.servingc)
			return
		case f, ok := <-s.runningc:
			if ok {
				f()
			}
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
			w.Done(base.Message{}, ErrTimeout)
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

func (s *session) OnAckTimeout(m base.Message) {
	s.finishAckWait(m, ErrAckTimeout)
	if len(m.Token) > 0 {
		s.finishResponseWait(m, ErrTimeout)
	}
}

func (s *session) recvData(data []byte) {
	s.lastRecvTimeUpdate()
	s.runningc <- func() {
		var m base.Message
		err := m.Unmarshal(data)
		if err != nil {
			log.Printf("message unmarshal: %v", err)
			handleError(s, m, err)
			return
		}
		s.recvMessage(m)
	}
}

func (s *session) recvMessage(m base.Message) {
	if Verbose == 2 {
		var mser base.MessageStringer
		log.Printf("recv: %s", mser.MessageString(m))
		//log.Printf("recv: %s\n", m.String())
	}

	if err := s.stack.Recv(m); err != nil {
		log.Printf("stack recv: %v", err)
	}
}

func (s *session) Recv(m base.Message) error {
	if Verbose == 1 {
		log.Printf("recv: %s\n", m.String())
	}

	switch m.Type {
	case base.CON, base.NON:
		s.handleMSG(m)
	case base.ACK:
		s.handleACK(m)
	case base.RST:
		s.handleRST(m)
	default:
	}
	return nil
}

func (s *session) handleMSG(m base.Message) {
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

func (s *session) handleRequest(m base.Message) {
	if s.handler == nil {
		log.Printf("handler is nil")
		if err := s.sendRST(m.MessageID); err != nil {
			log.Printf("send rst: %v", err)
		}
		return
	}

	// 将选项编码成URL
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
			Confirmable: m.Type == base.CON,
			Method:      Code(m.Code),
			Options:     m.Options,
			URL:         url,
			Token:       Token(m.Token),
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

func (s *session) handleResponse(m base.Message) {
	options := Options(m.Options)
	if options.Contain(Observe) {
		s.handleObserveResponse(m)
	} else {
		s.handleNormalResponse(m)
	}
}

func (s *session) handleObserveResponse(m base.Message) {
	if s.observer == nil {
		log.Printf("observer is nil")
		if err := s.sendRST(m.MessageID); err != nil {
			log.Printf("send rst: %v", err)
		}
		return
	}

	// 由serving协程调用上层观察者接口处理订阅响应
	s.servingc <- func() {
		resp := &Response{
			Ack:        m.Type == base.ACK,
			Status:     Code(m.Code),
			Options:    m.Options,
			Token:      Token(m.Token),
			Payload:    m.Payload,
			RemoteAddr: s.remoteAddr,
		}
		s.observer.ServeObserve(resp)
	}

	// 回复ACK
	if m.Type == base.CON {
		if err := s.sendACK(m.MessageID); err != nil {
			log.Printf("send ack: %v", err)
		}
	}
}

func (s *session) handleNormalResponse(m base.Message) {
	// 结束响应等待
	s.finishResponseWait(m, nil)

	// 回复ACK
	if m.Type == base.CON {
		if err := s.sendACK(m.MessageID); err != nil {
			log.Printf("send ack: %v", err)
		}
	}
}

func (s *session) handleACK(m base.Message) {
	s.finishAckWait(m, nil)
	if len(m.Token) > 0 {
		s.finishResponseWait(m, nil)
	}

	options := Options(m.Options)
	if options.Contain(Observe) {
		s.handleObserveACK(m)
	}
}

func (s *session) handleObserveACK(m base.Message) {
	if s.observer == nil {
		log.Printf("observer is nil")
		return
	}

	// 由serving协程调用上层观察者接口处理订阅响应
	s.servingc <- func() {
		resp := &Response{
			Ack:        m.Type == base.ACK,
			Status:     Code(m.Code),
			Options:    m.Options,
			Token:      Token(m.Token),
			Payload:    m.Payload,
			RemoteAddr: s.remoteAddr,
		}
		s.observer.ServeObserve(resp)
	}
}

func (s *session) handleRST(m base.Message) {
	s.finishAckWait(m, ErrReset)
	for k, w := range s.respWaiters {
		if w.messageID == m.MessageID {
			delete(s.respWaiters, k)
			w.Done(base.Message{}, ErrReset)
			break
		}
	}
}

func (s *session) Send(m base.Message) error {
	if Verbose == 2 {
		var mser base.MessageStringer
		log.Printf("send: %s", mser.MessageString(m))
		//log.Printf("send: %s\n", m.String())
	}
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	_, err = s.writer.Write(data)
	return err
}

func randomInt64(min, max int64) int64 {
	n := max - min
	if n <= 0 {
		return min
	}
	return min + mrand.Int63n(n)
}

func randomDuration() time.Duration {
	const (
		min = 50 * int64(time.Millisecond)
		max = 500 * int64(time.Millisecond)
	)
	return time.Duration(randomInt64(min, max))
}

func (s *session) postMessage(m base.Message) {
	fn := func() {
		if err := s.sendMessage(m); err != nil {
			log.Printf("send message: %v", err)
		}
	}

	select {
	case s.runningc <- fn:
	default:
		time.AfterFunc(randomDuration(), func() { s.runningc <- fn })
	}
}

func (s *session) sendMessage(m base.Message) error {
	if Verbose == 1 {
		log.Printf("send: %s\n", m.String())
	}
	return s.stack.Send(m)
}

func (s *session) postResponse(r *response) {
	fn := func() {
		if err := s.sendResponse(r); err != nil {
			log.Printf("send response: %v", err)
		}
	}

	select {
	case s.runningc <- fn:
	default:
		time.AfterFunc(randomDuration(), func() { s.runningc <- fn })
	}
}

func (s *session) sendResponse(r *response) error {
	if !r.needAck {
		// 非可靠请求的响应
		m := base.Message{
			Type:      base.NON,
			Code:      uint8(r.code),
			MessageID: s.genMessageID(),
			Token:     r.token,
			Options:   r.options,
			Payload:   r.buffer.Bytes(),
		}
		if r.confirmable {
			m.Type = base.CON
		}
		return s.sendMessage(m)
	}

	if r.acked {
		// 单独响应
		if r.code != Content || len(r.options) > 0 || r.buffer.Len() > 0 {
			m := base.Message{
				Type:      base.NON,
				Code:      uint8(r.code),
				MessageID: s.genMessageID(),
				Token:     r.token,
				Options:   r.options,
				Payload:   r.buffer.Bytes(),
			}
			if r.confirmable {
				m.Type = base.CON
			}
			return s.sendMessage(m)
		}
	} else {
		// 附带响应
		m := base.Message{
			Type:      base.ACK,
			Code:      uint8(r.code),
			MessageID: r.messageID,
			Token:     r.token,
			Options:   r.options,
			Payload:   r.buffer.Bytes(),
		}
		return s.sendMessage(m)
	}

	return nil
}

func (s *session) postRequestAndWaitAck(r *Request) (*Response, error) {
	w := newAckWaiter()
	s.runningc <- func() {
		if err := s.sendRequestWithAckWaiter(r, w); err != nil {
			log.Printf("send request with ack waiter: %v", err)
		}
	}
	return w.Wait()
}

func (s *session) sendRequestWithAckWaiter(r *Request, w *ackWaiter) (err error) {
	defer func() {
		if err != nil {
			w.Done(base.Message{}, err)
		}
	}()

	// 非可靠请求没有ACK
	if !r.Confirmable {
		return fmt.Errorf("send non request with ack waiter")
	}

	// 构造消息
	m := s.makeRequestMessage(r)

	// 检查MessageID
	if _, ok := s.ackWaiters[m.MessageID]; ok {
		return fmt.Errorf("MessageID(%d) duplicate", m.MessageID)
	}

	// 发送消息
	if err = s.sendMessage(m); err != nil {
		return err
	}

	// 设置应答等待
	s.ackWaiters[m.MessageID] = w

	return nil
}

func (s *session) postRequestWithCache(req *Request) (*Response, error) {
	if !EnableCache {
		return s.postRequestAndWaitResponse(req)
	}
	if resp, ok := s.cache.Get(req); ok {
		return resp, nil
	}
	resp, err := s.postRequestAndWaitResponse(req)
	if err != nil {
		return nil, err
	}
	s.cache.Add(req, resp)
	return resp, nil
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
		if err := s.sendRequestWithResponseWaiter(r, w); err != nil {
			log.Printf("send request with response waiter: %v", err)
		}
	}
	return w.Wait()
}

func (s *session) sendRequestWithResponseWaiter(r *Request, w *responseWaiter) (err error) {
	defer func() {
		if err != nil {
			w.Done(base.Message{}, err)
		}
	}()

	// 构造消息
	m := s.makeRequestMessage(r)

	// 检查Token
	if _, ok := s.respWaiters[m.Token]; ok {
		return fmt.Errorf("Token(%s) duplicate", m.Token)
	}

	// 发送消息
	if err = s.sendMessage(m); err != nil {
		return err
	}

	// 设置响应等待
	w.messageID = m.MessageID
	s.respWaiters[m.Token] = w

	return nil
}

func (s *session) makeRequestMessage(r *Request) base.Message {
	m := base.Message{
		Type:      base.NON,
		Code:      uint8(r.Method),
		MessageID: s.genMessageID(),
		Options:   r.Options,
		Payload:   r.Payload,
	}
	if r.Confirmable {
		m.Type = base.CON
	}
	if r.useToken {
		m.Token = string(r.Token)
	} else {
		m.Token = s.genToken()
	}
	return m
}

func (s *session) sendACK(messageID uint16) error {
	m := base.Message{
		Type:      base.ACK,
		MessageID: messageID,
	}
	return s.sendMessage(m)
}

func (s *session) sendRST(messageID uint16) error {
	m := base.Message{
		Type:      base.RST,
		MessageID: messageID,
	}
	return s.sendMessage(m)
}

func (s *session) directSendRST(messageID uint16) error {
	m := base.Message{
		Type:      base.RST,
		MessageID: messageID,
	}
	return s.Send(m)
}

func (s *session) directSendBadOptionACK(messageID uint16, token string) error {
	payload := `Unrecognized options of class "critical" that occur in a Confirmable request`
	m := base.Message{
		Type:      base.ACK,
		Code:      base.BadOption,
		MessageID: messageID,
		Token:     token,
		Payload:   []byte(payload),
	}
	return s.Send(m)
}

func (s *session) finishAckWait(m base.Message, err error) {
	if w, ok := s.ackWaiters[m.MessageID]; ok {
		delete(s.ackWaiters, m.MessageID)
		w.Done(m, err)
	}
}

func (s *session) finishResponseWait(m base.Message, err error) {
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
	scheme := s.scheme
	host, ok := options.Get(URIHost).(string)
	if !ok {
		host = s.host
	}
	port, ok := options.Get(URIPort).(uint32)
	if !ok {
		port = s.port
	}
	path := options.GetPath()
	query := options.GetQuery()
	urlstr := fmt.Sprintf("%s://%s:%d/%s", scheme, host, port, path)
	if len(query) > 0 {
		urlstr = urlstr + "?" + query
	}
	return url.Parse(urlstr)
}
