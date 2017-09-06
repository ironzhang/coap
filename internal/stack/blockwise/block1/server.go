package block1

import (
	"bytes"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

type sstate struct {
	deleted   bool
	start     time.Time
	waitAck   bool
	token     string
	buffer    bytes.Buffer
	messageID uint16
	block1    uint32
}

func (s *sstate) WaitAck(messageID uint16, block1 uint32) {
	s.waitAck = true
	s.messageID = messageID
	s.block1 = block1
}

type sstatus struct {
	states []*sstate
}

func (p *sstatus) add(token string) *sstate {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if s.waitAck {
			continue
		}
		if s.token == token {
			return s
		}
	}

	for _, s := range p.states {
		if !s.deleted {
			continue
		}
		s.deleted = false
		s.start = time.Now()
		s.waitAck = false
		s.token = token
		s.buffer.Reset()
		return s
	}
	s := &sstate{start: time.Now(), token: token}
	p.states = append(p.states, s)
	return s
}

func (p *sstatus) del(messageID uint16) {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if !s.waitAck {
			continue
		}
		if s.messageID == messageID {
			s.deleted = true
		}
	}
}

func (p *sstatus) get(messageID uint16) (*sstate, bool) {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if !s.waitAck {
			continue
		}
		if s.messageID == messageID {
			return s, true
		}
	}
	return nil, false
}

func (p *sstatus) update(timeout time.Duration) {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if time.Since(s.start) > timeout {
			s.deleted = true
		}
	}
}

type server struct {
	base    *base.BaseLayer
	timeout time.Duration
	status  sstatus
}

func (s *server) init(b *base.BaseLayer, timeout time.Duration) {
	s.base = b
	s.timeout = timeout
}

func (s *server) Update() {
	s.status.update(s.timeout)
}

func (s *server) Recv(m base.Message) error {
	opt, ok := base.ParseBlock1Option(m)
	if !ok {
		return s.base.Recv(m)
	}

	state := s.status.add(m.Token)
	if state.buffer.Len() == int(opt.Num*opt.Size) {
		state.buffer.Write(m.Payload)
		if opt.More {
			return s.ackContinue(m.MessageID, m.Token, opt.Value())
		}
		state.WaitAck(m.MessageID, opt.Value())
		m.Payload = copyBuffer(state.buffer.Bytes())
		return s.base.Recv(m)
	}
	return s.ackIncomplete(m.MessageID, m.Token)
}

func (s *server) Send(m base.Message) error {
	if state, ok := s.status.get(m.MessageID); ok {
		s.status.del(m.MessageID)
		m.SetOption(base.Block1, state.block1)
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

func copyBuffer(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
