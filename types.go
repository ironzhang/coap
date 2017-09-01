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
