package deduplication

import (
	"errors"
	"log"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

var (
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
		msg, ok := s.GetMessage()
		if !ok {
			return nil
		}
		if msg.Token == "" || msg.Token == m.Token {
			// 正常情况，回复保存的消息
			if err := l.BaseLayer.Send(msg); err != nil {
				log.Printf("send: %v", err)
			}
		} else {
			// 异常情况，回复RST
			log.Printf("ingore %s, send rst for duplicate %s", msg, m)
			if err := l.BaseLayer.SendRST(m.MessageID); err != nil {
				log.Printf("send rst: %v", err)
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
		log.Printf("ingore %s, do nothing", m)
		return nil
	}
	return nil
}

func (l *Layer) Send(m base.Message) error {
	if m.Type != base.ACK && m.Type != base.RST {
		return l.BaseLayer.Send(m)
	}

	// 保存消息
	s, ok := l.getState(m.MessageID)
	if !ok {
		return l.BaseLayer.Send(m)
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
