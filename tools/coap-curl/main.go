package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ironzhang/coap"
	"github.com/ironzhang/coap/internal/stack/base"
	"github.com/ironzhang/coap/tools/coaputil"
)

type Options struct {
	Confirmable   bool
	Options       coaputil.StringsValue
	EmptyOptions  coaputil.StringsValue
	UintOptions   coaputil.StringsValue
	StringOptions coaputil.StringsValue
	OpaqueOptions coaputil.StringsValue
	Data          string
	InFile        string
	OutFile       string
	Method        coap.Code
	URL           string
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

// usage
// coap-curl --empty-option "" --uint-option "" --string-option "" --opaque-option"" --data '{"Name": "xx"}' url
func (a *Options) Parse() error {
	var err error
	var method string

	flag.BoolVar(&a.Confirmable, "con", false, "confirmable")
	flag.Var(&a.Options, "option", "option")
	flag.Var(&a.EmptyOptions, "empty-option", "empty option")
	flag.Var(&a.UintOptions, "uint-option", "uint option")
	flag.Var(&a.StringOptions, "string-option", "string option")
	flag.Var(&a.OpaqueOptions, "opaque-option", "opaque option")
	flag.StringVar(&a.Data, "data", "", "data")
	flag.StringVar(&a.InFile, "in-file", "", "in file")
	flag.StringVar(&a.OutFile, "out-file", "", "out file")
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

func MakePayload(data string, infile string) (payload []byte, err error) {
	if data != "" {
		return []byte(data), nil
	} else if infile != "" {
		payload, err = ioutil.ReadFile(infile)
		if err != nil {
			return nil, err
		}
		return payload, nil
	}
	return nil, nil
}

func AddOptionsByName(opts *coap.Options, ss []string) error {
	for _, s := range ss {
		opt, err := coaputil.ParseOptionByName(s)
		if err != nil {
			return err
		}
		opts.Add(opt.ID, opt.Value)
	}
	return nil
}

func AddOptionsByID(opts *coap.Options, format int, ss []string) error {
	for _, s := range ss {
		opt, err := coaputil.ParseOptionByID(format, s)
		if err != nil {
			return err
		}
		opts.Add(opt.ID, opt.Value)
	}
	return nil
}

func AddOptions(r *coap.Request, a *Options) (err error) {
	if err = AddOptionsByName(&r.Options, a.Options); err != nil {
		return err
	}
	if err = AddOptionsByID(&r.Options, base.EmptyValue, a.EmptyOptions); err != nil {
		return err
	}
	if err = AddOptionsByID(&r.Options, base.UintValue, a.UintOptions); err != nil {
		return err
	}
	if err = AddOptionsByID(&r.Options, base.StringValue, a.StringOptions); err != nil {
		return err
	}
	if err = AddOptionsByID(&r.Options, base.OpaqueValue, a.OpaqueOptions); err != nil {
		return err
	}
	return nil
}

func MakeRequest(a *Options) (*coap.Request, error) {
	payload, err := MakePayload(a.Data, a.InFile)
	if err != nil {
		return nil, err
	}
	req, err := coap.NewRequest(a.Confirmable, a.Method, a.URL, payload)
	if err != nil {
		return nil, err
	}
	if err = AddOptions(req, a); err != nil {
		return nil, err
	}
	return req, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	var opts Options
	err := opts.Parse()
	if err != nil {
		fmt.Printf("parse options: %v\n", err)
		flag.Usage()
		return
	}

	req, err := MakeRequest(&opts)
	if err != nil {
		fmt.Printf("make request: %v\n", err)
		return
	}
	coap.PrintRequest(os.Stdout, req, true)

	resp, err := coap.DefaultClient.SendRequest(req)
	if err != nil {
		fmt.Printf("send request: %v\n", err)
		return
	}
	coap.PrintResponse(os.Stdout, resp, true)

	if opts.OutFile != "" {
		if err = ioutil.WriteFile(opts.OutFile, resp.Payload, 0664); err != nil {
			fmt.Printf("write file: %v\n", err)
		}
	}
}
