package block1

import (
	"errors"
	"io"

	"github.com/ironzhang/coap/internal/stack/base"
)

type buffer struct {
	seq uint32
	off uint32
	buf []byte
}

func (r *buffer) Reset(buf []byte) *buffer {
	r.seq = 0
	r.off = 0
	r.buf = buf
	return r
}

func (r *buffer) Read(seq uint32, size uint32) (base.BlockOption, []byte, error) {
	if r.off >= uint32(len(r.buf)) {
		return base.BlockOption{}, nil, io.EOF
	}
	if r.seq != seq {
		return base.BlockOption{}, nil, errors.New("sequence confusion")
	}
	r.seq = r.off/size + 1

	start := r.off
	remaining := uint32(len(r.buf[start:]))
	if size < remaining {
		r.off += size
	} else {
		r.off += remaining
	}

	opt := base.BlockOption{
		Num:  r.seq - 1,
		More: r.off < uint32(len(r.buf)),
		Size: size,
	}
	return opt, r.buf[start:r.off], nil
}
