package coap

import (
	"time"

	"github.com/ironzhang/coap/internal/message"
)

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
	msg       message.Message
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

func (w *responseWaiter) Done(msg message.Message, err error) {
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
		Ack:     w.msg.Type == ACK,
		Status:  w.msg.Code,
		Options: w.msg.Options,
		Token:   w.msg.Token,
		Payload: w.msg.Payload,
	}, nil
}
