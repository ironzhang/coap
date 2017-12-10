package coaputil

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ironzhang/coap/internal/stack/base"
)

func makeEmptyOption(id uint16, value string) (base.Option, error) {
	return base.Option{ID: id}, nil
}

func makeUintOption(id uint16, value string) (base.Option, error) {
	u, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return base.Option{}, err
	}
	return base.Option{ID: id, Value: uint32(u)}, nil
}

func makeStringOption(id uint16, value string) (base.Option, error) {
	return base.Option{ID: id, Value: value}, nil
}

func makeOpaqueOption(id uint16, value string) (base.Option, error) {
	return base.Option{ID: id, Value: []byte(value)}, nil
}

func makeOption(format int, id uint16, value string) (base.Option, error) {
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
		return base.Option{}, fmt.Errorf("unsupport option format: %d", format)
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

func ParseOptionByName(s string) (base.Option, error) {
	name, value, err := splitOption(s)
	if err != nil {
		return base.Option{}, err
	}
	def, ok := base.LookupOptionDefByName(name)
	if !ok {
		return base.Option{}, fmt.Errorf("not found option define: %s", name)
	}
	return makeOption(def.Format, def.ID, value)
}

func ParseOptionByID(format int, s string) (base.Option, error) {
	name, value, err := splitOption(s)
	if err != nil {
		return base.Option{}, err
	}
	id, err := strconv.ParseUint(name, 10, 16)
	if err != nil {
		return base.Option{}, err
	}
	return makeOption(format, uint16(id), value)
}
