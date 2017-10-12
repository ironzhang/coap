package deduplication

import (
	"errors"
	"log"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

var (
	ErrStateNotFound = errors.New("not found message state")
	ErrAckNonMessage = errors.New("non message not need ack")
	ErrMessageSaved  = errors.New("message already saved")
)

type state struct {
	Time    time.Time
	Type    uint8
	Saved   bool
	Message base.Message
}

func (s *state) Timeout(d time.Duration) bool {
	return time.Since(s.Time) > d
}

func (s *state) PutMessage(m base.Message) bool {
	if s.Saved {
		return false
	}
	s.Saved = true
	s.Message = m
	return true
}

func (s *state) GetMessage() (base.Message, bool) {
	return s.Message, s.Saved
}

var _ base.Layer = &Layer{}

type Layer struct {
	base.BaseLayer
	NonLifetime      time.Duration
	ExchangeLifetime time.Duration

	states map[uint16]*state
}

func NewLayer() *Layer {
	return &Layer{
		BaseLayer:        base.BaseLayer{Name: "deduplication"},
		NonLifetime:      base.NON_LIFETIME,
		ExchangeLifetime: base.EXCHANGE_LIFETIME,
		states:           make(map[uint16]*state),
	}
}

func (l *Layer) Update() {
	for id, s := range l.states {
		if l.timeout(s) {
			delete(l.states, id)
		}
	}
}

func (l *Layer) Recv(m base.Message) error {
	if m.Type != base.CON && m.Type != base.NON {
		return l.BaseLayer.Recv(m)
	}

	s, ok := l.getState(m.MessageID)
	if !ok {
		return l.recv(m)
	}

	switch {
	case s.Type == base.NON && m.Type == base.NON:
		// 正常分支，忽略重复的NON消息
		return nil

	case s.Type == base.CON && m.Type == base.CON:
		// 正常分支，忽略或回复保存的消息
		if msg, ok := s.GetMessage(); ok && (msg.Token == "" || msg.Token == m.Token) {
			log.Printf("ack message(%s) for duplicate con message(%s)", msg, m)
			if err := l.BaseLayer.Send(msg); err != nil {
				log.Printf("send: %v", err)
			}
		}
		return nil

	case s.Type == base.NON && m.Type == base.CON:
		// 异常分支，回复RST
		if err := l.BaseLayer.SendRST(m.MessageID); err != nil {
			log.Printf("send rst: %v", err)
		}
		return nil

	case s.Type == base.CON && m.Type == base.NON:
		// 异常分支，忽略消息
		return nil
	}

	panic("never arrive")
}

func (l *Layer) Send(m base.Message) error {
	if m.Type != base.ACK && m.Type != base.RST {
		return l.BaseLayer.Send(m)
	}

	// 保存消息
	s, ok := l.getState(m.MessageID)
	if !ok {
		return l.BaseLayer.NewError(ErrStateNotFound)
	}
	if s.Type == base.NON && m.Type == base.ACK {
		// NON消息不可能有ACK
		return l.BaseLayer.NewError(ErrAckNonMessage)
	}
	if !s.PutMessage(m) {
		// 消息已保存过
		return l.BaseLayer.NewError(ErrMessageSaved)
	}

	return l.BaseLayer.Send(m)
}

func (l *Layer) getState(id uint16) (*state, bool) {
	s, ok := l.states[id]
	if !ok {
		return nil, false
	}
	if l.timeout(s) {
		return nil, false
	}
	return s, true
}

func (l *Layer) timeout(s *state) bool {
	switch s.Type {
	case base.CON:
		return s.Timeout(l.ExchangeLifetime)
	case base.NON:
		return s.Timeout(l.NonLifetime)
	}
	return true
}

func (l *Layer) recv(m base.Message) error {
	l.states[m.MessageID] = &state{Time: time.Now(), Type: m.Type}
	return l.BaseLayer.Recv(m)
}
