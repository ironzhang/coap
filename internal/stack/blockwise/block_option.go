package blockwise

const (
	szxMask  = 0x07
	moreMask = 1 << 3
)

type blockOption struct {
	num  uint32
	more bool
	size uint32
}

func parseBlockOption(value uint32) blockOption {
	return blockOption{
		num:  value >> 4,
		more: (value & moreMask) == moreMask,
		size: exponentToBlockSize(value & szxMask),
	}
}

func (o blockOption) Value() uint32 {
	value := o.num << 4
	if o.more {
		value |= moreMask
	}
	value |= blockSizeToExponent(o.size)
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

func fixBlockSize(size uint32) uint32 {
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
