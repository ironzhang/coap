package reliability

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

func TestRandAckTimeout(t *testing.T) {
	seed := time.Now().Unix()
	rand.Seed(seed)

	l := NewLayer(nil)
	min := l.AckTimeout
	max := time.Duration(float64(l.AckTimeout) * l.AckRandomFactor)
	for i := 0; i < 10000; i++ {
		d := l.randAckTimeout()
		if d < min || d > max {
			t.Fatalf("%d: seed=%d, d=%s, min=%s, max=%s", i, seed, d, min, max)
		}
		//fmt.Printf("%d: %v\n", i, d)
	}
}

func TestRecvAck(t *testing.T) {
	r := base.CountRecver{}
	s := base.CountSender{}
	l := NewLayer(nil)
	l.AckTimeout = 10 * time.Millisecond
	l.BaseLayer.Recver = &r
	l.BaseLayer.Sender = &s

	m := base.Message{Type: base.CON, Code: base.GET, MessageID: 1}
	if err := l.Send(m); err != nil {
		t.Fatalf("send: %v", err)
	}
	for r.Count <= 0 {
		time.Sleep(2 * l.AckTimeout)
		l.Update()
		l.Recv(base.Message{Type: base.ACK, Code: base.Content, MessageID: 1})
	}
	time.Sleep(2 * l.AckTimeout)
	l.Update()
	if got, want := s.Count, 2; got != want {
		t.Errorf("Retransmit: %d != %d", got, want)
	}
}

type Timeout struct {
	timeout bool
}

func (t *Timeout) Timeout(m base.Message) {
	t.timeout = true
}

func TestAckTimeout(t *testing.T) {
	var timeout Timeout
	r := base.CountRecver{}
	s := base.CountSender{}
	l := NewLayer(timeout.Timeout)
	l.AckTimeout = 10 * time.Millisecond
	l.BaseLayer.Recver = &r
	l.BaseLayer.Sender = &s

	m := base.Message{Type: base.CON, Code: base.GET, MessageID: 1}
	if err := l.Send(m); err != nil {
		t.Fatalf("send: %v", err)
	}
	for !timeout.timeout {
		time.Sleep(l.AckTimeout)
		l.Update()
	}
	if got, want := s.Count, l.MaxRetransmit; got != want {
		t.Errorf("Retransmit: %d != %d", got, want)
	}
}
