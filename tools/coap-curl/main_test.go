package main

import (
	"flag"
	"strings"
	"testing"
)

func TestStringValue(t *testing.T) {
	var v StringsValue
	v.Set("hello")
	v.Set("world")
	if got, want := v.String(), "hello,world"; got != want {
		t.Errorf("%q != %q", got, want)
	}
}

func TestStringValueFlag(t *testing.T) {
	tests := []struct {
		args []string
		v1   []string
		v2   []string
	}{
		{},
		{
			args: []string{"--v1", "1", "--v2", "2"},
			v1:   []string{"1"},
			v2:   []string{"2"},
		},
		{
			args: []string{"-v1", "1", "-v2", "2", "-v1", "3", "-v2", "4", "-v2", "5"},
			v1:   []string{"1", "3"},
			v2:   []string{"2", "4", "5"},
		},
	}
	for _, tt := range tests {
		var flagSet flag.FlagSet
		var v1, v2 StringsValue
		flagSet.Var(&v1, "v1", "value1")
		flagSet.Var(&v2, "v2", "value2")

		if err := flagSet.Parse(tt.args); err != nil {
			t.Fatalf("parse: %v", err)
		}
		if got, want := v1.String(), strings.Join(tt.v1, ","); got != want {
			t.Errorf("v1: %q != %q", got, want)
		}
		if got, want := v2.String(), strings.Join(tt.v2, ","); got != want {
			t.Errorf("v2: %q != %q", got, want)
		}
		t.Logf("v1: %q, v2: %q", v1.String(), v2.String())
	}
}
