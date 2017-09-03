package block2

import (
	"errors"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/internal/stack/blockwise/block"
)

type server struct {
	baseLayer *base.BaseLayer
	blockSize uint32

	busy      bool
	timestamp time.Time
	message   base.Message
	buffer    buffer
}

func (s *server) init(baseLayer *base.BaseLayer, blockSize uint32) {
	s.baseLayer = baseLayer
	s.blockSize = blockSize
}

func (s *server) send(m base.Message) error {
	if s.isBusy() {
		return errors.New("transmitter is busy")
	}
	if len(m.Payload) <= int(s.blockSize) {
		return s.baseLayer.Send(m)
	}
	s.busy = true
	s.timestamp = time.Now()
	s.message = m
	s.buffer.Reset(m.Payload)
	return s.sendBlockMessage(m.MessageID, 0, s.blockSize)
}

func (t *server) recv(m base.Message) error {
	if !t.isBusy() {
		return t.baseLayer.Recv(m)
	}
	opt, ok := block.ParseBlock2Option(m)
	if !ok {
		return errors.New("no block2 option")
	}
	return t.sendBlockMessage(m.MessageID, opt.Num, opt.Size)
}

func (t *server) sendBlockMessage(messageID uint16, seq, size uint32) error {
	opt, payload, err := t.buffer.Read(seq, size)
	if err != nil {
		return err
	}
	m := base.Message{
		Type:      base.ACK,
		Code:      t.message.Code,
		MessageID: messageID,
		Payload:   payload,
	}
	if !opt.More {
		t.busy = false
		m.Token = t.message.Token
		m.Options = t.message.Options
	}
	m.SetOption(base.Block2, opt.Value())
	return t.baseLayer.Send(m)
}

func (t *server) isBusy() bool {
	return t.busy
}
