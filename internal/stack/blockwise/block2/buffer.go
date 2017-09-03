package block2

import (
	"errors"
	"io"

	"github.com/ironzhang/coap/internal/stack/blockwise/block"
)

type buffer struct {
	off uint32
	buf []byte
}

func (b *buffer) Reset(buf []byte) *buffer {
	b.off = 0
	b.buf = buf
	return b
}

func (b *buffer) Read(seq uint32, size uint32) (block.Option, []byte, error) {
	if b.off >= uint32(len(b.buf)) {
		return block.Option{}, nil, io.EOF
	}
	if seq*size != b.off {
		return block.Option{}, nil, errors.New("sequence confusion")
	}

	start := b.off
	remaining := uint32(len(b.buf[start:]))
	if size < remaining {
		b.off += size
	} else {
		b.off += remaining
	}

	opt := block.Option{
		Num:  seq,
		More: b.off < uint32(len(b.buf)),
		Size: size,
	}
	return opt, b.buf[start:b.off], nil
}
