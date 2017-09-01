package block1

import "github.com/ironzhang/coap/internal/stack/base"

const MaxBlockSize = 1024

var _ base.Layer = &Layer{}

type Layer struct {
	base.BaseLayer
	receiver    receiver
	transmitter transmitter
}

func NewLayer(generator MessageIDGenerator) *Layer {
	return new(Layer).init(generator)
}

func (l *Layer) init(generator MessageIDGenerator) *Layer {
	l.BaseLayer.Name = "block1"
	l.receiver.init(&l.BaseLayer)
	l.transmitter.init(&l.BaseLayer, generator, MaxBlockSize)
	return l
}

func (l *Layer) Update() {
}

func (l *Layer) Recv(m base.Message) error {
	switch {
	case isConRequest(m):
		return l.receiver.recv(m)
	case m.Type == base.ACK:
		return l.transmitter.recv(m)
	default:
		return l.BaseLayer.Recv(m)
	}
}

func (l *Layer) Send(m base.Message) error {
	switch {
	case isConRequest(m):
		return l.transmitter.send(m)
	case m.Type == base.ACK:
		return l.receiver.send(m)
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
