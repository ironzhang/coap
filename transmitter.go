package coap

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

type callback func(m message)

type WaitAck func(timeout time.Duration) bool

type transmitter struct {
	conn net.PacketConn
	addr net.Addr

	dones     map[uint16]func()
	callbacks map[string]callback
}

func (t *transmitter) Recv(m message) {
	switch m.Type {
	case CON, NON:
		// 处理单独响应

	case ACK:
		// 处理ACK

	case RST:
		// 处理RST
	}
}

func (t *transmitter) handleResp(m message) {
	token := string(m.Token)
	if cb, ok := t.callbacks[token]; ok {
		delete(t.callbacks, token)
		cb(m)
	}
}

func (t *transmitter) handleACK(m message) {
	if done, ok := t.dones[m.MessageID]; ok {
		delete(t.dones, m.MessageID)
		done()
	}
	t.handleResp(m)
}

func (t *transmitter) handleRST(m message) {
}

func (t *transmitter) Send(m message, cb callback) (WaitAck, error) {
	token := string(m.Token)
	if _, ok := t.callbacks[token]; ok {
		return func(d time.Duration) bool { return true }, errors.Errorf("duplicate token(%s)", token)
	}
	if err := t.send(m); err != nil {
		return func(d time.Duration) bool { return true }, err
	}
	t.callbacks[token] = cb

	if m.Type != CON {
		return func(d time.Duration) bool { return true }, nil
	}

	done := make(chan struct{})
	t.dones[m.MessageID] = func() { close(done) }
	return func(d time.Duration) bool {
		<-done
		return true
	}, nil
}

func (t *transmitter) send(m message) error {
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	if _, err = t.conn.WriteTo(data, t.addr); err != nil {
		return err
	}
	return nil
}
