package coap

import "net/url"

type Request struct {
	Confirmable bool
	Method      Code
	Options     Options
	URL         *url.URL
	Token       string
	Payload     []byte
	Callback    func(*Response)
}

func NewRequest(confirmable bool, method Code, urlstr string, payload []byte) (*Request, error) {
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
