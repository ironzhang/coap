package block1

import (
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

type client struct {
	baseLayer *base.BaseLayer
	generator func() uint16
	blockSize uint32
	timeout   time.Duration

	busy           bool
	timestamp      time.Time
	message        base.Message
	buffer         base.BlockBuffer
	blockMessageID uint16
}

func (c *client) init(baseLayer *base.BaseLayer, generator func() uint16, blockSize uint32, timeout time.Duration) {
	c.baseLayer = baseLayer
	c.generator = generator
	c.blockSize = blockSize
	c.timeout = timeout
}

func (c *client) update() {
	if c.busy && time.Since(c.timestamp) > c.timeout {
		c.busy = false
	}
}

func (c *client) send(m base.Message) error {
	if c.busy {
		return c.baseLayer.NewError(base.ErrClientBusy)
	}
	if len(m.Payload) <= int(c.blockSize) {
		return c.baseLayer.Send(m)
	}
	c.busy = true
	c.timestamp = time.Now()
	c.message = m
	c.buffer = m.Payload
	return c.sendBlockMessage(m.MessageID, 0, c.blockSize)
}

func (c *client) recv(m base.Message) error {
	if !c.busy {
		return c.baseLayer.Recv(m)
	}
	switch m.Code {
	case base.Continue:
		return c.handleContinue(m)
	case base.RequestEntityIncomplete:
		fallthrough
	case base.RequestEntityTooLarge:
		fallthrough
	default:
		return c.handleError(m)
	}
}

func (c *client) onAckTimeout(m base.Message) {
	if m.MessageID != c.blockMessageID {
		c.baseLayer.OnAckTimeout(m)
	}
	c.busy = false
	c.baseLayer.OnAckTimeout(c.message)
}

func (c *client) handleContinue(m base.Message) error {
	if c.blockMessageID != m.MessageID {
		return c.baseLayer.NewError(base.ErrUnexpectMessageID)
	}
	opt, ok := base.ParseBlock1Option(m)
	if !ok {
		return c.baseLayer.NewError(base.ErrNoBlock1Option)
	}
	if opt.More {
		return c.sendBlockMessage(c.generator(), opt.Num+1, opt.Size)
	}
	c.busy = false
	m.MessageID = c.message.MessageID
	return c.baseLayer.Recv(m)
}

func (c *client) handleError(m base.Message) error {
	c.busy = false
	m.MessageID = c.message.MessageID
	return c.baseLayer.Recv(m)
}

func (c *client) sendBlockMessage(messageID uint16, seq, size uint32) error {
	opt, payload, err := c.buffer.Read(seq, size)
	if err != nil {
		return err
	}
	c.blockMessageID = messageID
	m := base.Message{
		Type:      base.CON,
		Code:      c.message.Code,
		MessageID: messageID,
		Payload:   payload,
	}
	if !opt.More {
		m.Token = c.message.Token
		m.Options = c.message.Options
	}
	m.SetOption(base.Block1, opt.Value())
	return c.baseLayer.Send(m)
}
