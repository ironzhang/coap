package coap

import (
	"errors"
	"io"
	"net"
)

type Client struct {
}

func (c *Client) SendRequest(req *Request, cb func(*Client, *Response)) error {
	if req.URL == nil {
		return errors.New("coap: nil Request.URL")
	}
	if len(req.URL.Host) <= 0 {
		return errors.New("coap: invalid Request.URL.Host")
	}

	addr, err := net.ResolveUDPAddr("udp", req.URL.Host)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	req.Options.SetPath(req.URL.Path)
	reqMsg := message{
		Type:      NON,
		Code:      req.Method,
		MessageID: 1,   // 生成MessageID
		Token:     nil, // 生成token
		Options:   req.Options,
		Payload:   req.Payload,
	}
	if req.Confirmable {
		reqMsg.Type = CON
	}

	if err = c.writeMessage(conn, reqMsg); err != nil {
		return err
	}

	if req.Confirmable {
		resMsg, err := c.readMessage(conn)
		if err != nil {
			return err
		}
		if cb != nil {
			resp := &Response{
				Ack:     resMsg.Type == ACK,
				Status:  resMsg.Code,
				Options: resMsg.Options,
				Token:   resMsg.Token,
				Payload: resMsg.Payload,
			}
			cb(c, resp)
		}
	}

	return nil
}

func (c *Client) writeMessage(w io.Writer, m message) error {
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	if _, err = w.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Client) readMessage(r io.Reader) (message, error) {
	buf := make([]byte, 1500)
	n, err := r.Read(buf)
	if err != nil {
		return message{}, err
	}
	return parseMessage(buf[:n])
}
