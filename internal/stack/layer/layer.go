package layer

import (
	"fmt"

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
	Setter
	Recver
	Sender
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
	return &Error{Layer: l.Name, Cause: cause}
}

func (l *BaseLayer) Errorf(cause error, format string, a ...interface{}) error {
	return &Error{Layer: l.Name, Cause: cause, Details: fmt.Sprintf(format, a...)}
}

func (l *BaseLayer) SendRST(messageID uint16) error {
	m := message.Message{
		Type:      message.RST,
		Code:      0,
		MessageID: messageID,
	}
	return l.Send(m)
}
