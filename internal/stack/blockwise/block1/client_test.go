package block1

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/ironzhang/coap/internal/stack/base"
)

func TestCStatus(t *testing.T) {
	var status cstatus
	messages := []base.Message{
		{MessageID: 0},
		{MessageID: 1},
		{MessageID: 2},
	}
	for _, m := range messages {
		if _, err := status.add(m); err != nil {
			t.Fatalf("add message(%d): %v", m.MessageID, err)
		}
	}
	for _, m := range messages {
		if _, err := status.add(m); err == nil {
			t.Fatalf("add duplicate message(%d) success", m.MessageID)
		}
	}
	for _, m := range messages {
		if _, ok := status.get(m.MessageID); !ok {
			t.Fatalf("not find state(%d)", m.MessageID)
		}
	}
	for _, m := range messages {
		status.del(m.MessageID)
	}
	for _, m := range messages {
		if _, ok := status.get(m.MessageID); ok {
			t.Fatalf("find deleted state(%d)", m.MessageID)
		}
	}
}

type IDGen struct {
	seq uint16
}

func (p *IDGen) gen() uint16 {
	p.seq++
	return p.seq
}

func NewTestClient(r base.Recver, s base.Sender, f func() uint16, bs uint32) *client {
	var b base.BaseLayer
	b.SetRecver(r)
	b.SetSender(s)
	var c client
	c.init(&b, f, bs)
	return &c
}

type TestSender struct {
	Count  int
	Buffer bytes.Buffer
}

func (s *TestSender) Send(m base.Message) error {
	s.Count++
	s.Buffer.Write(m.Payload)
	//fmt.Printf("%s: (%v)%v\n", m.String(), len(m.Payload), m.Payload)
	return nil
}

func MakeTestPayload(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(i)
	}
	return b
}

func TestClient(t *testing.T) {
	var id uint16
	f := func() uint16 { id++; return id }
	r := &base.CountRecver{ /*Writer: os.Stdout*/ }
	s := &TestSender{}
	c := NewTestClient(r, s, f, 16)
	p := MakeTestPayload(50)

	m := base.Message{
		Type:      base.CON,
		Code:      base.PUT,
		MessageID: f(),
		Token:     "1",
		Payload:   p,
	}
	if err := c.Send(m); err != nil {
		t.Fatalf("send: %v", err)
	}

	opt := base.BlockOption{}
	for i := 0; i <= 2; i++ {
		opt.Num = uint32(i)
		opt.More = true
		opt.Size = 16
		ack := base.Message{
			Type:      base.ACK,
			Code:      base.Continue,
			MessageID: id,
			Token:     "1",
		}
		ack.SetOption(base.Block1, opt.Value())
		if err := c.Recv(ack); err != nil {
			t.Fatalf("recv: %v", err)
		}
	}
	opt.Num = 3
	opt.More = false
	opt.Size = 16
	ack := base.Message{
		Type:      base.ACK,
		Code:      base.Changed,
		MessageID: id,
		Token:     "1",
	}
	ack.SetOption(base.Block1, opt.Value())
	if err := c.Recv(ack); err != nil {
		t.Fatalf("recv: %v", err)
	}

	if got, want := r.Count, 1; got != want {
		t.Errorf("recv: %d != %d", got, want)
	}
	if got, want := s.Count, 4; got != want {
		t.Errorf("send: %d != %d", got, want)
	}
	if got, want := s.Buffer.Bytes(), p; !reflect.DeepEqual(got, want) {
		t.Errorf("payload: (%d)%v != (%d)%v", len(got), got, len(want), want)
	}
}
