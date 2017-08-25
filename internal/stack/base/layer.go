package base

import (
	"fmt"
	"io"
	"time"

	"github.com/ironzhang/coap/internal/message"
)

type Recver interface {
	Recv(m message.Message) error
}

type Sender interface {
	Send(m message.Message) error
}

type Setter interface {
	SetRecver(Recver)
	SetSender(Sender)
}

type Layer interface {
	Update()
	Recver
	Sender
	Setter
}

type BaseLayer struct {
	Name string
	Recver
	Sender
}

func (l *BaseLayer) SetRecver(recver Recver) {
	l.Recver = recver
}

func (l *BaseLayer) SetSender(sender Sender) {
	l.Sender = sender
}

func (l *BaseLayer) NewError(cause error) error {
	return Error{Layer: l.Name, Cause: cause}
}

func (l *BaseLayer) Errorf(cause error, format string, a ...interface{}) error {
	return Error{Layer: l.Name, Cause: cause, Details: fmt.Sprintf(format, a...)}
}

func (l *BaseLayer) SendRST(messageID uint16) error {
	m := message.Message{
		Type:      message.RST,
		Code:      0,
		MessageID: messageID,
	}
	return l.Send(m)
}

type CountRecver struct {
	Writer io.Writer
	Count  int
}

func (p *CountRecver) Recv(m message.Message) error {
	if p.Writer != nil {
		fmt.Fprintf(p.Writer, "[%s] Recv: %v\n", time.Now(), m.String())
	}
	p.Count++
	return nil
}

type CountSender struct {
	Writer io.Writer
	Count  int
}

func (p *CountSender) Send(m message.Message) error {
	if p.Writer != nil {
		fmt.Fprintf(p.Writer, "[%s] Send: %v\n", time.Now(), m.String())
	}
	p.Count++
	return nil
}
