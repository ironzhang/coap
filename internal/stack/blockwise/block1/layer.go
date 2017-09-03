package block1

import "github.com/ironzhang/coap/internal/stack/base"

const MaxBlockSize = 1024

var _ base.Layer = &Layer{}

type Layer struct {
	base.BaseLayer
	client client
	server server
}

func NewLayer(generator MessageIDGenerator) *Layer {
	return new(Layer).init(generator)
}

func (l *Layer) init(generator MessageIDGenerator) *Layer {
	l.BaseLayer.Name = "block1"
	l.client.init(&l.BaseLayer)
	l.server.init(&l.BaseLayer, generator, MaxBlockSize)
	return l
}

func (l *Layer) Update() {
}

func (l *Layer) OnAckTimeout(m base.Message) {
	l.server.onAckTimeout(m)
}

func (l *Layer) Recv(m base.Message) error {
	switch {
	case m.Type == base.CON:
		return l.client.recv(m)
	case m.Type == base.ACK:
		return l.server.recv(m)
	default:
		return l.BaseLayer.Recv(m)
	}
}

func (l *Layer) Send(m base.Message) error {
	switch {
	case m.Type == base.CON:
		return l.server.send(m)
	case m.Type == base.ACK:
		return l.client.send(m)
	default:
		return l.BaseLayer.Send(m)
	}
}

func isConRequest(m base.Message) bool {
	if m.Type != base.CON {
		return false
	}
	return (m.Code >> 5) == 0
}
