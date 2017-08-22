package coap

import "time"

// CoAP协议参数
var (
	ACK_TIMEOUT       = 2 * time.Second
	ACK_RANDOM_FACTOR = 1.5
	MAX_RETRANSMIT    = 4
	NSTART            = 1
	DEFAULT_LEISURE   = 5 * time.Second
	PROBING_RATE      = 1 // 1 byte/second
)

var (
	//MAX_TRANSMIT_SPAN = time.Duration(float64(ACK_TIMEOUT*time.Duration(MAX_RETRANSMIT*MAX_RETRANSMIT-1)) * ACK_RANDOM_FACTOR)
	//MAX_TRANSMIT_WAIT = time.Duration(float64(ACK_TIMEOUT) * float64((MAX_RETRANSMIT+1)*(MAX_RETRANSMIT+1)-1) * ACK_RANDOM_FACTOR)

	MAX_TRANSMIT_SPAN = 45 * time.Second
	MAX_TRANSMIT_WAIT = 93 * time.Second
	MAX_LATENCY       = 100 * time.Second
	PROCESSING_DELAY  = 2 * time.Second
	MAX_RTT           = 202 * time.Second
	EXCHANGE_LIFETIME = 247 * time.Second
	NON_LIFETIME      = 145 * time.Second
)
