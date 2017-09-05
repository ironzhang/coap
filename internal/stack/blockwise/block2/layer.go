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
	l.BaseLayer.Name = "block2"
	l.client.init(&l.BaseLayer, generator)
	l.server.init(&l.BaseLayer, base.MAX_BLOCKSIZE, base.EXCHANGE_LIFETIME)
	return l
}

func (l *Layer) Update() {
	l.server.Update()
}

func (l *Layer) OnAckTimeout(m base.Message) {
	l.client.OnAckTimeout(m)
}

func (l *Layer) Recv(m base.Message) error {
	switch m.Type {
	case base.CON:
		return l.server.Recv(m)
	case base.ACK:
		return l.client.Recv(m)
	default:
		return l.BaseLayer.Recv(m)
	}
}

func (l *Layer) Send(m base.Message) error {
	switch m.Type {
	case base.CON:
		return l.client.Send(m)
	case base.ACK:
		return l.server.Send(m)
	default:
		return l.BaseLayer.Send(m)
	}
}
