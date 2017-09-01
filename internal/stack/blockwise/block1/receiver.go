package block1

import (
	"bytes"
	"errors"

	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/internal/stack/blockwise/block"
)

type receiver struct {
	baseLayer *base.BaseLayer
	buffer    bytes.Buffer
	messageID uint16
	block1    uint32
}

func (r *receiver) init(baseLayer *base.BaseLayer) {
	r.baseLayer = baseLayer
}

func (r *receiver) recv(m base.Message) error {
	block1Opt, ok := block.ParseBlock1Option(m)
	if !ok {
		return r.baseLayer.Recv(m)
	}
	if r.buffer.Len() != int(block1Opt.Num*block1Opt.Size) {
		return errors.New("request entity incomplete")
	}
	r.buffer.Write(m.Payload)
	if block1Opt.More {
		return r.ackContinue(m.MessageID, block1Opt.Value())
	}
	r.messageID = m.MessageID
	r.block1 = block1Opt.Value()
	m.Payload = r.copyAndResetBuffer()
	return r.baseLayer.Recv(m)
}

func (r *receiver) send(m base.Message) error {
	if m.MessageID != r.messageID {
		return r.baseLayer.Send(m)
	}
	m.SetOption(base.Block1, r.block1)
	return r.baseLayer.Send(m)
}

func (r *receiver) ackContinue(messageID uint16, block1 uint32) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.Continue,
		MessageID: messageID,
	}
	m.SetOption(base.Block1, block1)
	return r.baseLayer.Send(m)
}

func (r *receiver) copyAndResetBuffer() []byte {
	b := make([]byte, r.buffer.Len())
	copy(b, r.buffer.Bytes())
	r.buffer.Reset()
	return b
}
