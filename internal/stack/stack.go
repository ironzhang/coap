package stack

import (
	"github.com/ironzhang/coap/internal/message"
	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/internal/stack/deduplication"
	"github.com/ironzhang/coap/internal/stack/reliability"
)

type Stack struct {
	base.Recver
	base.Sender
	layers []base.Layer
}

func NewStack(recver base.Recver, sender base.Sender, timeout func(message.Message)) *Stack {
	return new(Stack).Init(recver, sender, timeout)
}

func (s *Stack) Init(recver base.Recver, sender base.Sender, timeout func(message.Message)) *Stack {
	s.Recver, s.Sender, s.layers = makeLayers(
		recver, sender,
		deduplication.NewLayer(),
		reliability.NewLayer(timeout),
	)
	return s
}

func (s *Stack) Update() {
	for _, l := range s.layers {
		if u, ok := l.(base.Updater); ok {
			u.Update()
		}
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
