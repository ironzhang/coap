package coap

import "testing"

func TestArgs(t *testing.T) {
	t.Logf("MAX_TRANSMIT_SPAN: %f", MAX_TRANSMIT_SPAN.Seconds())
	t.Logf("MAX_TRANSMIT_WAIT: %f", MAX_TRANSMIT_WAIT.Seconds())
}
