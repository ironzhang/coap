package block1

import (
	"errors"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/internal/stack/blockwise/block"
)

type MessageIDGenerator func() uint16

type transmitter struct {
	baseLayer *base.BaseLayer
	generator MessageIDGenerator
	blockSize uint32

	busy      bool
	timestamp time.Time
	code      uint8
	messageID uint16
	token     string
	options   []base.Option
	reader    *block.Reader

	blockMessageID uint16
}

func (t *transmitter) init(baseLayer *base.BaseLayer, generator MessageIDGenerator, blockSize uint32) {
	t.baseLayer = baseLayer
	t.generator = generator
	t.blockSize = blockSize
}

func (t *transmitter) send(m base.Message) error {
	if t.isBusy() {
		return errors.New("transmitter is busy")
	}
	if len(m.Payload) <= int(t.blockSize) {
		return t.baseLayer.Send(m)
	}

	t.busy = true
	t.timestamp = time.Now()
	t.code = m.Code
	t.messageID = m.MessageID
	t.token = m.Token
	t.options = m.Options
	t.reader = block.NewReader(m.Payload)
	return t.sendBlockMessage(m.MessageID, 0, t.blockSize)
}

func (t *transmitter) recv(m base.Message) error {
	if !t.isBusy() {
		return t.baseLayer.Recv(m)
	}
	if t.blockMessageID != m.MessageID {
		return errors.New("unexpect block message id")
	}
	block1Opt, ok := block.ParseBlock1Option(m)
	if !ok {
		return errors.New("no block1 option")
	}
	if block1Opt.More {
		return t.sendBlockMessage(t.generator(), block1Opt.Num+1, block1Opt.Size)
	}
	t.busy = false
	m.MessageID = t.messageID
	return t.baseLayer.Recv(m)
}

func (t *transmitter) sendBlockMessage(messageID uint16, seq, size uint32) error {
	block1Opt, payload, err := t.reader.Read(seq, size)
	if err != nil {
		return err
	}
	t.blockMessageID = messageID
	m := base.Message{
		Type:      base.CON,
		Code:      t.code,
		MessageID: messageID,
		Payload:   payload,
	}
	if !block1Opt.More {
		m.Token = t.token
		m.Options = t.options
	}
	m.SetOption(base.Block1, block1Opt.Value())
	return t.baseLayer.Send(m)
}

func (t *transmitter) isBusy() bool {
	return t.busy
}
