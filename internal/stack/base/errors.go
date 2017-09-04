package base

import (
	"errors"
	"fmt"
)

var (
	ErrClientBusy        = errors.New("coap stack client busy")
	ErrServerBusy        = errors.New("coap stack server busy")
	ErrNoBlock1Option    = errors.New("no block1 option")
	ErrNoBlock2Option    = errors.New("no block2 option")
	ErrUnexpectMessageID = errors.New("unexpect message id")
)

type Error struct {
	Layer   string
	Cause   error
	Details string
}

func (e Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %v(%s)", e.Layer, e.Cause, e.Details)
	}
	return fmt.Sprintf("%s: %v", e.Layer, e.Cause)
}
