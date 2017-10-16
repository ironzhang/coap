package coap

import (
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

const defaultResponseTimeout = 20 * time.Second

type ackWaiter struct {
	done chan struct{}
	err  error
}

func newAckWaiter() *ackWaiter {
	return &ackWaiter{
		done: make(chan struct{}),
	}
}

func (w *ackWaiter) Done(err error) {
	w.err = err
	close(w.done)
}

func (w *ackWaiter) Wait() error {
	<-w.done
	return w.err
}

type responseWaiter struct {
	done      chan struct{}
	start     time.Time
	timeout   time.Duration
	messageID uint16
	err       error
	msg       base.Message
}

func newResponseWaiter() *responseWaiter {
	return &responseWaiter{
		done:    make(chan struct{}),
		start:   time.Now(),
		timeout: defaultResponseTimeout,
	}
}

func (w *responseWaiter) Timeout() bool {
	return time.Since(w.start) > w.timeout
}

func (w *responseWaiter) Done(msg base.Message, err error) {
	w.msg = msg
	w.err = err
	close(w.done)
}

func (w *responseWaiter) Wait() (*Response, error) {
	<-w.done
	if w.err != nil {
		return nil, w.err
	}
	return &Response{
		Ack:     w.msg.Type == base.ACK,
		Status:  Code(w.msg.Code),
		Options: w.msg.Options,
		Token:   Token(w.msg.Token),
		Payload: w.msg.Payload,
	}, nil
}
