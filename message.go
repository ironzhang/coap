package coap

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
)

type Type uint8

const (
	CON Type = 0
	NON Type = 1
	ACK Type = 2
	RST Type = 3
)

var typeNames = [256]string{
	CON: "Confirmable",
	NON: "NonConfirmable",
	ACK: "Acknowledgement",
	RST: "Reset",
}

func init() {
	for i := range typeNames {
		if typeNames[i] == "" {
			typeNames[i] = fmt.Sprintf("Unknown (0x%x)", i)
		}
	}
}

func (t Type) String() string {
	return typeNames[t]
}

type Code uint8

const (
	// Request Codes
	GET    Code = 1
	POST   Code = 2
	PUT    Code = 3
	DELETE Code = 4

	// Responses Codes
	Created               Code = 65
	Deleted               Code = 66
	Valid                 Code = 67
	Changed               Code = 68
	Content               Code = 69
	BadRequest            Code = 128
	Unauthorized          Code = 129
	BadOption             Code = 130
	Forbidden             Code = 131
	NotFound              Code = 132
	MethodNotAllowed      Code = 133
	NotAcceptable         Code = 134
	PreconditionFailed    Code = 140
	RequestEntityTooLarge Code = 141
	UnsupportedMediaType  Code = 143
	InternalServerError   Code = 160
	NotImplemented        Code = 161
	BadGateway            Code = 162
	ServiceUnavailable    Code = 163
	GatewayTimeout        Code = 164
	ProxyingNotSupported  Code = 165
)

var codeNames = [256]string{
	GET:                   "GET",
	POST:                  "POST",
	PUT:                   "PUT",
	DELETE:                "DELETE",
	Created:               "Created",
	Deleted:               "Deleted",
	Valid:                 "Valid",
	Changed:               "Changed",
	Content:               "Content",
	BadRequest:            "BadRequest",
	Unauthorized:          "Unauthorized",
	BadOption:             "BadOption",
	Forbidden:             "Forbidden",
	NotFound:              "NotFound",
	MethodNotAllowed:      "MethodNotAllowed",
	NotAcceptable:         "NotAcceptable",
	PreconditionFailed:    "PreconditionFailed",
	RequestEntityTooLarge: "RequestEntityTooLarge",
	UnsupportedMediaType:  "UnsupportedMediaType",
	InternalServerError:   "InternalServerError",
	NotImplemented:        "NotImplemented",
	BadGateway:            "BadGateway",
	ServiceUnavailable:    "ServiceUnavailable",
	GatewayTimeout:        "GatewayTimeout",
	ProxyingNotSupported:  "ProxyingNotSupported",
}

func init() {
	for i := range codeNames {
		if codeNames[i] == "" {
			codeNames[i] = fmt.Sprintf("Unknown (0x%x)", i)
		}
	}
}

func (c Code) String() string {
	return codeNames[c]
}

// MediaType specifies the content type of a message.
type MediaType uint16

// Content types.
const (
	TextPlain     MediaType = 0  // text/plain;charset=utf-8
	AppLinkFormat MediaType = 40 // application/link-format
	AppXML        MediaType = 41 // application/xml
	AppOctets     MediaType = 42 // application/octet-stream
	AppExi        MediaType = 47 // application/exi
	AppJSON       MediaType = 50 // application/json
)

// OptionID identifies an option in a message.
type OptionID uint16

// Option IDs.
/*
   +-----+----+---+---+---+----------------+--------+--------+---------+
   | No. | C  | U | N | R | Name           | Format | Length | Default |
   +-----+----+---+---+---+----------------+--------+--------+---------+
   |   1 | x  |   |   | x | If-Match       | opaque | 0-8    | (none)  |
   |   3 | x  | x | - |   | Uri-Host       | string | 1-255  | (see    |
   |     |    |   |   |   |                |        |        | below)  |
   |   4 |    |   |   | x | ETag           | opaque | 1-8    | (none)  |
   |   5 | x  |   |   |   | If-None-Match  | empty  | 0      | (none)  |
   |   7 | x  | x | - |   | Uri-Port       | uint   | 0-2    | (see    |
   |     |    |   |   |   |                |        |        | below)  |
   |   8 |    |   |   | x | Location-Path  | string | 0-255  | (none)  |
   |  11 | x  | x | - | x | Uri-Path       | string | 0-255  | (none)  |
   |  12 |    |   |   |   | Content-Format | uint   | 0-2    | (none)  |
   |  14 |    | x | - |   | Max-Age        | uint   | 0-4    | 60      |
   |  15 | x  | x | - | x | Uri-Query      | string | 0-255  | (none)  |
   |  17 | x  |   |   |   | Accept         | uint   | 0-2    | (none)  |
   |  20 |    |   |   | x | Location-Query | string | 0-255  | (none)  |
   |  35 | x  | x | - |   | Proxy-Uri      | string | 1-1034 | (none)  |
   |  39 | x  | x | - |   | Proxy-Scheme   | string | 1-255  | (none)  |
   |  60 |    |   | x |   | Size1          | uint   | 0-4    | (none)  |
   +-----+----+---+---+---+----------------+--------+--------+---------+
*/
const (
	IfMatch       OptionID = 1
	URIHost       OptionID = 3
	ETag          OptionID = 4
	IfNoneMatch   OptionID = 5
	Observe       OptionID = 6
	URIPort       OptionID = 7
	LocationPath  OptionID = 8
	URIPath       OptionID = 11
	ContentFormat OptionID = 12
	MaxAge        OptionID = 14
	URIQuery      OptionID = 15
	Accept        OptionID = 17
	LocationQuery OptionID = 20
	ProxyURI      OptionID = 35
	ProxyScheme   OptionID = 39
	Size1         OptionID = 60
)

type OptionValueFormat int

const (
	UnknownValueFormat OptionValueFormat = iota
	EmptyValueFormat
	OpaqueValueFormat
	UintValueFormat
	StringValueFormat
)

type optionDef struct {
	name   string
	format OptionValueFormat
	minLen int
	maxLen int
}

var optionDefs = map[OptionID]optionDef{
	IfMatch:       optionDef{name: "If-Match", format: OpaqueValueFormat, minLen: 0, maxLen: 8},
	URIHost:       optionDef{name: "Uri-Host", format: StringValueFormat, minLen: 1, maxLen: 255},
	ETag:          optionDef{name: "ETag", format: OpaqueValueFormat, minLen: 1, maxLen: 8},
	IfNoneMatch:   optionDef{name: "If-None-Match", format: EmptyValueFormat, minLen: 0, maxLen: 0},
	Observe:       optionDef{name: "Observe", format: UintValueFormat, minLen: 0, maxLen: 3},
	URIPort:       optionDef{name: "Uri-Port", format: UintValueFormat, minLen: 0, maxLen: 2},
	LocationPath:  optionDef{name: "Location-Path", format: StringValueFormat, minLen: 0, maxLen: 255},
	URIPath:       optionDef{name: "Uri-Path", format: StringValueFormat, minLen: 0, maxLen: 255},
	ContentFormat: optionDef{name: "Content-Format", format: UintValueFormat, minLen: 0, maxLen: 2},
	MaxAge:        optionDef{name: "Max-Age", format: UintValueFormat, minLen: 0, maxLen: 4},
	URIQuery:      optionDef{name: "Uri-Query", format: StringValueFormat, minLen: 0, maxLen: 255},
	Accept:        optionDef{name: "Accept", format: UintValueFormat, minLen: 0, maxLen: 2},
	LocationQuery: optionDef{name: "Location-Query", format: StringValueFormat, minLen: 0, maxLen: 255},
	ProxyURI:      optionDef{name: "Proxy-Uri", format: StringValueFormat, minLen: 1, maxLen: 1034},
	ProxyScheme:   optionDef{name: "Proxy-Scheme", format: StringValueFormat, minLen: 1, maxLen: 255},
	Size1:         optionDef{name: "Size1", format: UintValueFormat, minLen: 0, maxLen: 4},
}

func SetOptionDef(id OptionID, name string, format OptionValueFormat, minLen, maxLen int) {
	if _, ok := optionDefs[id]; ok {
		panic(fmt.Errorf("%d option id def is seted", id))
	}
	optionDefs[id] = optionDef{name: name, format: format, minLen: minLen, maxLen: maxLen}
}

func (id OptionID) String() string {
	def := optionDefs[id]
	if len(def.name) > 0 {
		return fmt.Sprintf("%d(%s)", id, def.name)
	}
	return fmt.Sprintf("%d", id)
}

// 消息格式
/*
	|       0       |       1       |       2       |       3       |
	|7 6 5 4 3 2 1 0|7 6 5 4 3 2 1 0|7 6 5 4 3 2 1 0|7 6 5 4 3 2 1 0|
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|Ver| T |  TKL  |      Code     |          Message ID           |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|   Token (if any, TKL bytes) ...
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|   Options (if any) ...
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|1 1 1 1 1 1 1 1|    Payload (if any) ...
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

// option格式
/*
	 7   6   5   4   3   2   1   0
	+---------------+---------------+
	|               |               |
	|  Option Delta | Option Length |   1 byte
	|               |               |
	+---------------+---------------+
	\                               \
	/         Option Delta          /   0-2 bytes
	\          (extended)           \
	+-------------------------------+
	\                               \
	/         Option Length         /   0-2 bytes
	\          (extended)           \
	+-------------------------------+
	\                               \
	/                               /
	\                               \
	/         Option Value          /   0 or more bytes
	\                               \
	/                               /
	\                               \
	+-------------------------------+
*/

type Option struct {
	ID     OptionID
	Values []interface{}
}

type message struct {
	Type      Type
	Code      Code
	MessageID uint16
	Token     string
	Options   []Option
	Payload   []byte
}

type fixHeader struct {
	Flags     uint8
	Code      uint8
	MessageID uint16
}

func (m *message) String() string {
	return fmt.Sprintf("Type: %s, Code: %s, MessageID: %d, Token: %s", m.Type.String(), m.Code.String(), m.MessageID,
		base64.StdEncoding.EncodeToString([]byte(m.Token)))
}

func (m *message) Marshal() ([]byte, error) {
	var err error
	var buf bytes.Buffer

	// header
	h := fixHeader{
		Flags:     (1 << 6) | (uint8(m.Type) << 4) | 0x0f&uint8(len(m.Token)),
		Code:      uint8(m.Code),
		MessageID: m.MessageID,
	}
	if err = binary.Write(&buf, binary.BigEndian, h); err != nil {
		return nil, err
	}

	// token
	buf.WriteString(m.Token)

	// options
	sort.Slice(m.Options, func(i, j int) bool {
		if m.Options[i].ID == m.Options[j].ID {
			return i < j
		}
		return m.Options[i].ID < m.Options[j].ID
	})
	var prev OptionID
	enc := optionEncoder{w: &buf}
	for _, opt := range m.Options {
		for _, val := range opt.Values {
			data, err := optionValueToBytes(val)
			if err != nil {
				return nil, err
			}
			if err = enc.Encode(uint32(opt.ID-prev), data); err != nil {
				return nil, err
			}
			prev = opt.ID
		}
	}

	// payload
	if len(m.Payload) > 0 {
		buf.WriteByte(0xff)
		buf.Write(m.Payload)
	}

	return buf.Bytes(), nil
}

func (m *message) Unmarshal(data []byte) (err error) {
	if len(data) < 4 {
		return errors.New("short packet")
	}

	buf := bytes.NewBuffer(data)

	// header
	var h fixHeader
	if err = binary.Read(buf, binary.BigEndian, &h); err != nil {
		return err
	}
	m.Type = Type((h.Flags >> 4) & 0x3)
	m.Code = Code(h.Code)
	m.MessageID = h.MessageID

	// token
	tokenLen := int(h.Flags & 0x0f)
	if buf.Len() < tokenLen {
		return errors.New("truncated")
	}
	if tokenLen > 0 {
		token := make([]byte, tokenLen)
		if _, err = io.ReadFull(buf, token); err != nil {
			return err
		}
		m.Token = string(token)
	}

	// options
	var prev OptionID
	dec := optionDecoder{r: buf}
	for buf.Len() > 0 {
		flag, err := buf.ReadByte()
		if err != nil {
			return err
		}
		if flag == 0xff {
			break
		}
		delta, data, err := dec.Decode(flag)
		if err != nil {
			return err
		}
		id := prev + OptionID(delta)
		val := bytesToOptionValue(id, data)
		if val != nil {
			m.addOption(id, val)
		}
		prev = id
	}

	// payload
	if buf.Len() > 0 {
		m.Payload = make([]byte, buf.Len())
		if _, err = io.ReadFull(buf, m.Payload); err != nil {
			return err
		}
	}

	return nil
}

func (m *message) addOption(id OptionID, value interface{}) {
	i := len(m.Options) - 1
	if i >= 0 && m.Options[i].ID == id {
		m.Options[i].Values = append(m.Options[i].Values, value)
	} else {
		m.Options = append(m.Options, Option{ID: id, Values: []interface{}{value}})
	}
}

func encodeUint8(v uint8) []byte {
	b := make([]byte, 1)
	b[0] = v
	return b
}

func encodeUint16(v uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return b
}

func encodeUint24(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b[1:]
}

func encodeUint32(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

func encodeUintVariant(v uint32) []byte {
	switch {
	case v == 0:
		return nil
	case v < 256:
		return encodeUint8(uint8(v))
	case v < 65536:
		return encodeUint16(uint16(v))
	case v < 16777216:
		return encodeUint24(v)
	default:
		return encodeUint32(v)
	}
}

func decodeUintVariant(b []byte) uint32 {
	data := make([]byte, 4)
	copy(data[4-len(b):], b)
	return binary.BigEndian.Uint32(data)
}

func optionValueToBytes(v interface{}) ([]byte, error) {
	switch tv := v.(type) {
	case string:
		return []byte(tv), nil
	case []byte:
		return []byte(tv), nil
	}

	var u uint32
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32:
		u = uint32(rv.Int())

	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32:
		u = uint32(rv.Uint())

	default:
		return nil, fmt.Errorf("optionValueToBytes: unsupport type(%s)", rv.Type())
	}
	return encodeUintVariant(u), nil
}

func bytesToOptionValue(id OptionID, buf []byte) interface{} {
	def := optionDefs[id]
	if def.format == UnknownValueFormat {
		return nil
	}
	if l := len(buf); l < def.minLen || l > def.maxLen {
		return nil
	}

	switch def.format {
	case EmptyValueFormat, OpaqueValueFormat:
		return buf
	case UintValueFormat:
		i := decodeUintVariant(buf)
		if id == ContentFormat || id == Accept {
			return MediaType(i)
		} else {
			return i
		}
	case StringValueFormat:
		return string(buf)
	}
	return nil
}

type encodeWriter interface {
	io.Writer
	io.ByteWriter
}

type decodeReader interface {
	io.Reader
	io.ByteReader
}

type optionEncoder struct {
	w encodeWriter
}

func (e *optionEncoder) Encode(delta uint32, value []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
				return
			}
			panic(r)
		}
	}()

	length := uint32(len(value))
	high, de := e.encodeHeader(delta)
	low, le := e.encodeHeader(length)
	e.writeByte(high<<4 | low)
	e.write(de)
	e.write(le)
	e.write(value)
	return nil
}

func (e *optionEncoder) writeByte(b byte) {
	if err := e.w.WriteByte(b); err != nil {
		panic(err)
	}
}

func (e *optionEncoder) write(p []byte) {
	if len(p) <= 0 {
		return
	}
	if _, err := e.w.Write(p); err != nil {
		panic(err)
	}
}

func (e *optionEncoder) encodeHeader(h uint32) (uint8, []byte) {
	if h < 13 {
		return uint8(h), nil
	} else if h < 269 {
		return 13, e.encodeUint8(h - 13)
	} else if h < 269+65535 {
		return 14, e.encodeUint16(h - 269)
	}
	panic(fmt.Errorf("encode option: invalid header(%d)", h))
}

func (e *optionEncoder) encodeUint8(x uint32) []byte {
	b := make([]byte, 1)
	b[0] = uint8(x)
	return b
}

func (e *optionEncoder) encodeUint16(x uint32) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(x))
	return b
}

type optionDecoder struct {
	r decodeReader
}

func (d *optionDecoder) Decode(flag byte) (delta uint32, value []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
				return
			}
			panic(r)
		}
	}()

	low := uint32(flag & 0x0f)
	high := uint32(flag >> 4)
	delta = d.decodeHeader(high)
	length := d.decodeHeader(low)
	value = d.readValue(length)
	return delta, value, nil
}

func (d *optionDecoder) readValue(n uint32) []byte {
	if n <= 0 {
		return nil
	}
	value := make([]byte, n)
	if _, err := io.ReadFull(d.r, value); err != nil {
		panic(err)
	}
	return value
}

func (d *optionDecoder) decodeHeader(h uint32) uint32 {
	if h < 13 {
		return h
	} else if h == 13 {
		return 13 + d.decodeUint8()
	} else if h == 14 {
		return 269 + d.decodeUint16()
	}
	panic(fmt.Errorf("decode option: invalid header(%d)", h))
}

func (d *optionDecoder) decodeUint8() uint32 {
	x, err := d.r.ReadByte()
	if err != nil {
		panic(err)
	}
	return uint32(x)
}

func (d *optionDecoder) decodeUint16() uint32 {
	b := make([]byte, 2)
	if _, err := io.ReadFull(d.r, b); err != nil {
		panic(err)
	}
	x := binary.BigEndian.Uint16(b)
	return uint32(x)
}
