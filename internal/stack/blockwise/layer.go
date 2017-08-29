package blockwise

import "github.com/ironzhang/coap/internal/stack/base"

var _ base.Layer = &Layer{}

type Layer struct {
	base.BaseLayer
}

func NewLayer() *Layer {
	return &Layer{
		BaseLayer: base.BaseLayer{Name: "blockwise"},
	}
}

func (l *Layer) Update() {
}

func (l *Layer) Recv(m base.Message) error {
	c := m.Code >> 5
	switch {
	case c == 0:
		// 请求
	case c >= 2 && c <= 5:
		// 响应
	default:
	}
	return nil
}

func (l *Layer) Send(m base.Message) error {
	return nil
}

func (l *Layer) recvRequest(m base.Message) error {
	//	opt := m.GetOption(base.Block1)
	//	if opt == nil {
	//		return l.BaseLayer.Recv(m)
	//	}
	return nil
}

func (l *Layer) sendACK(messageID uint16) error {
	m := base.Message{
		Type:      base.ACK,
		Code:      base.Continue,
		MessageID: messageID,
	}
	return l.BaseLayer.Send(m)
}
