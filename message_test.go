package coap

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEncodeUintVariant(t *testing.T) {
	tests := []struct {
		val uint32
		buf []byte
	}{
		{val: 0, buf: nil},
		{val: 0x01, buf: []byte{0x01}},
		{val: 0x0201, buf: []byte{0x02, 0x01}},
		{val: 0x030201, buf: []byte{0x03, 0x02, 0x01}},
		{val: 0x04030201, buf: []byte{0x04, 0x03, 0x02, 0x01}},
	}
	for i, tt := range tests {
		if got, want := encodeUintVariant(tt.val), tt.buf; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: got(%v) != want(%v)", i, got, want)
		}
	}
}

func TestDecodeUintVariant(t *testing.T) {
	tests := []struct {
		val uint32
		buf []byte
	}{
		{val: 0, buf: nil},
		{val: 0x01, buf: []byte{0x01}},
		{val: 0x0201, buf: []byte{0x02, 0x01}},
		{val: 0x030201, buf: []byte{0x03, 0x02, 0x01}},
		{val: 0x030201, buf: []byte{0x00, 0x03, 0x02, 0x01}},
		{val: 0x04030201, buf: []byte{0x04, 0x03, 0x02, 0x01}},
	}
	for i, tt := range tests {
		if got, want := decodeUintVariant(tt.buf), tt.val; got != want {
			t.Errorf("case%d: got(0x%x) != want(0x%x)", i, got, want)
		}
	}
}

func TestOptionValueToBytes(t *testing.T) {
	tests := []struct {
		val interface{}
		buf []byte
	}{
		{"", []byte{}},
		{[]byte{}, []byte{}},
		{"x", []byte{'x'}},
		{[]byte{'x'}, []byte{'x'}},
		{MediaType(3), []byte{0x3}},
		{3, []byte{0x3}},
		{838, []byte{0x3, 0x46}},
		{int32(838), []byte{0x3, 0x46}},
		{uint(838), []byte{0x3, 0x46}},
		{uint32(838), []byte{0x3, 0x46}},
	}
	for i, tt := range tests {
		buf, err := optionValueToBytes(tt.val)
		if err != nil {
			t.Errorf("case%d: optionValueToBytes: %v", i, err)
			continue
		}
		if got, want := buf, tt.buf; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: got(%v) != want(%v)", i, got, want)
		}
	}
}

func TestBytesToOptionValue(t *testing.T) {
	tests := []struct {
		id  OptionID
		buf []byte
		val interface{}
	}{
		{IfMatch, []byte{}, []byte{}},
		{IfMatch, []byte{'x'}, []byte{'x'}},
		{URIHost, []byte{'x'}, "x"},
		{IfNoneMatch, []byte{}, []byte{}},
		{URIPort, []byte{0x03, 0x46}, uint32(838)},
		{ContentFormat, []byte{0x03}, MediaType(3)},
	}
	for i, tt := range tests {
		val := bytesToOptionValue(tt.id, tt.buf)
		if val == nil {
			t.Errorf("case%d: bytesToOptionValue return nil", i)
			continue
		}
		if got, want := val, tt.val; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: got(%v) != want(%v)", i, got, want)
		}
	}
}

func TestOptionEncoder(t *testing.T) {
	tests := []struct {
		delta uint32
		value []byte
		data  []byte
	}{
		{delta: 0, value: nil, data: []byte{0x00}},
		{delta: 1, value: nil, data: []byte{0x10}},
		{delta: 2, value: []byte{0x00, 0x01, 0x02, 0x03}, data: []byte{0x24, 0x00, 0x01, 0x02, 0x03}},
		{delta: 256, value: []byte{0x00, 0x01, 0x02, 0x03}, data: []byte{0xd4, 0xf3, 0x00, 0x01, 0x02, 0x03}},
		{delta: 512, value: []byte{0x00, 0x01, 0x02, 0x03}, data: []byte{0xe4, 0x00, 0xf3, 0x00, 0x01, 0x02, 0x03}},
	}
	for i, tt := range tests {
		var buf bytes.Buffer
		e := optionEncoder{w: &buf}
		if err := e.Encode(tt.delta, tt.value); err != nil {
			t.Errorf("case%d: encode option: %v", i, err)
			continue
		}
		if got, want := buf.Bytes(), tt.data; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: got(%v) != want(%v)", i, got, want)
			continue
		}
	}
}

func TestOptionDecoder(t *testing.T) {
	tests := []struct {
		delta uint32
		value []byte
		data  []byte
	}{
		{delta: 0, value: nil, data: []byte{0x00}},
		{delta: 1, value: nil, data: []byte{0x10}},
		{delta: 2, value: []byte{0x00, 0x01, 0x02, 0x03}, data: []byte{0x24, 0x00, 0x01, 0x02, 0x03}},
		{delta: 256, value: []byte{0x00, 0x01, 0x02, 0x03}, data: []byte{0xd4, 0xf3, 0x00, 0x01, 0x02, 0x03}},
		{delta: 512, value: []byte{0x00, 0x01, 0x02, 0x03}, data: []byte{0xe4, 0x00, 0xf3, 0x00, 0x01, 0x02, 0x03}},
	}
	for i, tt := range tests {
		buf := bytes.NewBuffer(tt.data)
		d := optionDecoder{r: buf}
		flag, _ := buf.ReadByte()
		delta, value, err := d.Decode(flag)
		if err != nil {
			t.Errorf("case%d: decode option: %v", i, err)
			continue
		}
		if got, want := delta, tt.delta; got != want {
			t.Errorf("case%d: delta: got(%v) != want(%v)", i, got, want)
			continue
		}
		if got, want := value, tt.value; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: value: got(%v) != want(%v)", i, got, want)
			continue
		}
	}
}

func TestMessageAddOption(t *testing.T) {
	inputs := []struct {
		id  OptionID
		val interface{}
	}{
		{OptionID(1), 1},
		{OptionID(1), 2},
		{OptionID(1), 3},
		{OptionID(2), 1},
		{OptionID(2), 2},
		{OptionID(2), 3},
		{OptionID(4), 1},
	}
	options := []Option{
		{ID: OptionID(1), Values: []interface{}{1, 2, 3}},
		{ID: OptionID(2), Values: []interface{}{1, 2, 3}},
		{ID: OptionID(4), Values: []interface{}{1}},
	}

	var m message
	for _, i := range inputs {
		m.addOption(i.id, i.val)
	}
	if got, want := m.Options, options; !reflect.DeepEqual(got, want) {
		t.Errorf("options: got(%v) != want(%v)", got, want)
	}
}

func TestMessage(t *testing.T) {
	tests := []struct {
		m message
		b []byte
	}{
		{
			m: message{
				Type:      CON,
				Code:      GET,
				MessageID: 12345,
			},
			b: []byte{0x40, 0x1, 0x30, 0x39},
		},
		{
			m: message{
				Type:      CON,
				Code:      GET,
				MessageID: 12345,
				Options: []Option{
					{ID: ETag, Values: []interface{}{[]byte("weetag")}},
					{ID: MaxAge, Values: []interface{}{uint32(3)}},
				},
			},
			b: []byte{
				0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
				0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
			},
		},
		{
			m: message{
				Type:      CON,
				Code:      GET,
				MessageID: 12345,
				Options: []Option{
					{ID: ETag, Values: []interface{}{[]byte("weetag")}},
					{ID: MaxAge, Values: []interface{}{uint32(3)}},
				},
				Payload: []byte("hi"),
			},
			b: []byte{
				0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
				0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
				0xff, 'h', 'i',
			},
		},
	}
	for i, tt := range tests {
		b, err := tt.m.Marshal()
		if err != nil {
			t.Fatalf("case%d: message marshal: %v", i, err)
			continue
		}
		if got, want := b, tt.b; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: message marshal: got(%v) != want(%v)", i, got, want)
			continue
		}
	}
	for i, tt := range tests {
		var m message
		err := m.Unmarshal(tt.b)
		if err != nil {
			t.Fatalf("case%d: message unmarshal: %v", i, err)
			continue
		}
		if got, want := m, tt.m; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: message unmarshal: got(%#v) != want(%#v)", i, got, want)
			continue
		}
	}
}

func TestInvalidMessageParsing(t *testing.T) {
	var invalidPackets = [][]byte{
		nil,
		{0x40},
		{0x40, 0},
		{0x40, 0, 0},
		{0xff, 0, 0, 0, 0, 0},
		{0x4f, 0, 0, 0, 0, 0},
		{0x45, 0, 0, 0, 0, 0},                // TKL=5 but packet is truncated
		{0x40, 0x01, 0x30, 0x39, 0x4d},       // Extended word length but no extra length byte
		{0x40, 0x01, 0x30, 0x39, 0x4e, 0x01}, // Extended word length but no full extra length word
	}
	for _, data := range invalidPackets {
		var m message
		if err := m.Unmarshal(data); err == nil {
			t.Errorf("Unexpected success parsing short message (%#v): %v", data, m)
		} else {
			t.Logf("short message unmarshal: (%#v): %v", data, err)
		}
	}
}
