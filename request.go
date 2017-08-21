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

	err   error
	donec chan struct{}
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
		donec:       make(chan struct{}),
	}
	return r, nil
}

func (r *Request) done(err error) {
	r.err = err
	close(r.donec)
}

func (r *Request) wait() error {
	<-r.donec
	return r.err
}
