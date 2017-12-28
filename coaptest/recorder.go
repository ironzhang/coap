package coaptest

import (
	"bytes"

	"github.com/ironzhang/coap"
)

type ResponseRecorder struct {
	Confirmable bool
	Code        coap.Code
	Header      coap.Options
	Body        bytes.Buffer
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		Code:   coap.Content,
		Header: make(coap.Options, 0),
	}
}

func (rw *ResponseRecorder) Ack(code coap.Code) {
	rw.Code = code
}

func (rw *ResponseRecorder) SetConfirmable() {
	rw.Confirmable = true
}

func (rw *ResponseRecorder) Options() *coap.Options {
	return &rw.Header
}

func (rw *ResponseRecorder) WriteCode(code coap.Code) {
	rw.Code = code
}

func (rw *ResponseRecorder) Write(buf []byte) (int, error) {
	return rw.Body.Write(buf)
}
