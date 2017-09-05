package block2

import (
	"bytes"
	"errors"
	"log"

	"github.com/ironzhang/coap/internal/stack/base"
)

type cstate struct {
	deleted   bool
	messageID uint16
	source    base.Message
	buffer    bytes.Buffer
}

type cstatus struct {
	states []*cstate
}

func (p *cstatus) add(m base.Message) error {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if s.messageID == m.MessageID {
			return errors.New("message id duplicate")
		}
	}

	for _, s := range p.states {
		if !s.deleted {
			continue
		}
		s.deleted = false
		s.messageID = m.MessageID
		s.source = m
		s.buffer.Reset()
		return nil
	}
	p.states = append(p.states, &cstate{messageID: m.MessageID, source: m})
	return nil
}

func (p *cstatus) del(messageID uint16) {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if s.messageID == messageID {
			s.deleted = true
		}
	}
}

func (p *cstatus) get(messageID uint16) (*cstate, error) {
	for _, s := range p.states {
		if s.deleted {
			continue
		}
		if s.messageID == messageID {
			return s, nil
		}
	}
	return nil, errors.New("client message state not found")
}

type client struct {
	base      *base.BaseLayer
	generator func() uint16
	status    cstatus
}

func (c *client) init(b *base.BaseLayer, f func() uint16) {
	c.base = b
	c.generator = f
}

func (c *client) OnAckTimeout(m base.Message) {
	state, err := c.status.get(m.MessageID)
	if err == nil {
		c.status.del(m.MessageID)
		c.base.OnAckTimeout(state.source)
	} else {
		log.Printf("get state: %v", err)
	}
}

func (c *client) Send(m base.Message) error {
	if err := c.status.add(m); err != nil {
		return c.base.NewError(err)
	}
	return c.base.Send(m)
}

func (c *client) Recv(m base.Message) error {
	state, err := c.status.get(m.MessageID)
	if err != nil {
		return c.base.NewError(err)
	}

	opt, ok := base.ParseBlock2Option(m)
	if !ok {
		c.status.del(m.MessageID)
		return c.base.Recv(m)
	}
	state.buffer.Write(m.Payload)
	if !opt.More {
		c.status.del(m.MessageID)
		m.MessageID = state.source.MessageID
		m.Payload = copyBuffer(state.buffer.Bytes())
		return c.base.Recv(m)
	}
	return c.getNextBlockMessage(state, opt.Num+1, opt.Size)
}

func (c *client) getNextBlockMessage(state *cstate, num, size uint32) error {
	state.messageID = c.generator()
	o := base.BlockOption{
		Num:  num,
		Size: size,
	}
	m := base.Message{
		Type:      base.CON,
		Code:      state.source.Code,
		MessageID: state.messageID,
		Token:     state.source.Token,
		Options:   state.source.Options,
	}
	m.SetOption(base.Block2, o.Value())
	return c.base.Send(m)
}

func copyBuffer(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
