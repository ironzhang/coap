package base

import (
	"bytes"
	"reflect"
	"testing"
)

func TestCode(t *testing.T) {
	tests := []struct {
		code uint8
		want uint8
	}{
		{code: GET, want: 1},
		{code: POST, want: 2},
		{code: PUT, want: 3},
		{code: DELETE, want: 4},

		{code: Created, want: 65},
		{code: Deleted, want: 66},
		{code: Valid, want: 67},
		{code: Changed, want: 68},
		{code: Content, want: 69},
		{code: Continue, want: 95},

		{code: BadRequest, want: 128},
		{code: Unauthorized, want: 129},
		{code: BadOption, want: 130},
		{code: Forbidden, want: 131},
		{code: NotFound, want: 132},
		{code: MethodNotAllowed, want: 133},
		{code: NotAcceptable, want: 134},
		{code: PreconditionFailed, want: 140},
		{code: RequestEntityTooLarge, want: 141},
		{code: UnsupportedContentFormat, want: 143},

		{code: InternalServerError, want: 160},
		{code: NotImplemented, want: 161},
		{code: BadGateway, want: 162},
		{code: ServiceUnavailable, want: 163},
		{code: GatewayTimeout, want: 164},
		{code: ProxyingNotSupported, want: 165},
	}
	for _, tt := range tests {
		if tt.code != tt.want {
			t.Errorf("%s: %d != %d", CodeName(tt.code), tt.code, tt.want)
		}
	}
}

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
		id  uint16
		buf []byte
		val interface{}
	}{
		{IfMatch, []byte{}, []byte{}},
		{IfMatch, []byte{'x'}, []byte{'x'}},
		{URIHost, []byte{'x'}, "x"},
		{IfNoneMatch, []byte{}, []byte{}},
		{URIPort, []byte{0x03, 0x46}, uint32(838)},
		{ContentFormat, []byte{0x03}, uint32(3)},
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

func TestMessageString(t *testing.T) {
	tests := []struct {
		m Message
		s string
	}{
		{
			m: Message{Type: ACK, Code: GET, MessageID: 1, Token: string([]byte{1, 2, 3, 4, 0xff})},
			s: "Acknowledgement,GET,1,01020304ff",
		},
	}
	for i, tt := range tests {
		if got, want := tt.m.String(), tt.s; got != want {
			t.Errorf("case%d: %q != %q", i, got, want)
		}
	}
}

func TestMessageGetOption(t *testing.T) {
	options := []Option{
		{1, "1"},
		{2, "2"},
		{3, 3},
	}
	m := Message{Options: options}
	for _, o := range options {
		if got, want := m.GetOption(o.ID), o.Value; !reflect.DeepEqual(got, want) {
			t.Errorf("id=%d: %#v != %#v", o.ID, got, want)
		}
	}
}

func TestMessage(t *testing.T) {
	tests := []struct {
		m Message
		b []byte
	}{
		{
			m: Message{
				Type:      CON,
				Code:      GET,
				MessageID: 12345,
			},
			b: []byte{0x40, 0x1, 0x30, 0x39},
		},
		{
			m: Message{
				Type:      CON,
				Code:      GET,
				MessageID: 12345,
				Options: []Option{
					{ID: ETag, Value: []byte("weetag")},
					{ID: MaxAge, Value: uint32(3)},
				},
			},
			b: []byte{
				0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
				0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
			},
		},
		{
			m: Message{
				Type:      CON,
				Code:      GET,
				MessageID: 12345,
				Options: []Option{
					{ID: ETag, Value: []byte("weetag")},
					{ID: MaxAge, Value: uint32(3)},
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
		var m Message
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
		var m Message
		if err := m.Unmarshal(data); err == nil {
			t.Errorf("Unexpected success parsing short message (%#v): %v", data, m)
		} else {
			t.Logf("short message unmarshal: (%#v): %v", data, err)
		}
	}
}
