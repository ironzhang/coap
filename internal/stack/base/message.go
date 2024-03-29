// the file is borrowed from github.com/dustin/go-coap/message.go

package base

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
)

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

// 消息类型
const (
	CON = 0
	NON = 1
	ACK = 2
	RST = 3
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

func TypeName(t uint8) string {
	return typeNames[t]
}

// Request Codes
const (
	GET    = 0<<5 | 1
	POST   = 0<<5 | 2
	PUT    = 0<<5 | 3
	DELETE = 0<<5 | 4
)

// Responses Codes
const (
	Created  = 2<<5 | 1
	Deleted  = 2<<5 | 2
	Valid    = 2<<5 | 3
	Changed  = 2<<5 | 4
	Content  = 2<<5 | 5
	Continue = 2<<5 | 31

	BadRequest               = 4<<5 | 0
	Unauthorized             = 4<<5 | 1
	BadOption                = 4<<5 | 2
	Forbidden                = 4<<5 | 3
	NotFound                 = 4<<5 | 4
	MethodNotAllowed         = 4<<5 | 5
	NotAcceptable            = 4<<5 | 6
	RequestEntityIncomplete  = 4<<5 | 8
	PreconditionFailed       = 4<<5 | 12
	RequestEntityTooLarge    = 4<<5 | 13
	UnsupportedContentFormat = 4<<5 | 15

	InternalServerError  = 5<<5 | 0
	NotImplemented       = 5<<5 | 1
	BadGateway           = 5<<5 | 2
	ServiceUnavailable   = 5<<5 | 3
	GatewayTimeout       = 5<<5 | 4
	ProxyingNotSupported = 5<<5 | 5
)

var codeNames = [256]string{
	GET:                      "GET",
	POST:                     "POST",
	PUT:                      "PUT",
	DELETE:                   "DELETE",
	Created:                  "Created",
	Deleted:                  "Deleted",
	Valid:                    "Valid",
	Changed:                  "Changed",
	Content:                  "Content",
	Continue:                 "Continue",
	BadRequest:               "BadRequest",
	Unauthorized:             "Unauthorized",
	BadOption:                "BadOption",
	Forbidden:                "Forbidden",
	NotFound:                 "NotFound",
	MethodNotAllowed:         "MethodNotAllowed",
	NotAcceptable:            "NotAcceptable",
	RequestEntityIncomplete:  "RequestEntityIncomplete",
	PreconditionFailed:       "PreconditionFailed",
	RequestEntityTooLarge:    "RequestEntityTooLarge",
	UnsupportedContentFormat: "UnsupportedContentFormat",
	InternalServerError:      "InternalServerError",
	NotImplemented:           "NotImplemented",
	BadGateway:               "BadGateway",
	ServiceUnavailable:       "ServiceUnavailable",
	GatewayTimeout:           "GatewayTimeout",
	ProxyingNotSupported:     "ProxyingNotSupported",
}

func init() {
	for i := range codeNames {
		if codeNames[i] == "" {
			c1, c2 := i>>5, i&0x1f
			codeNames[i] = fmt.Sprintf("Unknown (%d.%02d)", c1, c2)
		}
	}
}

func CodeName(c uint8) string {
	return codeNames[c]
}

type fixHeader struct {
	Flags     uint8
	Code      uint8
	MessageID uint16
}

// Option COAP消息选项
type Option struct {
	ID    uint16
	Value interface{}
}

var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")

// WriteOptions 以可读形式输出选项
func WriteOptions(w io.Writer, options []Option) {
	sort.Slice(options, func(i, j int) bool {
		if options[i].ID == options[j].ID {
			return i < j
		}
		return options[i].ID < options[j].ID
	})

	for _, o := range options {
		s, ok := o.Value.(string)
		if ok {
			s = headerNewlineToSpace.Replace(s)
			fmt.Fprintf(w, "%s: %s\r\n", OptionName(o.ID), s)
		} else {
			fmt.Fprintf(w, "%s: %v\r\n", OptionName(o.ID), o.Value)
		}
	}
}

// TokenString 返回可读的token字符串
func TokenString(token string) string {
	var buf bytes.Buffer
	for _, b := range []byte(token) {
		fmt.Fprintf(&buf, "%02x", b)
	}
	return buf.String()
}

// Message COAP消息
type Message struct {
	Type      uint8
	Code      uint8
	MessageID uint16
	Token     string
	Options   []Option
	Payload   []byte
}

func (m Message) String() string {
	if len(m.Token) <= 0 {
		return fmt.Sprintf("%s,%s,%d", TypeName(m.Type), CodeName(m.Code), m.MessageID)
	}
	return fmt.Sprintf("%s,%s,%d,%s", TypeName(m.Type), CodeName(m.Code), m.MessageID, TokenString(m.Token))
}

func (m *Message) AddOption(id uint16, v interface{}) {
	m.Options = append(m.Options, Option{ID: id, Value: v})
}

func (m *Message) DelOption(id uint16) {
	options := make([]Option, 0, len(m.Options))
	for _, o := range m.Options {
		if o.ID != id {
			options = append(options, o)
		}
	}
	m.Options = options
}

func (m *Message) SetOption(id uint16, v interface{}) {
	m.DelOption(id)
	m.AddOption(id, v)
}

func (m *Message) GetOption(id uint16) interface{} {
	for _, o := range m.Options {
		if o.ID == id {
			return o.Value
		}
	}
	return nil
}

func (m *Message) GetOptions(id uint16) (values []interface{}) {
	for _, o := range m.Options {
		if o.ID == id {
			values = append(values, o.Value)
		}
	}
	return values
}

func (m *Message) Marshal() ([]byte, error) {
	var err error
	var buf bytes.Buffer

	// header
	h := fixHeader{
		Flags:     1<<6 | m.Type<<4 | 0x0f&uint8(len(m.Token)),
		Code:      m.Code,
		MessageID: m.MessageID,
	}
	if err = binary.Write(&buf, binary.BigEndian, h); err != nil {
		return nil, err
	}

	// token
	if len(m.Token) > 8 {
		return nil, errors.New("invalid token")
	}
	buf.WriteString(m.Token)

	// options
	sort.Slice(m.Options, func(i, j int) bool {
		if m.Options[i].ID == m.Options[j].ID {
			return i < j
		}
		return m.Options[i].ID < m.Options[j].ID
	})
	var prev uint16
	enc := optionEncoder{w: &buf}
	for _, opt := range m.Options {
		data, err := optionValueToBytes(opt.Value)
		if err != nil {
			return nil, err
		}
		if err = enc.Encode(uint32(opt.ID-prev), data); err != nil {
			return nil, err
		}
		prev = opt.ID
	}

	// payload
	if len(m.Payload) > 0 {
		buf.WriteByte(0xff)
		buf.Write(m.Payload)
	}

	return buf.Bytes(), nil
}

func (m *Message) Unmarshal(data []byte) (err error) {
	if len(data) < 4 {
		return errors.New("short packet")
	}

	buf := bytes.NewBuffer(data)

	// header
	var h fixHeader
	if err = binary.Read(buf, binary.BigEndian, &h); err != nil {
		return err
	}
	if version := h.Flags >> 6; version != 1 {
		return errors.New("invalid version")
	}
	m.Type = (h.Flags >> 4) & 0x3
	m.Code = h.Code
	m.MessageID = h.MessageID

	// token
	tokenLen := int(h.Flags & 0x0f)
	if tokenLen > 8 {
		return messageFormatError{"token length too long"}
	}
	if buf.Len() < tokenLen {
		return messageFormatError{"token truncated"}
	}
	if tokenLen > 0 {
		token := make([]byte, tokenLen)
		if _, err = io.ReadFull(buf, token); err != nil {
			return err
		}
		m.Token = string(token)
	}

	// options
	var id uint16
	var repeat int
	var unrecognized bool
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
		if delta == 0 {
			repeat++
		} else {
			repeat = 1
			id += uint16(delta)
		}

		if !recognize(id, data, repeat) {
			if !Critical(id) {
				continue
			}
			unrecognized = true
		}
		val := bytesToOptionValue(id, data)
		m.Options = append(m.Options, Option{ID: id, Value: val})
	}

	// payload
	if buf.Len() > 0 {
		m.Payload = make([]byte, buf.Len())
		if _, err = io.ReadFull(buf, m.Payload); err != nil {
			return err
		}
	}

	if unrecognized {
		return badOptionsError{`Unrecognized options of class "critical" that occur in a Confirmable request`}
	}
	return nil
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
	if v == nil {
		return nil, nil
	}

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

func bytesToOptionValue(id uint16, buf []byte) interface{} {
	format := OpaqueValue
	if def, ok := optionDefs[id]; ok {
		format = def.Format
	}
	switch format {
	case EmptyValue:
		return struct{}{}
	case UintValue:
		return decodeUintVariant(buf)
	case StringValue:
		return string(buf)
	case OpaqueValue:
		return buf
	default:
		return buf
	}
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

type MessageStringer struct {
	WritePayload func(w io.Writer, payload []byte)
}

func (p *MessageStringer) MessageString(m Message) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s\r\n", m.String())
	WriteOptions(&buf, m.Options)
	if p.WritePayload != nil {
		p.WritePayload(&buf, m.Payload)
	}
	return buf.String()
}
