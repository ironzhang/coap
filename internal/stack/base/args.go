package base

import "time"

// CoAP协议参数
const (
	ACK_TIMEOUT       = 2 * time.Second
	ACK_RANDOM_FACTOR = 1.5
	MAX_RETRANSMIT    = 4
	NSTART            = 1
	DEFAULT_LEISURE   = 5 * time.Second
	PROBING_RATE      = 1 // 1 byte/second
)

// 传输参数衍生时间
const (
	MAX_TRANSMIT_SPAN = 45 * time.Second
	MAX_TRANSMIT_WAIT = 93 * time.Second
	MAX_LATENCY       = 100 * time.Second
	PROCESSING_DELAY  = 2 * time.Second
	MAX_RTT           = 202 * time.Second
	EXCHANGE_LIFETIME = 247 * time.Second
	NON_LIFETIME      = 145 * time.Second
)
