package coap

import "net"

type Response struct {
	Ack        bool
	Status     Code
	Options    Options
	Token      string
	Payload    []byte
	RemoteAddr net.Addr
}
