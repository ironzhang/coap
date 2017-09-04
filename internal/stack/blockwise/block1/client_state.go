package block1

import (
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

type clientOptions struct {
	Timeout   time.Duration
	BlockSize uint32
}

type normalTransferClientState struct {
	base    *base.BaseLayer
	machine *machine
	options *clientOptions
}

func newNormalTransferClientState(b *base.BaseLayer, m *machine, o *clientOptions) *normalTransferClientState {
	return &normalTransferClientState{
		base:    b,
		machine: m,
		options: o,
	}
}

func (s *normalTransferClientState) Name() string {
	return "normalTransferClientState"
}

func (s *normalTransferClientState) OnStart() {
}

func (s *normalTransferClientState) OnFinish() {
}

func (s *normalTransferClientState) Update() {
}

func (s *normalTransferClientState) OnAckTimeout(m base.Message) {
	s.base.OnAckTimeout(m)
}

func (s *normalTransferClientState) Recv(m base.Message) error {
	return s.base.Recv(m)
}

func (s *normalTransferClientState) Send(m base.Message) error {
	if len(m.Payload) > int(s.options.BlockSize) {
		s.machine.SetState("blockTransferClientState")
		return s.machine.Send(m)
	}
	return s.base.Send(m)
}

type blockTransferClientState struct {
	base      *base.BaseLayer
	machine   *machine
	options   *clientOptions
	generator func() uint16

	start    time.Time
	message  base.Message
	buffer   base.BlockBuffer
	blockMID uint16
}

func newBlockTransferClientState(b *base.BaseLayer, m *machine, o *clientOptions, f func() uint16) *blockTransferClientState {
	return &blockTransferClientState{
		base:      b,
		machine:   m,
		options:   o,
		generator: f,
	}
}

func (s *blockTransferClientState) Name() string {
	return "blockTransferClientState"
}

func (s *blockTransferClientState) OnStart() {
	s.start = time.Now()
	s.buffer = nil
}

func (s *blockTransferClientState) OnFinish() {
}

func (s *blockTransferClientState) Update() {
	if time.Since(s.start) > s.options.Timeout {
		s.machine.SetState("normalTransferClientState")
		s.base.OnAckTimeout(s.message)
	}
}

func (s *blockTransferClientState) OnAckTimeout(m base.Message) {
	if m.MessageID != s.blockMID {
		s.base.OnAckTimeout(m)
	} else {
		s.base.OnAckTimeout(s.message)
	}
}

func (s *blockTransferClientState) Recv(m base.Message) error {
	if m.Code == base.RequestEntityIncomplete || m.Code == base.RequestEntityTooLarge {
		s.machine.SetState("normalTransferClientState")
		m.MessageID = s.message.MessageID
		return s.base.Recv(m)
	}

	opt, ok := base.ParseBlock1Option(m)
	if !ok {
		return s.base.NewError(base.ErrNoBlock1Option)
	}
	s.options.BlockSize = opt.Size
	if opt.More {
		return s.sendBlockMessage(s.generator(), opt.Num+1, opt.Size)
	}
	s.machine.SetState("normalTransferClientState")
	m.MessageID = s.message.MessageID
	return s.base.Recv(m)
}

func (s *blockTransferClientState) Send(m base.Message) error {
	if len(s.buffer) > 0 {
		return s.base.NewError(base.ErrClientBusy)
	}
	s.message = m
	s.buffer = m.Payload
	return s.sendBlockMessage(m.MessageID, 0, s.options.BlockSize)
}

func (s *blockTransferClientState) sendBlockMessage(messageID uint16, num, size uint32) error {
	opt, payload, err := s.buffer.Read(num, size)
	if err != nil {
		return err
	}
	s.blockMID = messageID
	m := base.Message{
		Type:      base.CON,
		Code:      s.message.Code,
		MessageID: messageID,
		Payload:   payload,
	}
	if !opt.More {
		m.Token = s.message.Token
		m.Options = s.message.Options
	}
	m.SetOption(base.Block1, opt.Value())
	return s.base.Send(m)
}
