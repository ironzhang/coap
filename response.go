package coap

import "net"

// Response COAP响应
type Response struct {
	// 是否为应答附带响应
	Ack bool

	// 响应码
	Status Code

	// COAP选项
	Options Options

	// 消息令牌
	Token Token

	// 消息负载
	Payload []byte

	// 远程地址，主动上报的响应(即由Observe接口处理的Response)，该字段才有意义
	RemoteAddr net.Addr
}
