package stack

import (
	"github.com/ironzhang/coap/internal/stack/deduplication"
	"github.com/ironzhang/coap/internal/stack/layer"
)

type Stack struct {
	layer.Recver
	layer.Sender
}

func NewStack(recver layer.Recver, sender layer.Sender) *Stack {
	return new(Stack).Init(recver, sender)
}

func (s *Stack) Init(recver layer.Recver, sender layer.Sender) *Stack {
	s.Recver, s.Sender = makeLayers(recver, sender, deduplication.NewLayer())
	return s
}

func makeLayers(recver layer.Recver, sender layer.Sender, layers ...layer.Layer) (layer.Recver, layer.Sender) {
	for i := len(layers) - 1; i >= 0; i-- {
		layers[i].SetRecver(recver)
		recver = layers[i]
	}
	for i := 0; i < len(layers); i++ {
		layers[i].SetSender(sender)
		sender = layers[i]
	}
	return recver, sender
}
