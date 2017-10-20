package base

import (
	"errors"
	"fmt"
)

var (
	ErrNoBlock1Option = errors.New("no block1 option")
	ErrNoBlock2Option = errors.New("no block2 option")
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

var _ MessageFormatError = messageFormatError{}

type MessageFormatError interface {
	error
	FormatError() bool
}

type messageFormatError struct {
	err string
}

func (e messageFormatError) Error() string {
	return e.err
}

func (e messageFormatError) FormatError() bool {
	return true
}

var _ BadOptionsError = badOptionsError{}

type BadOptionsError interface {
	error
	BadOptions() bool
}

type badOptionsError struct {
	err string
}

func (e badOptionsError) Error() string {
	return e.err
}

func (e badOptionsError) BadOptions() bool {
	return true
}
