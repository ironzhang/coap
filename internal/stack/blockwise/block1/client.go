package block1

import (
	"errors"

	"github.com/ironzhang/coap/internal/stack/base"
)

type cmstate struct {
	deleted   bool
	messageID uint16
	source    base.Message
	buffer    base.BlockBuffer
}

type cmstatus struct {
	states []*cmstate
}

func (p *cmstatus) add(m base.Message) (*cmstate, error) {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if s.messageID == m.MessageID {
			return nil, errors.New("message id duplicate")
		}
	}

	for _, s := range p.states {
		if !s.deleted {
			continue
		}
		s.deleted = false
		s.messageID = m.MessageID
		s.source = m
		s.buffer = m.Payload
		return s, nil
	}
	s := &cmstate{messageID: m.MessageID, source: m, buffer: m.Payload}
	p.states = append(p.states, s)
	return s, nil
}

func (p *cmstatus) del(messageID uint16) {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if s.messageID == messageID {
			s.deleted = true
		}
	}
}

func (p *cmstatus) get(messageID uint16) (*cmstate, bool) {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if s.messageID == messageID {
			return s, true
		}
	}
	return nil, false
}

type client struct {
	base      *base.BaseLayer
	generator func() uint16
	blockSize uint32
	status    cmstatus
}

func (c *client) OnAckTimeout(m base.Message) {
	if state, ok := c.status.get(m.MessageID); ok {
		c.status.del(m.MessageID)
		c.base.OnAckTimeout(state.source)
	} else {
		c.base.OnAckTimeout(m)
	}
}

func (c *client) Send(m base.Message) error {
	if len(m.Payload) <= int(c.blockSize) {
		return c.base.Send(m)
	}
	state, err := c.status.add(m)
	if err != nil {
		return c.base.NewError(err)
	}
	return c.sendBlockMessage(m.MessageID, state, 0, c.blockSize)
}

func (c *client) Recv(m base.Message) error {
	state, ok := c.status.get(m.MessageID)
	if !ok {
		return c.base.Recv(m)
	}
	opt, ok := base.ParseBlock1Option(m)
	if !ok {
		c.status.del(m.MessageID)
		return c.base.NewError(base.ErrNoBlock1Option)
	}
	if !opt.More {
		c.status.del(m.MessageID)
		m.MessageID = state.source.MessageID
		return c.base.Recv(m)
	}
	c.blockSize = opt.Size
	return c.sendBlockMessage(c.generator(), state, opt.Num+1, opt.Size)
}

func (c *client) sendBlockMessage(messageID uint16, state *cmstate, num, size uint32) error {
	state.messageID = messageID
	opt, payload, err := state.buffer.Read(num, size)
	if err != nil {
		return c.base.NewError(err)
	}
	m := base.Message{
		Type:      base.CON,
		Code:      state.source.Code,
		MessageID: messageID,
		Token:     state.source.Token,
		Options:   state.source.Options,
		Payload:   payload,
	}
	m.SetOption(base.Block1, opt.Value())
	return c.base.Send(m)
}
