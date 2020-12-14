package coap

import "github.com/ironzhang/coap/internal/stack/base"

// Code COAP消息码
type Code uint8

func (c Code) String() string {
	return base.CodeName(uint8(c))
}

// Request Codes
const (
	GET    Code = base.GET
	POST   Code = base.POST
	PUT    Code = base.PUT
	DELETE Code = base.DELETE
)

// Responses Codes
const (
	Created  Code = base.Created
	Deleted  Code = base.Deleted
	Valid    Code = base.Valid
	Changed  Code = base.Changed
	Content  Code = base.Content
	Continue Code = base.Continue

	BadRequest               Code = base.BadRequest
	Unauthorized             Code = base.Unauthorized
	BadOption                Code = base.BadOption
	Forbidden                Code = base.Forbidden
	NotFound                 Code = base.NotFound
	MethodNotAllowed         Code = base.MethodNotAllowed
	NotAcceptable            Code = base.NotAcceptable
	RequestEntityIncomplete  Code = base.RequestEntityIncomplete
	PreconditionFailed       Code = base.PreconditionFailed
	RequestEntityTooLarge    Code = base.RequestEntityTooLarge
	UnsupportedContentFormat Code = base.UnsupportedContentFormat

	InternalServerError  Code = base.InternalServerError
	NotImplemented       Code = base.NotImplemented
	BadGateway           Code = base.BadGateway
	ServiceUnavailable   Code = base.ServiceUnavailable
	GatewayTimeout       Code = base.GatewayTimeout
	ProxyingNotSupported Code = base.ProxyingNotSupported
)

// Option IDs
const (
	IfMatch       = base.IfMatch
	URIHost       = base.URIHost
	ETag          = base.ETag
	IfNoneMatch   = base.IfNoneMatch
	Observe       = base.Observe
	URIPort       = base.URIPort
	LocationPath  = base.LocationPath
	URIPath       = base.URIPath
	ContentFormat = base.ContentFormat
	MaxAge        = base.MaxAge
	URIQuery      = base.URIQuery
	Accept        = base.Accept
	LocationQuery = base.LocationQuery
	Block2        = base.Block2
	Block1        = base.Block1
	Size2         = base.Size2
	ProxyURI      = base.ProxyURI
	ProxyScheme   = base.ProxyScheme
	Size1         = base.Size1
)

// Content类型定义
const (
	TextPlain     = uint32(0)  // text/plain;charset=utf-8
	AppLinkFormat = uint32(40) // application/link-format
	AppXML        = uint32(41) // application/xml
	AppOctets     = uint32(42) // application/octet-stream
	AppExi        = uint32(47) // application/exi
	AppJSON       = uint32(50) // application/json
)

// Token 消息令牌
type Token string

func (t Token) String() string {
	return base.TokenString(string(t))
}

const (
	EmptyValue  = base.EmptyValue
	UintValue   = base.UintValue
	StringValue = base.StringValue
	OpaqueValue = base.OpaqueValue
)

// RegisterOptionDef 注册选项定义.
//
// repeat参数定义了一个消息最多可包含多少个该选项, <=0则不做限制.
//
// 若重复注册同一编号的选项定义则会引发panic.
func RegisterOptionDef(id uint16, repeat int, name string, format, minlen, maxlen int) {
	base.RegisterOptionDef(id, repeat, name, format, minlen, maxlen)
}
