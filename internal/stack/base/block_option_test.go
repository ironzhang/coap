package base

import "testing"

func TestBlockSizeExponent(t *testing.T) {
	tests := []struct {
		exp  uint32
		size uint32
	}{
		{exp: 0, size: 16},
		{exp: 1, size: 32},
		{exp: 2, size: 64},
		{exp: 3, size: 128},
		{exp: 4, size: 256},
		{exp: 5, size: 512},
		{exp: 6, size: 1024},
	}
	for i, tt := range tests {
		if got, want := exponentToBlockSize(tt.exp), tt.size; got != want {
			t.Errorf("case%d: size: %v != %v", i, got, want)
		}
		if got, want := blockSizeToExponent(tt.size), tt.exp; got != want {
			t.Errorf("case%d: exponent: %v != %v", i, got, want)
		}
	}
}

func TestBlockOption(t *testing.T) {
	tests := []struct {
		val uint32
		opt BlockOption
	}{
		{val: 0x00, opt: BlockOption{Num: 0, More: false, Size: 16}},
		{val: 0x01, opt: BlockOption{Num: 0, More: false, Size: 32}},
		{val: 0x09, opt: BlockOption{Num: 0, More: true, Size: 32}},
		{val: 0x19, opt: BlockOption{Num: 1, More: true, Size: 32}},
		{val: 0x1e, opt: BlockOption{Num: 1, More: true, Size: 1024}},
	}
	for i, tt := range tests {
		if got, want := ParseBlockOption(tt.val), tt.opt; got != want {
			t.Errorf("case%d: option: %v != %v", i, got, want)
		}
		if got, want := tt.opt.Value(), tt.val; got != want {
			t.Errorf("case%d: value: %v != %v", i, got, want)
		}
	}
}
