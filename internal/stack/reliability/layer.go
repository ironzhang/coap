package reliability

import (
	"errors"
	"math/rand"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

var (
	ErrDupMessageID = errors.New("message id duplicate")
)

// state 消息状态
type state struct {
	Start          time.Time
	Message        base.Message
	LastRetransmit time.Time
	Retransmit     int
	Timeout        time.Duration
}

var _ base.Layer = &Layer{}

type Layer struct {
	base.BaseLayer
	MaxRetransmit   int
	MaxTransmitSpan time.Duration
	MaxTransmitWait time.Duration
	AckTimeout      time.Duration
	AckRandomFactor float64

	states map[uint16]*state
}

func NewLayer() *Layer {
	return &Layer{
		BaseLayer:       base.BaseLayer{Name: "reliability"},
		MaxRetransmit:   base.MAX_RETRANSMIT,
		MaxTransmitSpan: base.MAX_TRANSMIT_SPAN,
		MaxTransmitWait: base.MAX_TRANSMIT_WAIT,
		AckTimeout:      base.ACK_TIMEOUT,
		AckRandomFactor: base.ACK_RANDOM_FACTOR,
		states:          make(map[uint16]*state),
	}
}

func (l *Layer) Update() {
	for _, s := range l.states {
		if s.LastRetransmit.Sub(s.Start) >= l.MaxTransmitSpan {
			l.doTimeout(s)
			continue
		}

		if time.Since(s.Start) >= l.MaxTransmitWait {
			l.doTimeout(s)
			continue
		}

		if time.Since(s.LastRetransmit) >= s.Timeout {
			if s.Retransmit >= l.MaxRetransmit {
				l.doTimeout(s)
				continue
			}
			l.send(s)
		}
	}
}

func (l *Layer) Recv(m base.Message) error {
	if m.Type != base.ACK && m.Type != base.RST {
		return l.BaseLayer.Recv(m)
	}
	if l.delState(m) {
		return l.BaseLayer.Recv(m)
	}
	return nil
}

func (l *Layer) Send(m base.Message) error {
	if m.Type != base.CON {
		return l.BaseLayer.Send(m)
	}
	s, ok := l.addState(m)
	if !ok {
		return l.BaseLayer.NewError(ErrDupMessageID)
	}
	return l.send(s)
}

func (l *Layer) send(s *state) error {
	s.LastRetransmit = time.Now()
	if s.Retransmit == 0 {
		s.Timeout = l.randAckTimeout()
	} else {
		s.Timeout *= 2
	}
	s.Retransmit++
	return l.BaseLayer.Send(s.Message)
}

func (l *Layer) doTimeout(s *state) {
	delete(l.states, s.Message.MessageID)
	l.BaseLayer.OnAckTimeout(s.Message)
}

func (l *Layer) addState(m base.Message) (*state, bool) {
	if _, ok := l.states[m.MessageID]; ok {
		return nil, false
	}
	s := &state{Start: time.Now(), Message: m}
	l.states[m.MessageID] = s
	return s, true
}

func (l *Layer) delState(m base.Message) bool {
	if _, ok := l.states[m.MessageID]; ok {
		delete(l.states, m.MessageID)
		return true
	}
	return false
}

func init() {
	rand.Seed(time.Now().Unix())
}

func (l *Layer) randAckTimeout() time.Duration {
	factor := l.AckRandomFactor - 1
	if factor < 0 {
		factor = 0
	}
	return l.AckTimeout + time.Duration(rand.Float64()*factor*float64(l.AckTimeout))
}
