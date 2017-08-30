package blockwise

import "github.com/ironzhang/coap/internal/stack/base"

var _ base.Layer = &Layer{}

type state struct {
	block1 uint32
}

type Layer struct {
	base.BaseLayer
	upload buffer
	states map[uint16]state
}

func NewLayer() *Layer {
	return &Layer{
		BaseLayer: base.BaseLayer{Name: "blockwise"},
		states:    make(map[uint16]state),
	}
}

func (l *Layer) Update() {
}

func (l *Layer) Recv(m base.Message) error {
	switch m.Type {
	case base.CON:
		return l.recvCON(m)
	case base.NON:
		l.BaseLayer.Recv(m)
	case base.ACK:
	case base.RST:
	}
	return nil
}

func (l *Layer) Send(m base.Message) error {
	switch m.Type {
	case base.CON:
	case base.NON:
	case base.ACK:
	case base.RST:
	}
	return nil
}

func (l *Layer) recvCON(m base.Message) error {
	c := m.Code >> 5
	switch {
	case c == 0:
		// 请求
		return l.recvRequest(m)
	case c >= 2 && c <= 5:
		// 响应
	default:
	}
	return nil
}

func (l *Layer) recvRequest(m base.Message) error {
	block1, ok := getBlockOption(m, base.Block1)
	if ok {
		return l.recvBlockRequest(block1, m)
	}
	return l.BaseLayer.Recv(m)
}

func (l *Layer) sendRequest(m base.Message) error {
	//	size1, ok := getBlockSize(m, base.Size1)
	//	if ok {
	//		return l.sendBlockRequest(size1, m)
	//	}
	//	if len(m.Payload) > l.MaxBlockSize {
	//		return l.sendBlockRequest(l.MaxBlockSize, m)
	//	}
	return nil
}

func (l *Layer) sendACK(m base.Message) error {
	if s, ok := l.getState(m.MessageID); ok {
		delete(l.states, m.MessageID)
		m.Options = append(m.Options, base.Option{ID: base.Block1, Value: s.block1})
	}
	return l.BaseLayer.Send(m)
}

func (l *Layer) recvBlockRequest(o blockOption, m base.Message) error {
	b := &block{
		more:    o.more,
		seq:     o.num,
		payload: m.Payload,
	}
	if err := l.upload.WriteBlock(b); err != nil {
		return err
	}
	if o.more {
		return l.ack(m.MessageID, o.Value())
	}
	m.Payload = l.upload.ReadPayload()
	l.states[m.MessageID] = state{block1: o.Value()}
	return l.BaseLayer.Recv(m)
}

func (l *Layer) ack(messageID uint16, block1 uint32) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.Continue,
		MessageID: messageID,
		Options: []base.Option{
			{ID: base.Block1, Value: block1},
		},
	}
	return l.BaseLayer.Send(m)
}

func (l *Layer) getState(id uint16) (state, bool) {
	s, ok := l.states[id]
	return s, ok
}

func getBlockOption(m base.Message, id uint16) (blockOption, bool) {
	block := m.GetOption(id)
	if block == nil {
		return blockOption{}, false
	}
	return parseBlockOption(block.(uint32)), true
}
