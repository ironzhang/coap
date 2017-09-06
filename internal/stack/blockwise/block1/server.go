package block1

import (
	"bytes"
	"errors"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

type blockState struct {
	start  time.Time
	buffer bytes.Buffer
}

type ackState struct {
	start  time.Time
	block1 uint32
}

type smstatus struct {
	acks   map[uint16]ackState
	blocks map[string]*blockState
}

func (p *smstatus) init() {
	p.acks = make(map[uint16]ackState)
	p.blocks = make(map[string]*blockState)
}

func (p *smstatus) getBlockState(token string) *blockState {
	s, ok := p.blocks[token]
	if !ok {
		s = &blockState{start: time.Now()}
		p.blocks[token] = s
	}
	return s
}

func (p *smstatus) getAckBlock1(messageID uint16) (uint32, bool) {
	if ack, ok := p.acks[messageID]; ok {
		return ack.block1, true
	}
	return 0, false
}

func (p *smstatus) changeToAckState(token string, messageID uint16, block1 uint32) error {
	s, ok := p.blocks[token]
	if !ok {
		return errors.New("message state not found")
	}
	if _, ok = p.acks[messageID]; ok {
		return errors.New("message id duplicate")
	}
	delete(p.blocks, token)
	p.acks[messageID] = ackState{start: s.start, block1: block1}
	return nil
}

func (p *smstatus) finishAckState(messageID uint16) {
	delete(p.acks, messageID)
}

func (p *smstatus) update(timeout time.Duration) {
	for id, ack := range p.acks {
		if time.Since(ack.start) > timeout {
			delete(p.acks, id)
		}
	}
	for token, block := range p.blocks {
		if time.Since(block.start) > timeout {
			delete(p.blocks, token)
		}
	}
}

type server struct {
	base    *base.BaseLayer
	timeout time.Duration
	status  smstatus
}

func (s *server) init(b *base.BaseLayer, timeout time.Duration) {
	s.base = b
	s.timeout = timeout
	s.status.init()
}

func (s *server) Update() {
	s.status.update(s.timeout)
}

func (s *server) Recv(m base.Message) error {
	opt, ok := base.ParseBlock1Option(m)
	if !ok {
		return s.base.Recv(m)
	}

	state := s.status.getBlockState(m.Token)
	if state.buffer.Len() != int(opt.Num*opt.Size) {
		return s.ackIncomplete(m.MessageID, m.Token)
	}
	state.buffer.Write(m.Payload)
	if opt.More {
		return s.ackContinue(m.MessageID, m.Token, opt.Value())
	}
	if err := s.status.changeToAckState(m.Token, m.MessageID, opt.Value()); err != nil {
		return s.base.NewError(err)
	}
	m.Payload = state.buffer.Bytes()
	return s.base.Recv(m)
}

func (s *server) Send(m base.Message) error {
	if block1, ok := s.status.getAckBlock1(m.MessageID); ok {
		s.status.finishAckState(m.MessageID)
		m.SetOption(base.Block1, block1)
		return s.base.Send(m)
	}
	return s.base.Send(m)
}

func (s *server) ackIncomplete(messageID uint16, token string) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.RequestEntityIncomplete,
		MessageID: messageID,
		Token:     token,
	}
	return s.base.Send(m)
}

func (s *server) ackContinue(messageID uint16, token string, block1 uint32) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.Continue,
		MessageID: messageID,
		Token:     token,
	}
	m.SetOption(base.Block1, block1)
	return s.base.Send(m)
}
