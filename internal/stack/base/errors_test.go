package base

import (
	"io"
	"testing"
)

func TestError(t *testing.T) {
	tests := []struct {
		err Error
		str string
	}{
		{
			err: Error{Layer: "Layer", Cause: io.EOF},
			str: "Layer: EOF",
		},
		{
			err: Error{Layer: "Layer", Cause: io.EOF, Details: "read a.txt"},
			str: "Layer: EOF(read a.txt)",
		},
	}
	for i, tt := range tests {
		if got, want := tt.err.Error(), tt.str; got != want {
			t.Errorf("case%d: %q != %q", i, got, want)
		}
	}
}
