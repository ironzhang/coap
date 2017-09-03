package base

import (
	"fmt"
	"io"
	"time"
)

type Recver interface {
	Recv(m Message) error
	OnAckTimeout(m Message)
}

type Sender interface {
	Send(m Message) error
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
	m := Message{
		Type:      RST,
		MessageID: messageID,
	}
	return l.Send(m)
}

type CountRecver struct {
	Writer  io.Writer
	Count   int
	Timeout int
}

func (p *CountRecver) Recv(m Message) error {
	if p.Writer != nil {
		fmt.Fprintf(p.Writer, "[%s] Recv: %v\n", time.Now(), m.String())
	}
	p.Count++
	return nil
}

func (p *CountRecver) OnAckTimeout(m Message) {
	if p.Writer != nil {
		fmt.Fprintf(p.Writer, "[%s] Recv: %v\n", time.Now(), m.String())
	}
	p.Timeout++
}

type CountSender struct {
	Writer io.Writer
	Count  int
}

func (p *CountSender) Send(m Message) error {
	if p.Writer != nil {
		fmt.Fprintf(p.Writer, "[%s] Send: %v\n", time.Now(), m.String())
	}
	p.Count++
	return nil
}
