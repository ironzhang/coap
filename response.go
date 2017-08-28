package coap

// Response COAP响应
type Response struct {
	// 是否为应答附带响应
	Ack bool

	// 响应码
	Status Code

	// COAP选项
	Options Options

	// 消息令牌
	Token string

	// 消息负载
	Payload []byte

	//RemoteAddr net.Addr
}
