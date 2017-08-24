package stack

import (
	"github.com/ironzhang/coap/internal/message"
	"github.com/ironzhang/coap/internal/stack/deduplication"
	"github.com/ironzhang/coap/internal/stack/layer"
	"github.com/ironzhang/coap/internal/stack/reliability"
)

type Stack struct {
	layer.Recver
	layer.Sender
	layers []layer.Layer
}

func NewStack(recver layer.Recver, sender layer.Sender, timeout func(message.Message)) *Stack {
	return new(Stack).Init(recver, sender, timeout)
}

func (s *Stack) Init(recver layer.Recver, sender layer.Sender, timeout func(message.Message)) *Stack {
	s.Recver, s.Sender, s.layers = makeLayers(
		recver, sender,
		deduplication.NewLayer(),
		reliability.NewLayer(timeout),
	)
	return s
}

func (s *Stack) Update() {
	for _, l := range s.layers {
		if u, ok := l.(layer.Updater); ok {
			u.Update()
		}
	}
}

func makeLayers(recver layer.Recver, sender layer.Sender, layers ...layer.Layer) (layer.Recver, layer.Sender, []layer.Layer) {
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
