package reliability

import (
	"errors"
	"time"

	"github.com/ironzhang/coap/internal/message"
	"github.com/ironzhang/coap/internal/stack/layer"
)

var (
	ErrDupMessageID = errors.New("message id duplicate")
)

// state 消息状态
type state struct {
	Start          time.Time
	Message        message.Message
	LastRetransmit time.Time
	Retransmit     int
	Timeout        time.Duration
}

var _ layer.Layer = &Layer{}

type Layer struct {
	layer.BaseLayer
	Timeout         func(message.Message)
	MaxRetransmit   int
	MaxTransmitSpan time.Duration
	AckTimeout      time.Duration
	AckRandomFactor float64

	states map[uint16]*state
}

func NewLayer(timeout func(message.Message)) *Layer {
	return &Layer{
		BaseLayer:       layer.BaseLayer{Name: "reliability"},
		Timeout:         timeout,
		MaxRetransmit:   4,
		MaxTransmitSpan: 45 * time.Second,
		AckTimeout:      2 * time.Second,
		AckRandomFactor: 1.5,
		states:          make(map[uint16]*state),
	}
}

func (l *Layer) Update() {
	for _, s := range l.states {
		if s.Retransmit >= l.MaxRetransmit || time.Since(s.Start) >= l.MaxTransmitSpan {
			l.timeout(s)
			continue
		}

		if time.Since(s.LastRetransmit) >= s.Timeout {
			l.send(s)
		}
	}
}

func (l *Layer) Recv(m message.Message) error {
	if m.Type != message.ACK && m.Type != message.RST {
		return l.BaseLayer.Recv(m)
	}
	if l.delState(m) {
		return l.BaseLayer.Recv(m)
	}
	return nil
}

func (l *Layer) Send(m message.Message) error {
	if m.Type != message.CON {
		return l.BaseLayer.Send(m)
	}
	s, ok := l.addState(m)
	if !ok {
		return l.BaseLayer.NewError(ErrDupMessageID)
	}
	return l.send(s)
}

func (l *Layer) send(s *state) error {
	s.LastRetransmit = time.Now()
	if s.Retransmit == 0 {
		s.Timeout = l.AckTimeout
	} else {
		s.Timeout *= 2
	}
	s.Retransmit++
	return l.BaseLayer.Send(s.Message)
}

func (l *Layer) timeout(s *state) {
	delete(l.states, s.Message.MessageID)
	if l.Timeout != nil {
		l.Timeout(s.Message)
	}
}

func (l *Layer) addState(m message.Message) (*state, bool) {
	if _, ok := l.states[m.MessageID]; ok {
		return nil, false
	}
	s := &state{Start: time.Now(), Message: m}
	l.states[m.MessageID] = s
	return s, true
}

func (l *Layer) delState(m message.Message) bool {
	if _, ok := l.states[m.MessageID]; ok {
		delete(l.states, m.MessageID)
		return true
	}
	return false
}

func (l *Layer) getState(m message.Message) (*state, bool) {
	s, ok := l.states[m.MessageID]
	return s, ok
}
