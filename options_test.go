package coap

import (
	"bytes"
	"reflect"
	"testing"
)

func OptionsString(o Options) string {
	var b bytes.Buffer
	o.Write(&b)
	return b.String()
}

func TestOptionsClone(t *testing.T) {
	tests := []struct {
		src Options
		dst Options
	}{
		{
			src: nil,
			dst: Options{},
		},
		{
			src: Options{},
			dst: Options{},
		},
		{
			src: Options{
				{ID: 1, Values: []interface{}{0}},
			},
			dst: Options{
				{ID: 1, Values: []interface{}{0}},
			},
		},
		{
			src: Options{
				{ID: 0, Values: []interface{}{0}},
				{ID: 1, Values: []interface{}{1, 1, 0}},
				{ID: 2, Values: []interface{}{0, 1}},
			},
			dst: Options{
				{ID: 0, Values: []interface{}{0}},
				{ID: 1, Values: []interface{}{1, 1, 0}},
				{ID: 2, Values: []interface{}{0, 1}},
			},
		},
	}
	for i, tt := range tests {
		if got, want := tt.src.clone(), tt.dst; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d:\ngot:\n%s\nwant:\n%s\n", i, OptionsString(got), OptionsString(want))
		}
	}
}

func TestOptionsAdd(t *testing.T) {
	datas := []struct {
		id  OptionID
		val interface{}
	}{
		{id: 0, val: 0},
		{id: 1, val: 1},
		{id: 2, val: 0},
		{id: 2, val: 1},
		{id: 1, val: 1},
		{id: 1, val: 0},
	}
	want := Options{
		{ID: 0, Values: []interface{}{0}},
		{ID: 1, Values: []interface{}{1, 1, 0}},
		{ID: 2, Values: []interface{}{0, 1}},
	}
	var got Options
	for _, data := range datas {
		got.Add(data.id, data.val)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:\n%s\nwant:\n%s\n", OptionsString(got), OptionsString(want))
	}
}

func TestOptionsSet(t *testing.T) {
	datas := []struct {
		id  OptionID
		val interface{}
	}{
		{id: 0, val: 0},
		{id: 1, val: 1},
		{id: 2, val: 0},
		{id: 2, val: 1},
	}
	want := Options{
		{ID: 0, Values: []interface{}{0}},
		{ID: 1, Values: []interface{}{1}},
		{ID: 2, Values: []interface{}{1}},
	}
	var got Options
	for _, data := range datas {
		got.Set(data.id, data.val)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:\n%s\nwant:\n%s\n", OptionsString(got), OptionsString(want))
	}
}

func TestOptionsGet(t *testing.T) {
	options := Options{
		{ID: 0, Values: []interface{}{0}},
		{ID: 1, Values: []interface{}{1}},
		{ID: 2, Values: []interface{}{0, 1}},
	}
	tests := []struct {
		id  OptionID
		val interface{}
	}{
		{id: 0, val: 0},
		{id: 1, val: 1},
		{id: 2, val: 0},
		{id: 2, val: 0},
		{id: 3, val: nil},
	}
	for i, tt := range tests {
		if got, want := options.Get(tt.id), tt.val; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d: %v != %v", i, got, want)
		}
	}
}

func TestOptionsDel(t *testing.T) {
	options := Options{
		{ID: 0, Values: []interface{}{0}},
		{ID: 1, Values: []interface{}{1}},
		{ID: 2, Values: []interface{}{0, 1}},
	}
	tests := []struct {
		ids     []OptionID
		options Options
	}{
		{
			ids: []OptionID{1},
			options: Options{
				{ID: 0, Values: []interface{}{0}},
				{ID: 2, Values: []interface{}{0, 1}},
			},
		},
		{
			ids: []OptionID{0, 2},
			options: Options{
				{ID: 1, Values: []interface{}{1}},
			},
		},
	}

	for i, tt := range tests {
		opts := options.clone()
		for _, id := range tt.ids {
			opts.Del(id)
		}
		if got, want := opts, tt.options; !reflect.DeepEqual(got, want) {
			t.Errorf("case%d:\ngot:\n%s\nwant:\n%s\n", i, OptionsString(got), OptionsString(want))
		}
	}
}

func TestOptionsGetOption(t *testing.T) {
	options := Options{
		{ID: 0, Values: []interface{}{0}},
		{ID: 1, Values: []interface{}{1}},
		{ID: 2, Values: []interface{}{0, 1}},
	}
	tests := []struct {
		id     OptionID
		ok     bool
		option Option
	}{
		{id: 0, ok: true, option: Option{ID: 0, Values: []interface{}{0}}},
		{id: 1, ok: true, option: Option{ID: 1, Values: []interface{}{1}}},
		{id: 2, ok: true, option: Option{ID: 2, Values: []interface{}{0, 1}}},
		{id: 3, ok: false},
	}
	for i, tt := range tests {
		option, ok := options.GetOption(tt.id)
		if got, want := ok, tt.ok; got != want {
			t.Errorf("case%d: ok: %v != %v", i, got, want)
		}
		if ok {
			if got, want := option, tt.option; !reflect.DeepEqual(got, want) {
				t.Errorf("case%d: option: %+v != %+v", i, got, want)
			}
		}
	}
}

func TestOptionsWrite(t *testing.T) {
	options := Options{
		{ID: 0, Values: []interface{}{0}},
		{ID: 1, Values: []interface{}{1}},
		{ID: 2, Values: []interface{}{0, 1}},
	}
	s := "0: 0\r\n1(If-Match): 1\r\n2: 0\r\n2: 1\r\n"

	var b bytes.Buffer
	options.Write(&b)
	if got, want := b.String(), s; got != want {
		t.Error("%q != %q", got, want)
	}
}
