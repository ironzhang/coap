package block2

import (
	"bytes"
	"errors"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/internal/stack/blockwise/block"
)

type client struct {
	baseLayer *base.BaseLayer
	generator func() uint16

	busy           bool
	timestamp      time.Time
	messageID      uint16
	buffer         bytes.Buffer
	blockMessageID uint16
}

func (c *client) init(baseLayer *base.BaseLayer, generator func() uint16) {
	c.baseLayer = baseLayer
	c.generator = generator
}

func (c *client) recv(m base.Message) error {
	opt, ok := block.ParseBlock2Option(m)
	if !ok {
		return c.baseLayer.Recv(m)
	}
	if c.buffer.Len() != int(opt.Num*opt.Size) {
		return errors.New("request entity incomplete")
	}
	if c.busy {
		if c.blockMessageID != m.MessageID {
			return errors.New("request entity incomplete")
		}
	} else {
		c.busy = true
		c.timestamp = time.Now()
		c.messageID = m.MessageID
	}
	c.buffer.Write(m.Payload)
	if opt.More {
		return c.getNextBlock(opt.Num+1, opt.Size)
	}
	c.busy = false
	m.MessageID = c.messageID
	m.Payload = c.copyAndResetBuffer()
	return c.baseLayer.Recv(m)
}

func (c *client) send(m base.Message) error {
	return c.baseLayer.Send(m)
}

func (c *client) getNextBlock(num, size uint32) error {
	c.blockMessageID = c.generator()
	o := block.Option{
		Num:  num,
		Size: size,
	}
	m := base.Message{
		Type:      base.CON,
		Code:      base.GET,
		MessageID: c.blockMessageID,
	}
	m.SetOption(base.Block2, o.Value())
	return c.baseLayer.Send(m)
}

func (c *client) copyAndResetBuffer() []byte {
	b := make([]byte, c.buffer.Len())
	copy(b, c.buffer.Bytes())
	c.buffer.Reset()
	return b
}
