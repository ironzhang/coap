package coaputil

import "strings"

type StringsValue []string

func (p *StringsValue) Set(s string) error {
	*p = append(*p, s)
	return nil
}

func (p *StringsValue) String() string {
	return strings.Join(*p, ",")
}
