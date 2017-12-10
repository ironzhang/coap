package coap

import (
	"fmt"
	"io"
)

func PrintRequest(w io.Writer, r *Request, body bool) {
	fmt.Fprintf(w, "CON[%t] %s %s\n", r.Confirmable, r.Method, r.URL.String())
	r.Options.Write(w)
	if body {
		fmt.Fprintf(w, "\n%s\n", r.Payload)
	}
}

func PrintResponse(w io.Writer, r *Response, body bool) {
	fmt.Fprintf(w, "ACK[%t] %s", r.Ack, r.Status)
	r.Options.Write(w)
	if body {
		fmt.Fprintf(w, "\n%s\n", r.Payload)
	}
}
