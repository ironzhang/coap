package block

import "github.com/ironzhang/coap/internal/stack/base"

const (
	szxMask  = 0x07
	moreMask = 1 << 3
)

func ParseBlock1Option(m base.Message) (Option, bool) {
	return getBlockOption(m, base.Block1)
}

func ParseBlock2Option(m base.Message) (Option, bool) {
	return getBlockOption(m, base.Block2)
}

func getBlockOption(m base.Message, id uint16) (Option, bool) {
	v := m.GetOption(id)
	if v == nil {
		return Option{}, false
	}
	return ParseOption(v.(uint32)), true
}

func ParseOption(value uint32) Option {
	return Option{
		Num:  value >> 4,
		More: (value & moreMask) == moreMask,
		Size: exponentToBlockSize(value & szxMask),
	}
}

type Option struct {
	Num  uint32
	More bool
	Size uint32
}

func (o Option) Value() uint32 {
	value := o.Num << 4
	if o.More {
		value |= moreMask
	}
	value |= blockSizeToExponent(o.Size)
	return value
}

func blockSizeToExponent(size uint32) uint32 {
	switch size {
	case 16:
		return 0
	case 32:
		return 1
	case 64:
		return 2
	case 128:
		return 3
	case 256:
		return 4
	case 512:
		return 5
	case 1024:
		return 6
	default:
		return 6
	}
}

func exponentToBlockSize(exp uint32) uint32 {
	switch exp {
	case 0:
		return 16
	case 1:
		return 32
	case 2:
		return 64
	case 3:
		return 128
	case 4:
		return 256
	case 5:
		return 512
	case 6:
		return 1024
	default:
		return 1024
	}
}

func FixBlockSize(size uint32) uint32 {
	if size < 32 {
		return 16
	} else if size < 64 {
		return 32
	} else if size < 128 {
		return 64
	} else if size < 256 {
		return 128
	} else if size < 512 {
		return 256
	} else if size < 1024 {
		return 512
	} else {
		return 1024
	}
}
