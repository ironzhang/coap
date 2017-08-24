package base

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/ironzhang/coap/internal/message"
)

func TestBaseLayer(t *testing.T) {
	l := BaseLayer{
		Name:   "base",
		Recver: NopRecver{os.Stdout},
		Sender: NopSender{os.Stdout},
	}

	l.Send(message.Message{})
	l.Recv(message.Message{})
	l.SendRST(1)
}

func TestBaseLayerError(t *testing.T) {
	l := BaseLayer{Name: "base"}
	fmt.Println(l.NewError(io.EOF))
	fmt.Println(l.Errorf(io.EOF, "read a.txt"))
}
