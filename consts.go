package coap

import "github.com/ironzhang/coap/message"

const (
	CON = message.CON
	NON = message.NON
	ACK = message.ACK
	RST = message.RST
)

const (
	// Request Codes
	GET    = message.GET
	POST   = message.POST
	PUT    = message.PUT
	DELETE = message.DELETE

	// Responses Codes
	Created               = message.Created
	Deleted               = message.Deleted
	Valid                 = message.Valid
	Changed               = message.Changed
	Content               = message.Content
	BadRequest            = message.BadRequest
	Unauthorized          = message.Unauthorized
	BadOption             = message.BadOption
	Forbidden             = message.Forbidden
	NotFound              = message.NotFound
	MethodNotAllowed      = message.MethodNotAllowed
	NotAcceptable         = message.NotAcceptable
	PreconditionFailed    = message.PreconditionFailed
	RequestEntityTooLarge = message.RequestEntityTooLarge
	UnsupportedMediaType  = message.UnsupportedMediaType
	InternalServerError   = message.InternalServerError
	NotImplemented        = message.NotImplemented
	BadGateway            = message.BadGateway
	ServiceUnavailable    = message.ServiceUnavailable
	GatewayTimeout        = message.GatewayTimeout
	ProxyingNotSupported  = message.ProxyingNotSupported
)

const (
	TextPlain     = message.TextPlain
	AppLinkFormat = message.AppLinkFormat
	AppXML        = message.AppXML
	AppOctets     = message.AppOctets
	AppExi        = message.AppExi
	AppJSON       = message.AppJSON
)

const (
	IfMatch       = message.IfMatch
	URIHost       = message.URIHost
	ETag          = message.ETag
	IfNoneMatch   = message.IfNoneMatch
	Observe       = message.Observe
	URIPort       = message.URIPort
	LocationPath  = message.LocationPath
	URIPath       = message.URIPath
	ContentFormat = message.ContentFormat
	MaxAge        = message.MaxAge
	URIQuery      = message.URIQuery
	Accept        = message.Accept
	LocationQuery = message.LocationQuery
	ProxyURI      = message.ProxyURI
	ProxyScheme   = message.ProxyScheme
	Size1         = message.Size1
)
