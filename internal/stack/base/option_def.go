package base

import "fmt"

// option id
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

	+-----+---+---+---+---+---------+--------+--------+---------+
	| No. | C | U | N | R | Name    | Format | Length | Default |
	+-----+---+---+---+---+---------+--------+--------+---------+
	|   6 |   | x | - |   | Observe | uint   | 0-3 B  | (none)  |
	+-----+---+---+---+---+---------+--------+--------+---------+

	+-----+---+---+---+---+--------+--------+--------+---------+
	| No. | C | U | N | R | Name   | Format | Length | Default |
	+-----+---+---+---+---+--------+--------+--------+---------+
	|  23 | C | U | - | - | Block2 | uint   |    0-3 | (none)  |
	|  27 | C | U | - | - | Block1 | uint   |    0-3 | (none)  |
	+-----+---+---+---+---+--------+--------+--------+---------+

	+-----+---+---+---+---+-------+--------+--------+---------+
	| No. | C | U | N | R | Name  | Format | Length | Default |
	+-----+---+---+---+---+-------+--------+--------+---------+
	|  28 |   |   | x |   | Size2 | uint   |    0-4 | (none)  |
	+-----+---+---+---+---+-------+--------+--------+---------+

	C=Critical, U=Unsafe, N=No-Cache-Key, R=Repeatable
*/
const (
	IfMatch       = 1
	URIHost       = 3
	ETag          = 4
	IfNoneMatch   = 5
	URIPort       = 7
	LocationPath  = 8
	URIPath       = 11
	ContentFormat = 12
	MaxAge        = 14
	URIQuery      = 15
	Accept        = 17
	LocationQuery = 20
	ProxyURI      = 35
	ProxyScheme   = 39
	Size1         = 60

	Observe = 6
	Block2  = 23
	Block1  = 27
	Size2   = 28
)

// option format
const (
	EmptyValue = iota
	UintValue
	StringValue
	OpaqueValue
)

type optionDef struct {
	id     uint16
	name   string
	format int
	repeat int
	minlen int
	maxlen int
}

var optionDefs = make(map[uint16]optionDef)

// RegisterOptionDef 注册选项定义.
//
// repeat参数定义了一个消息最多可包含多少个该选项, <=0则不做限制.
//
// 若重复注册同一编号的选项定义则会引发panic.
func RegisterOptionDef(id uint16, repeat int, name string, format, minlen, maxlen int) {
	if _, ok := optionDefs[id]; ok {
		panic("option registered")
	}
	optionDefs[id] = optionDef{
		id:     id,
		name:   name,
		format: format,
		repeat: repeat,
		minlen: minlen,
		maxlen: maxlen,
	}
}

// OptionName 返回选项名称.
func OptionName(id uint16) string {
	if def, ok := optionDefs[id]; ok && def.name != "" {
		return def.name
	}
	return fmt.Sprint(id)
}

func critical(id uint16) bool {
	return (id & 0x1) == 1
}

func recognize(id uint16, buf []byte, repeat int) bool {
	def, ok := optionDefs[id]
	if !ok {
		return false
	}
	if n := len(buf); n < def.minlen || n > def.maxlen {
		return false
	}
	if def.repeat > 0 && repeat > def.repeat {
		return false
	}
	return true
}

func init() {
	RegisterOptionDef(IfMatch, 0, "If-Match", OpaqueValue, 0, 8)
	RegisterOptionDef(URIHost, 1, "Uri-Host", StringValue, 1, 255)
	RegisterOptionDef(ETag, 0, "ETag", OpaqueValue, 1, 8)
	RegisterOptionDef(IfNoneMatch, 1, "If-None-Match", EmptyValue, 0, 0)
	RegisterOptionDef(URIPort, 1, "Uri-Port", UintValue, 0, 2)
	RegisterOptionDef(LocationPath, 0, "Location-Path", StringValue, 0, 255)
	RegisterOptionDef(URIPath, 0, "Uri-Path", StringValue, 0, 255)
	RegisterOptionDef(ContentFormat, 1, "Content-Format", UintValue, 0, 2)
	RegisterOptionDef(MaxAge, 1, "Max-Age", UintValue, 0, 4)
	RegisterOptionDef(URIQuery, 0, "Uri-Query", StringValue, 0, 255)
	RegisterOptionDef(Accept, 1, "Accept", UintValue, 0, 2)
	RegisterOptionDef(LocationQuery, 0, "Location-Query", StringValue, 0, 255)
	RegisterOptionDef(ProxyURI, 1, "Proxy-Uri", StringValue, 1, 1034)
	RegisterOptionDef(ProxyScheme, 1, "Proxy-Scheme", StringValue, 1, 255)
	RegisterOptionDef(Size1, 1, "Size1", UintValue, 0, 4)
	RegisterOptionDef(Observe, 1, "Observe", UintValue, 0, 3)
	RegisterOptionDef(Block2, 1, "Block2", UintValue, 0, 3)
	RegisterOptionDef(Block1, 1, "Block1", UintValue, 0, 3)
	RegisterOptionDef(Size2, 1, "Size2", UintValue, 0, 4)
}
