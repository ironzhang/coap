package block1

import "github.com/ironzhang/coap/internal/stack/base"

var _ base.Layer = &Layer{}

type Layer struct {
	base.BaseLayer
	clientOptions clientOptions
	client        machine
	serverOptions serverOptions
	server        machine
}

func NewLayer(generator func() uint16) *Layer {
	return new(Layer).init(generator)
}

func (l *Layer) init(generator func() uint16) *Layer {
	l.BaseLayer.Name = "block1"
	l.clientOptions.Timeout = base.EXCHANGE_LIFETIME
	l.clientOptions.BlockSize = base.MAX_BLOCKSIZE
	l.serverOptions.Timeout = base.EXCHANGE_LIFETIME
	l.client.Init(
		newNormalTransferClientState(&l.BaseLayer, &l.client, &l.clientOptions),
		newBlockTransferClientState(&l.BaseLayer, &l.client, &l.clientOptions, generator),
	)
	l.server.Init(
		newNormalTransferServerState(&l.BaseLayer, &l.server),
		newBlockTransferServerState(&l.BaseLayer, &l.server, &l.serverOptions),
	)
	return l
}

func (l *Layer) Update() {
	l.client.Update()
	l.server.Update()
}

func (l *Layer) OnAckTimeout(m base.Message) {
	l.client.OnAckTimeout(m)
}

func (l *Layer) Recv(m base.Message) error {
	switch {
	case m.Type == base.CON:
		return l.server.Recv(m)
	case m.Type == base.ACK:
		return l.client.Recv(m)
	default:
		return l.BaseLayer.Recv(m)
	}
}

func (l *Layer) Send(m base.Message) error {
	switch {
	case m.Type == base.CON:
		return l.client.Send(m)
	case m.Type == base.ACK:
		return l.server.Send(m)
	default:
		return l.BaseLayer.Send(m)
	}
}
