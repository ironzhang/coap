package coap

import "github.com/ironzhang/coap/internal/stack/base"

type Code uint8

// Request Codes
const (
	GET    Code = base.GET
	POST   Code = base.POST
	PUT    Code = base.PUT
	DELETE Code = base.DELETE
)

// Responses Codes
const (
	Created               Code = base.Created
	Deleted               Code = base.Deleted
	Valid                 Code = base.Valid
	Changed               Code = base.Changed
	Content               Code = base.Content
	BadRequest            Code = base.BadRequest
	Unauthorized          Code = base.Unauthorized
	BadOption             Code = base.BadOption
	Forbidden             Code = base.Forbidden
	NotFound              Code = base.NotFound
	MethodNotAllowed      Code = base.MethodNotAllowed
	NotAcceptable         Code = base.NotAcceptable
	PreconditionFailed    Code = base.PreconditionFailed
	RequestEntityTooLarge Code = base.RequestEntityTooLarge
	UnsupportedMediaType  Code = base.UnsupportedMediaType
	InternalServerError   Code = base.InternalServerError
	NotImplemented        Code = base.NotImplemented
	BadGateway            Code = base.BadGateway
	ServiceUnavailable    Code = base.ServiceUnavailable
	GatewayTimeout        Code = base.GatewayTimeout
	ProxyingNotSupported  Code = base.ProxyingNotSupported
)

type OptionID uint16

func (id OptionID) String() string {
	return base.OptionName(uint16(id))
}

// Option IDs
const (
	IfMatch       OptionID = base.IfMatch
	URIHost       OptionID = base.URIHost
	ETag          OptionID = base.ETag
	IfNoneMatch   OptionID = base.IfNoneMatch
	Observe       OptionID = base.Observe
	URIPort       OptionID = base.URIPort
	LocationPath  OptionID = base.LocationPath
	URIPath       OptionID = base.URIPath
	ContentFormat OptionID = base.ContentFormat
	MaxAge        OptionID = base.MaxAge
	URIQuery      OptionID = base.URIQuery
	Accept        OptionID = base.Accept
	LocationQuery OptionID = base.LocationQuery
	ProxyURI      OptionID = base.ProxyURI
	ProxyScheme   OptionID = base.ProxyScheme
	Size1         OptionID = base.Size1
)
