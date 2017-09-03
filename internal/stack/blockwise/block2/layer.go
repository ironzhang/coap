package block2

import "github.com/ironzhang/coap/internal/stack/base"

type Layer struct {
	base.BaseLayer
	client client
	server server
}

func NewLayer(generator func() uint16) *Layer {
	return new(Layer).init(generator)
}

func (l *Layer) init(generator func() uint16) *Layer {
	l.BaseLayer.Name = "block1"
	l.client.init(&l.BaseLayer, generator)
	l.server.init(&l.BaseLayer, 1024)
	return l
}

func (l *Layer) Update() {
}

func (l *Layer) Recv(m base.Message) error {
	switch m.Type {
	case base.CON:
		return l.server.recv(m)
	case base.ACK:
		return l.client.recv(m)
	default:
		return l.BaseLayer.Recv(m)
	}
}

func (l *Layer) Send(m base.Message) error {
	switch m.Type {
	case base.CON:
		return l.client.send(m)
	case base.ACK:
		return l.server.send(m)
	default:
		return l.BaseLayer.Send(m)
	}
}
