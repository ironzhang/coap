package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ironzhang/coap"
	"github.com/ironzhang/coap/internal/stack/base"
)

type StringsValue []string

func (p *StringsValue) Set(s string) error {
	*p = append(*p, s)
	return nil
}

func (p *StringsValue) String() string {
	return strings.Join(*p, ",")
}

func ParseMethod(s string) (coap.Code, error) {
	switch strings.ToUpper(s) {
	case "GET":
		return coap.GET, nil
	case "POST":
		return coap.POST, nil
	case "PUT":
		return coap.PUT, nil
	case "DELETE":
		return coap.DELETE, nil
	default:
		return 0, fmt.Errorf("unknown coap method: %v", s)
	}
}

type Args struct {
	Confirmable   bool
	Options       StringsValue
	EmptyOptions  StringsValue
	UintOptions   StringsValue
	StringOptions StringsValue
	OpaqueOptions StringsValue
	Data          string
	DataFile      string
	Method        coap.Code
	URL           string
}

// usage
// coap-curl --empty-option "" --uint-option "" --string-option "" --opaque-option"" --data '{"Name": "xx"}' --data-file xxx.coap url
func (a *Args) Parse() error {
	var err error
	var method string

	flag.BoolVar(&a.Confirmable, "con", true, "confirmable")
	flag.Var(&a.Options, "option", "option")
	flag.Var(&a.EmptyOptions, "empty-option", "empty option")
	flag.Var(&a.UintOptions, "uint-option", "uint option")
	flag.Var(&a.StringOptions, "string-option", "string option")
	flag.Var(&a.OpaqueOptions, "opaque-option", "opaque option")
	flag.StringVar(&a.Data, "data", "", "data")
	flag.StringVar(&a.DataFile, "data-file", "", "data file")
	flag.StringVar(&method, "X", "GET", "method")
	flag.IntVar(&coap.Verbose, "verbose", 0, "verbose")
	flag.Parse()

	a.Method, err = ParseMethod(method)
	if err != nil {
		return err
	}

	args := flag.Args()
	if len(args) < 1 {
		return fmt.Errorf("no url")
	}
	a.URL = args[0]

	return nil
}

func AddOptions(r *coap.Request, a *Args) (err error) {
	if err = addNameOptions(&r.Options, a.Options); err != nil {
		return err
	}
	if err = addIDOptions(&r.Options, base.EmptyValue, a.EmptyOptions); err != nil {
		return err
	}
	if err = addIDOptions(&r.Options, base.UintValue, a.UintOptions); err != nil {
		return err
	}
	if err = addIDOptions(&r.Options, base.StringValue, a.StringOptions); err != nil {
		return err
	}
	if err = addIDOptions(&r.Options, base.OpaqueValue, a.OpaqueOptions); err != nil {
		return err
	}
	return nil
}

func MakeRequest(a *Args) (*coap.Request, error) {
	req, err := coap.NewRequest(a.Confirmable, a.Method, a.URL, []byte(a.Data))
	if err != nil {
		return nil, err
	}
	if err = AddOptions(req, a); err != nil {
		return nil, err
	}
	return req, nil
}

func PrintRequest(r *coap.Request) {
	fmt.Printf("CON[%t] %s %s\n", r.Confirmable, r.Method, r.URL.String())
	r.Options.Write(os.Stdout)
	fmt.Printf("\n%s\n\n", r.Payload)
}

func PrintResponse(r *coap.Response) {
	fmt.Printf("ACK[%t] %s", r.Ack, r.Status)
	r.Options.Write(os.Stdout)
	fmt.Printf("\n%s\n\n", r.Payload)
}

func main() {
	var args Args
	err := args.Parse()
	if err != nil {
		fmt.Printf("parse args: %v\n", err)
		flag.Usage()
		return
	}

	req, err := MakeRequest(&args)
	if err != nil {
		fmt.Printf("make request: %v\n", err)
		return
	}
	PrintRequest(req)

	resp, err := coap.DefaultClient.SendRequest(req)
	if err != nil {
		fmt.Printf("send request: %v\n", err)
		return
	}
	PrintResponse(resp)
}
