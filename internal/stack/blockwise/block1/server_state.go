package block1

import (
	"bytes"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

type serverOptions struct {
	Timeout time.Duration
}

type normalTransferServerState struct {
	base    *base.BaseLayer
	machine *machine
}

func newNormalTransferServerState(b *base.BaseLayer, m *machine) *normalTransferServerState {
	return &normalTransferServerState{
		base:    b,
		machine: m,
	}
}

func (s *normalTransferServerState) Name() string {
	return "normalTransferServerState"
}

func (s *normalTransferServerState) OnStart() {
}

func (s *normalTransferServerState) OnFinish() {
}

func (s *normalTransferServerState) Update() {
}

func (s *normalTransferServerState) OnAckTimeout(m base.Message) {
}

func (s *normalTransferServerState) Recv(m base.Message) error {
	if _, ok := base.ParseBlock1Option(m); ok {
		s.machine.SetState("blockTransferServerState")
		return s.machine.Recv(m)
	}
	return s.base.Recv(m)
}

func (s *normalTransferServerState) Send(m base.Message) error {
	return s.base.Send(m)
}

type blockTransferServerState struct {
	base    *base.BaseLayer
	machine *machine
	options *serverOptions

	start     time.Time
	buffer    bytes.Buffer
	waitACK   bool
	block1    uint32
	messageID uint16
}

func newBlockTransferServerState(b *base.BaseLayer, m *machine, o *serverOptions) *blockTransferServerState {
	return &blockTransferServerState{
		base:    b,
		machine: m,
		options: o,
	}
}

func (s *blockTransferServerState) Name() string {
	return "blockTransferServerState"
}

func (s *blockTransferServerState) OnStart() {
	s.start = time.Now()
	s.buffer.Reset()
	s.waitACK = false
}

func (s *blockTransferServerState) OnFinish() {
}

func (s *blockTransferServerState) Update() {
	if time.Since(s.start) > s.options.Timeout {
		s.machine.SetState("normalTransferServerState")
	}
}

func (s *blockTransferServerState) OnAckTimeout(m base.Message) {
}

func (s *blockTransferServerState) Recv(m base.Message) error {
	opt, ok := base.ParseBlock1Option(m)
	if !ok {
		return s.base.NewError(base.ErrNoBlock1Option)
	}
	if s.buffer.Len() != int(opt.Num*opt.Size) {
		return s.ackIncomplete(m.MessageID)
	}
	s.buffer.Write(m.Payload)
	if opt.More {
		return s.ackContinue(m.MessageID, opt)
	}

	s.waitACK = true
	s.block1 = opt.Value()
	s.messageID = m.MessageID
	m.Payload = s.buffer.Bytes()
	return s.base.Recv(m)
}

func (s *blockTransferServerState) Send(m base.Message) error {
	if !s.waitACK {
		return s.base.Send(m)
	}
	if s.messageID != m.MessageID {
		return s.base.Send(m)
	}
	s.machine.SetState("normalTransferServerState")
	m.SetOption(base.Block1, s.block1)
	return s.base.Send(m)
}

func (s *blockTransferServerState) ackIncomplete(messageID uint16) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.RequestEntityIncomplete,
		MessageID: messageID,
	}
	return s.base.Send(m)
}

func (s *blockTransferServerState) ackContinue(messageID uint16, opt base.BlockOption) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.Continue,
		MessageID: messageID,
	}
	m.SetOption(base.Block1, opt.Value())
	return s.base.Send(m)
}
