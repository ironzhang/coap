package deduplication

import (
	"testing"
	"time"

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
			state:  state{Time: time.Now().Add(-l.ExchangeLifetime + 1*time.Second), Type: base.CON},
			result: false,
		},
		{
			state:  state{Time: time.Now().Add(-l.ExchangeLifetime - 1*time.Second), Type: base.CON},
			result: true,
		},
		{
			state:  state{Time: time.Now().Add(-l.NonLifetime + 1*time.Second), Type: base.NON},
			result: false,
		},
		{
			state:  state{Time: time.Now().Add(-l.NonLifetime - 1*time.Second), Type: base.NON},
			result: true,
		},
	}
	for i, tt := range tests {
		if got, want := l.timeout(&tt.state), tt.result; got != want {
			t.Errorf("case%d: got(%v) != want(%v)", i, got, want)
		}
	}
}

func TestLayerUpdate(t *testing.T) {
	r := base.CountRecver{}
	l := NewLayer()
	l.BaseLayer.Recver = &r
	l.NonLifetime = 100 * time.Millisecond
	l.ExchangeLifetime = 200 * time.Millisecond

	prepare := func(non, con, ack, rst int) {
		count := non + con + ack + rst
		for i := 0; i < count; i++ {
			m := base.Message{MessageID: uint16(i)}
			if i < non {
				m.Type = base.NON
			} else if i < non+con {
				m.Type = base.CON
			} else if i < non+con+ack {
				m.Type = base.ACK
			} else {
				m.Type = base.RST
			}
			l.Recv(m)
		}
	}

	non, con, ack, rst := 10, 20, 30, 40
	prepare(non, con, ack, rst)
	if got, want := r.Count, non+con+ack+rst; got != want {
		t.Fatalf("Recver.Count: %d != %d", got, want)
	}
	if got, want := len(l.states), non+con; got != want {
		t.Fatalf("Layer.states: %d != %d", got, want)
	}

	l.Update()
	if got, want := len(l.states), non+con; got != want {
		t.Fatalf("Layer.states: %d != %d", got, want)
	}

	time.Sleep(l.NonLifetime)
	l.Update()
	if got, want := len(l.states), con; got != want {
		t.Fatalf("Layer.states: %d != %d", got, want)
	}

	time.Sleep(l.ExchangeLifetime)
	l.Update()
	if got, want := len(l.states), 0; got != want {
		t.Fatalf("Layer.states: %d != %d", got, want)
	}
}

func TestRecvTimeout(t *testing.T) {
	r := base.CountRecver{}
	l := NewLayer()
	l.BaseLayer.Recver = &r
	l.NonLifetime = 100 * time.Millisecond
	l.ExchangeLifetime = 200 * time.Millisecond

	n := 100
	tests := []struct {
		mesg  base.Message
		sleep time.Duration
	}{
		{
			mesg: base.Message{
				Type:      base.CON,
				Code:      base.GET,
				MessageID: 1,
			},
			sleep: l.ExchangeLifetime,
		},
		{
			mesg: base.Message{
				Type:      base.NON,
				Code:      base.GET,
				MessageID: 2,
			},
			sleep: l.NonLifetime,
		},
	}
	for i, tt := range tests {
		for i := 0; i < n; i++ {
			l.Recv(tt.mesg)
		}
		if got, want := r.Count, 1; got != want {
			t.Fatalf("case%d: got(%d) != want(%d)", i, got, want)
		}
		time.Sleep(tt.sleep)
		for i := 0; i < n; i++ {
			l.Recv(tt.mesg)
		}
		if got, want := r.Count, 2; got != want {
			t.Fatalf("case%d: sleep: got(%d) != want(%d)", i, got, want)
		}
		r.Count = 0
	}
}

func TestRecvMessage(t *testing.T) {
	r := base.CountRecver{}
	s := base.CountSender{}
	l := NewLayer()
	l.BaseLayer.Recver = &r
	l.BaseLayer.Sender = &s

	tests := []struct {
		mesgs     []base.Message
		recvCount int
		sendCount int
	}{
		{
			mesgs:     []base.Message{},
			recvCount: 0,
			sendCount: 0,
		},
		{
			mesgs: []base.Message{
				{Type: base.CON, MessageID: 10},
				{Type: base.CON, MessageID: 11},
				{Type: base.CON, MessageID: 10},
				{Type: base.CON, MessageID: 11},
			},
			recvCount: 2,
			sendCount: 0,
		},
		{
			mesgs: []base.Message{
				{Type: base.CON, MessageID: 20},
				{Type: base.CON, MessageID: 21},
				{Type: base.NON, MessageID: 20},
				{Type: base.NON, MessageID: 21},
			},
			recvCount: 2,
			sendCount: 0,
		},
		{
			mesgs: []base.Message{
				{Type: base.NON, MessageID: 30},
				{Type: base.NON, MessageID: 31},
				{Type: base.NON, MessageID: 30},
				{Type: base.NON, MessageID: 31},
			},
			recvCount: 2,
			sendCount: 0,
		},
		{
			mesgs: []base.Message{
				{Type: base.NON, MessageID: 40},
				{Type: base.NON, MessageID: 41},
				{Type: base.CON, MessageID: 40},
				{Type: base.CON, MessageID: 41},
			},
			recvCount: 2,
			sendCount: 2,
		},
	}
	for i, tt := range tests {
		for _, m := range tt.mesgs {
			if err := l.Recv(m); err != nil {
				t.Errorf("case%d: recv: %v", i, err)
			}
		}
		if got, want := r.Count, tt.recvCount; got != want {
			t.Fatalf("case%d: Recver.Count: got(%d) != want(%d)", i, got, want)
		}
		if got, want := s.Count, tt.sendCount; got != want {
			t.Fatalf("case%d: Sender.Count: got(%d) != want(%d)", i, got, want)
		}
		r.Count = 0
		s.Count = 0
	}
}

func TestRecvSend(t *testing.T) {
	r := base.CountRecver{}
	s := base.CountSender{}
	l := NewLayer()
	l.BaseLayer.Recver = &r
	l.BaseLayer.Sender = &s

	tests := []struct {
		recv      base.Message
		send      base.Message
		recvCount int
		sendCount int
	}{
		{
			recv: base.Message{
				Type:      base.CON,
				Code:      base.GET,
				MessageID: 1,
			},
			send: base.Message{
				Type:      base.ACK,
				Code:      base.Content,
				MessageID: 1,
			},
			recvCount: 1,
			sendCount: 2,
		},
		{
			recv: base.Message{
				Type:      base.CON,
				Code:      base.GET,
				MessageID: 2,
			},
			send: base.Message{
				Type:      base.RST,
				Code:      base.Content,
				MessageID: 2,
			},
			recvCount: 1,
			sendCount: 2,
		},
		{
			recv: base.Message{
				Type:      base.CON,
				Code:      base.GET,
				MessageID: 3,
			},
			send: base.Message{
				Type:      base.CON,
				Code:      base.Content,
				MessageID: 3,
			},
			recvCount: 1,
			sendCount: 1,
		},
		{
			recv: base.Message{
				Type:      base.NON,
				Code:      base.GET,
				MessageID: 4,
			},
			send: base.Message{
				Type:      base.RST,
				Code:      base.Content,
				MessageID: 4,
			},
			recvCount: 1,
			sendCount: 1,
		},
	}
	for i, tt := range tests {
		if err := l.Recv(tt.recv); err != nil {
			t.Fatalf("case%d: recv: %v", i, err)
		}
		if err := l.Send(tt.send); err != nil {
			t.Fatalf("case%d: send: %v", i, err)
		}
		if got, want := r.Count, 1; got != want {
			t.Fatalf("case%d: Recver.Count: got(%d) != want(%d)", i, got, want)
		}
		if got, want := s.Count, 1; got != want {
			t.Fatalf("case%d: Sender.Count: got(%d) != want(%d)", i, got, want)
		}
		if err := l.Recv(tt.recv); err != nil {
			t.Fatalf("case%d: recv: %v", i, err)
		}
		if got, want := r.Count, tt.recvCount; got != want {
			t.Fatalf("case%d: Recver.Count: got(%d) != want(%d)", i, got, want)
		}
		if got, want := s.Count, tt.sendCount; got != want {
			t.Fatalf("case%d: Sender.Count: got(%d) != want(%d)", i, got, want)
		}
		r.Count = 0
		s.Count = 0
	}
}
