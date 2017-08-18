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

func NewRequest(confirmable bool, method Code, urlstr string, payload []byte) (*Request, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	return &Request{
		Confirmable: confirmable,
		Method:      method,
		URL:         u,
		Payload:     payload,
	}, nil
}
