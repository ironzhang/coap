package base

import "testing"

func TestOptionDefs(t *testing.T) {
	tests := []struct {
		id         uint16
		recognized bool
		repeat     bool
		format     int
		minlen     int
		maxlen     int
	}{
		{id: 0, recognized: false},
		{id: IfMatch, recognized: true, repeat: true, format: OpaqueValue, minlen: 0, maxlen: 8},
		{id: URIHost, recognized: true, repeat: false, format: StringValue, minlen: 1, maxlen: 255},
		{id: ETag, recognized: true, repeat: true, format: OpaqueValue, minlen: 1, maxlen: 8},
		{id: IfNoneMatch, recognized: true, repeat: false, format: EmptyValue, minlen: 0, maxlen: 0},
		{id: URIPort, recognized: true, repeat: false, format: UintValue, minlen: 0, maxlen: 2},
		{id: LocationPath, recognized: true, repeat: true, format: StringValue, minlen: 0, maxlen: 255},
	}
	for _, tt := range tests {
		def, ok := optionDefs[tt.id]
		if !tt.recognized {
			if ok {
				t.Errorf("%s option recognized, should be unrecognized", OptionName(tt.id))
			}
			continue
		}
		if !ok {
			t.Errorf("can not recognize %s option", OptionName(tt.id))
			continue
		}
		if got, want := def.id, tt.id; got != want {
			t.Errorf("%q option's id: %v != %v", OptionName(tt.id), got, want)
		}
		if got, want := (def.repeat != 1), tt.repeat; got != want {
			t.Errorf("%q option's repeatable: %v != %v", OptionName(tt.id), got, want)
		}
		if got, want := def.format, tt.format; got != want {
			t.Errorf("%q option's format: %v != %v", OptionName(tt.id), got, want)
		}
		if got, want := def.minlen, tt.minlen; got != want {
			t.Errorf("%q option's minlen: %v != %v", OptionName(tt.id), got, want)
		}
		if got, want := def.maxlen, tt.maxlen; got != want {
			t.Errorf("%q option's maxlen: %v != %v", OptionName(tt.id), got, want)
		}
	}
}
