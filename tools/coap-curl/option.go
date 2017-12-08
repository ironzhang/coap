package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ironzhang/coap"
	"github.com/ironzhang/coap/internal/stack/base"
)

type option struct {
	id    uint16
	value interface{}
}

func makeEmptyOption(id uint16, value string) (option, error) {
	return option{id: id}, nil
}

func makeUintOption(id uint16, value string) (option, error) {
	u, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return option{}, err
	}
	return option{id: id, value: uint32(u)}, nil
}

func makeStringOption(id uint16, value string) (option, error) {
	return option{id: id, value: value}, nil
}

func makeOpaqueOption(id uint16, value string) (option, error) {
	return option{id: id, value: []byte(value)}, nil
}

func makeOption(format int, id uint16, value string) (option, error) {
	switch format {
	case base.EmptyValue:
		return makeEmptyOption(id, value)
	case base.UintValue:
		return makeUintOption(id, value)
	case base.StringValue:
		return makeStringOption(id, value)
	case base.OpaqueValue:
		return makeOpaqueOption(id, value)
	default:
		return option{}, fmt.Errorf("unsupport option format: %d", format)
	}
}

func splitOption(s string) (string, string, error) {
	ss := strings.Split(s, ":")
	n := len(ss)
	if n == 1 {
		return strings.TrimSpace(ss[0]), "", nil
	} else if n == 2 {
		return strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1]), nil
	} else {
		return "", "", fmt.Errorf("option format ill: %s", s)
	}
}

func parseNameOption(s string) (option, error) {
	name, value, err := splitOption(s)
	if err != nil {
		return option{}, err
	}
	def, ok := base.LookupOptionDefByName(name)
	if !ok {
		return option{}, fmt.Errorf("not found option define: %s", name)
	}
	return makeOption(def.Format, def.ID, value)
}

func parseIDOption(format int, s string) (option, error) {
	name, value, err := splitOption(s)
	if err != nil {
		return option{}, err
	}
	id, err := strconv.ParseUint(name, 10, 16)
	if err != nil {
		return option{}, err
	}
	return makeOption(format, uint16(id), value)
}

func addNameOptions(opts *coap.Options, ss []string) error {
	for _, s := range ss {
		opt, err := parseNameOption(s)
		if err != nil {
			return err
		}
		opts.Add(opt.id, opt.value)
	}
	return nil
}

func addIDOptions(opts *coap.Options, format int, ss []string) error {
	for _, s := range ss {
		opt, err := parseIDOption(format, s)
		if err != nil {
			return err
		}
		opts.Add(opt.id, opt.value)
	}
	return nil
}
