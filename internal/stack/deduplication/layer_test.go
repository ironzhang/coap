package deduplication

import (
	"testing"
	"time"

	"github.com/ironzhang/coap/internal/message"
	"github.com/ironzhang/coap/internal/stack/base"
)

func TestStateTimeout(t *testing.T) {
	tests := []struct {
		time    time.Time
		timeout time.Duration
		result  bool
	}{
		{
			time:    time.Now().Add(-1 * time.Second),
			timeout: 2 * time.Second,
			result:  false,
		},
		{
			time:    time.Now().Add(-3 * time.Second),
			timeout: 2 * time.Second,
			result:  true,
		},
	}
	for i, tt := range tests {
		s := state{Time: tt.time}
		if got, want := s.Timeout(tt.timeout), tt.result; got != want {
			t.Errorf("case%d: got(%v) != want(%v)", i, got, want)
		}
	}
}

func TestLayerTimeout(t *testing.T) {
	l := NewLayer()
	tests := []struct {
		state  state
		result bool
	}{
		{
			state:  state{Time: time.Now().Add(-l.ExchangeLifetime + 1*time.Second), Type: message.CON},
			result: false,
		},
		{
			state:  state{Time: time.Now().Add(-l.ExchangeLifetime - 1*time.Second), Type: message.CON},
			result: true,
		},
		{
			state:  state{Time: time.Now().Add(-l.NonLifetime + 1*time.Second), Type: message.NON},
			result: false,
		},
		{
			state:  state{Time: time.Now().Add(-l.NonLifetime - 1*time.Second), Type: message.NON},
			result: true,
		},
	}
	for i, tt := range tests {
		if got, want := l.timeout(&tt.state), tt.result; got != want {
			t.Errorf("case%d: got(%v) != want(%v)", i, got, want)
		}
	}
}

func TestLayerRecvCON(t *testing.T) {
	r := base.CountRecver{}
	l := NewLayer()
	l.BaseLayer.Recver = &r
	l.ExchangeLifetime = time.Second

	n := 100
	m := message.Message{
		Type:      message.CON,
		Code:      message.GET,
		MessageID: 1,
	}

	for i := 0; i < n; i++ {
		l.Recv(m)
	}
	if got, want := r.Count, 1; got != want {
		t.Fatalf("got(%d) != want(%d)", got, want)
	}
	time.Sleep(l.ExchangeLifetime)
	for i := 0; i < n; i++ {
		l.Recv(m)
	}
	if got, want := r.Count, 2; got != want {
		t.Fatalf("sleep: got(%d) != want(%d)", got, want)
	}
}
