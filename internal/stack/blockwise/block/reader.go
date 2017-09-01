package block

import (
	"errors"
	"io"
)

func NewReader(buf []byte) *Reader {
	return new(Reader).Init(buf)
}

type Reader struct {
	seq uint32
	off uint32
	buf []byte
}

func (r *Reader) Init(buf []byte) *Reader {
	r.seq = 0
	r.off = 0
	r.buf = buf
	return r
}

func (r *Reader) Read(seq uint32, size uint32) (Option, []byte, error) {
	if r.off >= uint32(len(r.buf)) {
		return Option{}, nil, io.EOF
	}
	if r.seq != seq {
		return Option{}, nil, errors.New("sequence confusion")
	}
	r.seq = r.off/size + 1

	start := r.off
	remaining := uint32(len(r.buf[start:]))
	if size < remaining {
		r.off += size
	} else {
		r.off += remaining
	}

	opt := Option{
		Num:  r.seq - 1,
		More: r.off < uint32(len(r.buf)),
		Size: size,
	}
	return opt, r.buf[start:r.off], nil
}
