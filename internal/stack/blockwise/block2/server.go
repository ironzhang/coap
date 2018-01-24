package block2

import (
	"errors"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

type sstate struct {
	start  time.Time
	source base.Message
	buffer base.BlockBuffer
}

type sstatus struct {
	states map[string]*sstate
}

func (p *sstatus) add(m base.Message) (*sstate, error) {
	if p.states == nil {
		p.states = make(map[string]*sstate)
	}
	if _, ok := p.states[m.Token]; ok {
		return nil, errors.New("token duplicate")
	}
	s := &sstate{start: time.Now(), source: m, buffer: m.Payload}
	p.states[m.Token] = s
	return s, nil
}

func (p *sstatus) del(token string) {
	delete(p.states, token)
}

func (p *sstatus) get(token string) (*sstate, bool) {
	s, ok := p.states[token]
	return s, ok
}

func (p *sstatus) update(timeout time.Duration) {
	for token, s := range p.states {
		if time.Since(s.start) > timeout {
			delete(p.states, token)
		}
	}
}

type server struct {
	base      *base.BaseLayer
	blockSize uint32
	timeout   time.Duration
	status    sstatus
}

func (s *server) init(b *base.BaseLayer, blockSize uint32, timeout time.Duration) {
	s.base = b
	s.blockSize = blockSize
	s.timeout = timeout
}

func (s *server) Update() {
	s.status.update(s.timeout)
}

func (s *server) Send(m base.Message) error {
	if len(m.Payload) <= int(s.blockSize) {
		return s.base.Send(m)
	}
	state, err := s.status.add(m)
	if err != nil {
		return s.base.NewError(err)
	}
	return s.sendBlockMessage(m.MessageID, state, 0, s.blockSize)
}

func (s *server) Recv(m base.Message) error {
	state, ok := s.status.get(m.Token)
	if !ok {
		opt, ok := base.ParseBlock2Option(m)
		if ok {
			s.blockSize = opt.Size
		}
		return s.base.Recv(m)
	}
	opt, ok := base.ParseBlock2Option(m)
	if !ok {
		return s.base.NewError(base.ErrNoBlock2Option)
	}
	s.blockSize = opt.Size
	return s.sendBlockMessage(m.MessageID, state, opt.Num, opt.Size)
}

func (s *server) sendBlockMessage(messageID uint16, state *sstate, num, size uint32) error {
	opt, payload, err := state.buffer.Read(num, size)
	if err != nil {
		return s.base.NewError(err)
	}
	if !opt.More {
		s.status.del(state.source.Token)
	}
	m := base.Message{
		Type:      base.ACK,
		Code:      state.source.Code,
		MessageID: messageID,
		Token:     state.source.Token,
		Options:   state.source.Options,
		Payload:   payload,
	}
	m.SetOption(base.Block2, opt.Value())
	return s.base.Send(m)
}
