package stack

import (
	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/internal/stack/deduplication"
	"github.com/ironzhang/coap/internal/stack/reliability"
)

// Stack coap协议栈
type Stack struct {
	recver base.Recver
	sender base.Sender
	layers []base.Layer
}

func (s *Stack) Init(recver base.Recver, sender base.Sender, ackTimeout func(base.Message)) *Stack {
	s.recver, s.sender, s.layers = makeLayers(
		recver,
		sender,
		deduplication.NewLayer(),
		reliability.NewLayer(ackTimeout),
	)
	return s
}

func (s *Stack) Recv(m base.Message) error {
	return s.recver.Recv(m)
}

func (s *Stack) Send(m base.Message) error {
	return s.sender.Send(m)
}

func (s *Stack) Update() {
	for _, l := range s.layers {
		l.Update()
	}
}

func makeLayers(recver base.Recver, sender base.Sender, layers ...base.Layer) (base.Recver, base.Sender, []base.Layer) {
	for i := len(layers) - 1; i >= 0; i-- {
		layers[i].SetRecver(recver)
		recver = layers[i]
	}
	for i := 0; i < len(layers); i++ {
		layers[i].SetSender(sender)
		sender = layers[i]
	}
	return recver, sender, layers
}
