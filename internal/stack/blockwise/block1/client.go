package block1

import (
	"bytes"
	"errors"

	"github.com/ironzhang/coap/internal/stack/base"
)

type client struct {
	baseLayer *base.BaseLayer
	buffer    bytes.Buffer
	messageID uint16
	block1    uint32
}

func (c *client) init(baseLayer *base.BaseLayer) {
	c.baseLayer = baseLayer
}

func (c *client) recv(m base.Message) error {
	block1Opt, ok := base.ParseBlock1Option(m)
	if !ok {
		return c.baseLayer.Recv(m)
	}
	if c.buffer.Len() != int(block1Opt.Num*block1Opt.Size) {
		return errors.New("request entity incomplete")
	}
	c.buffer.Write(m.Payload)
	if block1Opt.More {
		return c.ackContinue(m.MessageID, block1Opt.Value())
	}
	c.messageID = m.MessageID
	c.block1 = block1Opt.Value()
	m.Payload = c.copyAndResetBuffer()
	return c.baseLayer.Recv(m)
}

func (c *client) send(m base.Message) error {
	if m.MessageID != c.messageID {
		return c.baseLayer.Send(m)
	}
	m.SetOption(base.Block1, c.block1)
	return c.baseLayer.Send(m)
}

func (c *client) ackContinue(messageID uint16, block1 uint32) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.Continue,
		MessageID: messageID,
	}
	m.SetOption(base.Block1, block1)
	return c.baseLayer.Send(m)
}

func (c *client) copyAndResetBuffer() []byte {
	b := make([]byte, c.buffer.Len())
	copy(b, c.buffer.Bytes())
	c.buffer.Reset()
	return b
}
