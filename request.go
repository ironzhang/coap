package coap

import "net/url"

type Request struct {
	Confirmable bool
	Method      Code
	Options     Options
	URL         *url.URL
	Token       []byte
	Payload     []byte
}
