package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/tools/coaputil"
)

type Args struct {
	Addr          string
	Type          int
	Code          string
	MessageID     int
	Token         string
	Options       coaputil.StringsValue
	EmptyOptions  coaputil.StringsValue
	UintOptions   coaputil.StringsValue
	StringOptions coaputil.StringsValue
	OpaqueOptions coaputil.StringsValue
	Payload       string
	Read          bool
}

func (p *Args) Parse() {
	flag.StringVar(&p.Addr, "addr", "localhost:5683", "address")
	flag.IntVar(&p.Type, "type", 0, "message type")
	flag.StringVar(&p.Code, "code", "0.00", "message code")
	flag.IntVar(&p.MessageID, "id", 0, "message id")
	flag.StringVar(&p.Token, "token", "", "")
	flag.Var(&p.Options, "option", "option")
	flag.Var(&p.EmptyOptions, "empty-option", "empty option")
	flag.Var(&p.UintOptions, "uint-option", "uint option")
	flag.Var(&p.StringOptions, "string-option", "string option")
	flag.Var(&p.OpaqueOptions, "opaque-option", "opaque option")
	flag.StringVar(&p.Payload, "payload", "", "message payload")
	flag.BoolVar(&p.Read, "read", false, "read message")
	flag.Parse()
}

func AddOptionsByName(m *base.Message, ss []string) error {
	for _, s := range ss {
		opt, err := coaputil.ParseOptionByName(s)
		if err != nil {
			return err
		}
		m.AddOption(opt.ID, opt.Value)
	}
	return nil
}

func AddOptionsByID(m *base.Message, format int, ss []string) error {
	for _, s := range ss {
		opt, err := coaputil.ParseOptionByID(format, s)
		if err != nil {
			return err
		}
		m.AddOption(opt.ID, opt.Value)
	}
	return nil
}

func AddOptions(m *base.Message, a *Args) (err error) {
	if err = AddOptionsByName(m, a.Options); err != nil {
		return err
	}
	if err = AddOptionsByID(m, base.EmptyValue, a.EmptyOptions); err != nil {
		return err
	}
	if err = AddOptionsByID(m, base.UintValue, a.UintOptions); err != nil {
		return err
	}
	if err = AddOptionsByID(m, base.StringValue, a.StringOptions); err != nil {
		return err
	}
	if err = AddOptionsByID(m, base.OpaqueValue, a.OpaqueOptions); err != nil {
		return err
	}
	return nil
}

func ParseCode(s string) (uint8, error) {
	ss := strings.Split(s, ".")
	c, err := strconv.ParseUint(ss[0], 10, 8)
	if err != nil {
		return 0, err
	}
	c = (c & 0xff) << 5
	if len(ss) > 1 {
		u, err := strconv.ParseUint(ss[1], 10, 8)
		if err != nil {
			return 0, err
		}
		c |= (u & 0x1f)
	}
	return uint8(c), nil
}

func MakeMessage(a *Args) (base.Message, error) {
	c, err := ParseCode(a.Code)
	if err != nil {
		return base.Message{}, err
	}
	m := base.Message{
		Type:      uint8(a.Type),
		Code:      c,
		MessageID: uint16(a.MessageID),
		Token:     a.Token,
		Payload:   []byte(a.Payload),
	}
	if err := AddOptions(&m, a); err != nil {
		return base.Message{}, err
	}
	return m, nil
}

func WriteMessage(w io.Writer, m base.Message) error {
	data, err := m.Marshal()
	if err != nil {
		return err
	}
	if _, err = w.Write(data); err != nil {
		return err
	}
	return nil
}

func ReadMessage(r io.Reader) (m base.Message, err error) {
	var buf [1500]byte
	n, err := r.Read(buf[:])
	if err != nil {
		return base.Message{}, err
	}
	if err = m.Unmarshal(buf[:n]); err != nil {
		return base.Message{}, err
	}
	return m, nil
}

var mser = base.MessageStringer{
	WritePayload: func(w io.Writer, payload []byte) {
		fmt.Fprintf(w, "%s\n", payload)
	},
}

func PrintMessage(w io.Writer, m base.Message) {
	fmt.Fprintf(w, mser.MessageString(m))
}

func main() {
	var args Args
	args.Parse()

	conn, err := net.Dial("udp", args.Addr)
	if err != nil {
		fmt.Printf("dial: %v", err)
		return
	}

	msg, err := MakeMessage(&args)
	if err != nil {
		fmt.Printf("make message: %v", err)
		return
	}

	fmt.Printf("coap server: %v\n", args.Addr)
	PrintMessage(os.Stdout, msg)
	if err = WriteMessage(conn, msg); err != nil {
		fmt.Printf("write message: %v", err)
		return
	}

	if args.Read {
		rmsg, err := ReadMessage(conn)
		if err != nil {
			fmt.Printf("read message: %v", err)
			return
		}
		PrintMessage(os.Stdout, rmsg)
	}
}
