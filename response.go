package coap

type Response struct {
	Ack     bool
	Status  Code
	Options Options
	Token   string
	Payload []byte
}
