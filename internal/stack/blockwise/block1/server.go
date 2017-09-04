package block1

import (
	"bytes"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

type server struct {
	baseLayer *base.BaseLayer
	timeout   time.Duration

	busy      bool
	timestamp time.Time
	buffer    bytes.Buffer
	messageID uint16
	block1    uint32
}

func (s *server) init(baseLayer *base.BaseLayer) {
	s.baseLayer = baseLayer
}

func (s *server) update() {
	if s.busy && time.Since(s.timestamp) > s.timeout {
		s.busy = false
	}
}

func (s *server) recv(m base.Message) error {
	if s.busy {
		return s.busyStateRecv(m)
	}
	return s.idleStateRecv(m)
}

func (s *server) send(m base.Message) error {
	if !s.busy {
		return s.baseLayer.Send(m)
	}
	if m.MessageID != s.messageID {
		return s.baseLayer.Send(m)
	}
	s.busy = false
	m.SetOption(base.Block1, s.block1)
	return s.baseLayer.Send(m)
}

func (s *server) busyStateRecv(m base.Message) error {
	opt, ok := base.ParseBlock1Option(m)
	if !ok {
		return l.baseLayer.NewError(base.ErrNoBlock1Option)
	}
	if s.buffer.Len() != int(opt.Num*opt.Size) {
		return s.ackRequestEntityIncomplete(m.MessageID)
	}
	s.buffer.Write(m.Payload)
	if opt.More {
		return s.ackContinue(m.MessageID, opt.Value())
	}
	s.messageID = m.MessageID
	s.block1 = opt.Value()
	m.Payload = s.copyAndResetBuffer()
	return s.baseLayer.Recv(m)
}

func (s *server) idleStateRecv(m base.Message) error {
	opt, ok := base.ParseBlock1Option(m)
	if !ok {
		return l.baseLayer.Recv(m)
	}

	s.busy = true
	s.timestamp = time.Now()
	s.buffer.Reset()
	s.buffer.Write(m.Payload)
	if opt.More {
		return s.ackContinue(m.MessageID, opt.Value())
	}
	s.messageID = m.MessageID
	s.block1 = opt.Value()
	m.Payload = s.copyAndResetBuffer()
	return s.baseLayer.Recv(m)
}

func (s *server) ackContinue(messageID uint16, block1 uint32) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.Continue,
		MessageID: messageID,
	}
	m.SetOption(base.Block1, block1)
	return s.baseLayer.Send(m)
}

func (s *server) ackRequestEntityIncomplete(messageID uint16) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.RequestEntityIncomplete,
		MessageID: messageID,
	}
	return s.baseLayer.Send(m)
}

func (s *server) copyAndResetBuffer() []byte {
	b := make([]byte, s.buffer.Len())
	copy(b, s.buffer.Bytes())
	s.buffer.Reset()
	return b
}
