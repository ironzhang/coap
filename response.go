package coap

import "github.com/ironzhang/coap/internal/message"

type Response struct {
	Ack     bool
	Status  message.Code
	Options Options
	Token   string
	Payload []byte
	//RemoteAddr net.Addr
}
