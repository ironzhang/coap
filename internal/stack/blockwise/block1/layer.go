package block1

import "github.com/ironzhang/coap/internal/stack/base"

var _ base.Layer = &Layer{}

type Layer struct {
	base.BaseLayer
	server server
	client client
}

func NewLayer(generator func() uint16) *Layer {
	return new(Layer).init(generator)
}

func (l *Layer) init(generator func() uint16) *Layer {
	l.BaseLayer.Name = "block1"
	l.server.init(&l.BaseLayer)
	l.client.init(&l.BaseLayer, generator, base.MAX_BLOCKSIZE, base.EXCHANGE_LIFETIME)
	return l
}

func (l *Layer) Update() {
	l.client.update()
}

func (l *Layer) OnAckTimeout(m base.Message) {
	l.client.onAckTimeout(m)
}

func (l *Layer) Recv(m base.Message) error {
	switch {
	case m.Type == base.CON:
		return l.server.recv(m)
	case m.Type == base.ACK:
		return l.client.recv(m)
	default:
		return l.BaseLayer.Recv(m)
	}
}

func (l *Layer) Send(m base.Message) error {
	switch {
	case m.Type == base.CON:
		return l.client.send(m)
	case m.Type == base.ACK:
		return l.server.send(m)
	default:
		return l.BaseLayer.Send(m)
	}
}
