package reliability

import (
	"github.com/ironzhang/coap/internal/message"
	"github.com/ironzhang/coap/internal/stack/layer"
)

var _ layer.Layer = &Layer{}

type Layer struct {
	layer.BaseLayer
}

func (l *Layer) Update() {
}

func (l *Layer) Recv(m message.Message) error {
	if m.Type != message.ACK && m.Type != message.RST {
		return l.BaseLayer.Recv(m)
	}

	return nil
}

func (l *Layer) Send(m message.Message) error {
	return nil
}
