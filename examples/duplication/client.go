package main

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:5683")
	if err != nil {
		log.Fatalf("resolve udp addr: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("dial udp: %v", err)
	}

	messages := []base.Message{
		{
			Type:      base.CON,
			Code:      base.PUT,
			MessageID: 1,
			Token:     "1",
			Payload:   []byte("Confirmable Message"),
		},
		{
			Type:      base.NON,
			Code:      base.PUT,
			MessageID: 2,
			Token:     "2",
			Payload:   []byte("NonConfirmable Message"),
		},
	}
	for _, m := range messages {
		if err = SendMessage(conn, m, 3); err != nil {
			log.Fatalf("send %s: %v", m.String(), err)
		}
	}
}

func SendMessage(w io.Writer, m base.Message, retry int) error {
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	for i := 0; i < retry; i++ {
		if _, err := w.Write(data); err != nil {
			return err
		}
		log.Printf("write: %s", m.String())
		time.Sleep(time.Second)
	}
	return nil
}
