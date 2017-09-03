package block1

import (
	"errors"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/internal/stack/blockwise/block"
)

type MessageIDGenerator func() uint16

type server struct {
	baseLayer *base.BaseLayer
	generator MessageIDGenerator
	blockSize uint32

	busy      bool
	timestamp time.Time
	message   base.Message
	reader    *block.Reader

	blockMessageID uint16
}

func (s *server) init(baseLayer *base.BaseLayer, generator MessageIDGenerator, blockSize uint32) {
	s.baseLayer = baseLayer
	s.generator = generator
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
	s.reader = block.NewReader(m.Payload)
	return s.sendBlockMessage(m.MessageID, 0, s.blockSize)
}

func (s *server) recv(m base.Message) error {
	if !s.isBusy() {
		return s.baseLayer.Recv(m)
	}
	if s.blockMessageID != m.MessageID {
		return errors.New("unexpect block message id")
	}
	block1Opt, ok := block.ParseBlock1Option(m)
	if !ok {
		return errors.New("no block1 option")
	}
	if block1Opt.More {
		return s.sendBlockMessage(s.generator(), block1Opt.Num+1, block1Opt.Size)
	}
	s.busy = false
	m.MessageID = s.message.MessageID
	return s.baseLayer.Recv(m)
}

func (s *server) onAckTimeout(m base.Message) {
	if m.MessageID != s.blockMessageID {
		s.baseLayer.OnAckTimeout(m)
	}
	s.baseLayer.OnAckTimeout(s.message)
}

func (s *server) sendBlockMessage(messageID uint16, seq, size uint32) error {
	block1Opt, payload, err := s.reader.Read(seq, size)
	if err != nil {
		return err
	}
	s.blockMessageID = messageID
	m := base.Message{
		Type:      base.CON,
		Code:      s.message.Code,
		MessageID: messageID,
		Payload:   payload,
	}
	if !block1Opt.More {
		m.Token = s.message.Token
		m.Options = s.message.Options
	}
	m.SetOption(base.Block1, block1Opt.Value())
	return s.baseLayer.Send(m)
}

func (s *server) isBusy() bool {
	return s.busy
}
