package blockwise

import (
	"bytes"
	"errors"
)

type block struct {
	more    bool
	seq     uint32
	payload []byte
}

type buffer struct {
	seq uint32
	buf bytes.Buffer
}

func (p *buffer) ReadBlock(size uint32) (*block, error) {
	payload := make([]byte, size)
	n, err := p.buf.Read(payload)
	if err != nil {
		return nil, err
	}
	p.seq++
	return &block{
		more:    p.buf.Len() > 0,
		seq:     p.seq - 1,
		payload: payload[:n],
	}, nil
}

func (p *buffer) WriteBlock(b *block) error {
	if p.seq != b.seq {
		return errors.New("block message sequence confusion")
	}
	p.buf.Write(b.payload)
	p.seq++
	return nil
}

func (p *buffer) ReadPayload() []byte {
	payload := make([]byte, p.buf.Len())
	p.buf.Read(payload)
	p.Reset()
	return payload
}

func (p *buffer) WritePayload(payload []byte) {
	p.Reset()
	p.buf.Write(payload)
}

func (p *buffer) Reset() {
	p.seq = 0
	p.buf.Reset()
}

func (p *buffer) Len() int {
	return p.buf.Len()
}
