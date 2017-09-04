package base

import "io"

type BlockBuffer []byte

func (b BlockBuffer) Read(num, size uint32) (BlockOption, []byte, error) {
	blen := uint32(len(b))
	start := num * size
	if start >= blen {
		return BlockOption{}, nil, io.EOF
	}
	off := start + size
	if off > blen {
		off = blen
	}
	opt := BlockOption{
		Num:  num,
		More: off < blen,
		Size: size,
	}
	return opt, b[start:off], nil
}
