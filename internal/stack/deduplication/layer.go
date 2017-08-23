package deduplication

import (
	"errors"
	"log"
	"time"

	"github.com/ironzhang/coap/internal/stack/layer"
	"github.com/ironzhang/coap/message"
)

var (
	ErrStateNotFound = errors.New("not found message state")
	ErrAckNonMessage = errors.New("non message not need ack")
	ErrMessageSaved  = errors.New("message already saved")
)

type state struct {
	Time    time.Time
	Type    message.Type
	Saved   bool
	Message message.Message
}

func (s *state) Timeout(d time.Duration) bool {
	return time.Since(s.Time) > d
}

func (s *state) SaveMessage(m message.Message) bool {
	if s.Saved {
		return false
	}
	s.Saved = true
	s.Message = m
	return true
}

func (s *state) GetMessage() (message.Message, bool) {
	return s.Message, s.Saved
}

var _ layer.Layer = &Layer{}

type Layer struct {
	layer.BaseLayer
	NonLifetime      time.Duration
	ExchangeLifetime time.Duration

	states map[uint16]*state
}

func NewLayer() *Layer {
	return &Layer{
		BaseLayer:        layer.BaseLayer{Name: "deduplication"},
		NonLifetime:      145 * time.Second,
		ExchangeLifetime: 247 * time.Second,
		states:           make(map[uint16]*state),
	}
}

func (l *Layer) Update() {
}

func (l *Layer) Recv(m message.Message) error {
	if m.Type != message.CON && m.Type != message.NON {
		return l.BaseLayer.Recv(m)
	}

	s, ok := l.getState(m.MessageID)
	if !ok {
		return l.recv(m)
	}

	switch {
	case s.Type == message.NON && m.Type == message.NON:
		// 正常分支，忽略重复的NON消息
		return nil

	case s.Type == message.CON && m.Type == message.CON:
		// 正常分支，忽略或回复保存的消息
		if msg, ok := s.GetMessage(); ok {
			if err := l.BaseLayer.Send(msg); err != nil {
				log.Printf("send: %v", err)
			}
		}
		return nil

	case s.Type == message.NON && m.Type == message.CON:
		// 异常分支，回复RST
		if err := l.BaseLayer.SendRST(m.MessageID); err != nil {
			log.Printf("send rst: %v", err)
		}
		return nil

	case s.Type == message.CON && m.Type == message.NON:
		// 异常分支，忽略消息
		return nil
	}

	panic("never arrive")
}

func (l *Layer) Send(m message.Message) error {
	if m.Type != message.ACK && m.Type != message.RST {
		return l.BaseLayer.Send(m)
	}

	// 保存消息
	s, ok := l.getState(m.MessageID)
	if !ok {
		return l.BaseLayer.NewError(ErrStateNotFound)
	}
	if s.Type == message.NON && m.Type == message.ACK {
		// NON消息不可能有ACK
		return l.BaseLayer.NewError(ErrAckNonMessage)
	}
	if !s.SaveMessage(m) {
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
	case message.CON:
		return s.Timeout(l.ExchangeLifetime)
	case message.NON:
		return s.Timeout(l.NonLifetime)
	}
	return true
}

func (l *Layer) recv(m message.Message) error {
	l.states[m.MessageID] = &state{Time: time.Now(), Type: m.Type}
	return l.BaseLayer.Recv(m)
}
