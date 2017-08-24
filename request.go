package coap

import (
	"net"
	"net/url"

	"github.com/ironzhang/coap/internal/message"
)

type Request struct {
	Confirmable bool
	Method      message.Code
	Options     Options
	URL         *url.URL
	Token       string
	Payload     []byte
	RemoteAddr  net.Addr
	Callback    func(*Response)
}

func NewRequest(confirmable bool, method message.Code, urlstr string, payload []byte) (*Request, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	r := &Request{
		Confirmable: confirmable,
		Method:      method,
		URL:         u,
		Payload:     payload,
	}
	return r, nil
}
