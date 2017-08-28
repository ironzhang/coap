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
				{ID: 1, Value: 0},
			},
			dst: Options{
				{ID: 1, Value: 0},
			},
		},
		{
			src: Options{
				{ID: 0, Value: 0},
				{ID: 1, Value: 1},
				{ID: 1, Value: 1},
				{ID: 1, Value: 0},
				{ID: 2, Value: 0},
				{ID: 2, Value: 1},
			},
			dst: Options{
				{ID: 0, Value: 0},
				{ID: 1, Value: 1},
				{ID: 1, Value: 1},
				{ID: 1, Value: 0},
				{ID: 2, Value: 0},
				{ID: 2, Value: 1},
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
		id  uint16
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
		{ID: 0, Value: 0},
		{ID: 1, Value: 1},
		{ID: 2, Value: 0},
		{ID: 2, Value: 1},
		{ID: 1, Value: 1},
		{ID: 1, Value: 0},
	}
	var got Options
	for _, data := range datas {
		got.Add(data.id, data.val)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:\n%s\nwant:\n%s\n", OptionsString(got), OptionsString(want))
	}
}

func TestOptionsDel(t *testing.T) {
	options := Options{
		{ID: 0, Value: 0},
		{ID: 1, Value: 1},
		{ID: 2, Value: 0},
		{ID: 2, Value: 1},
	}
	tests := []struct {
		ids     []uint16
		options Options
	}{
		{
			ids: []uint16{1},
			options: Options{
				{ID: 0, Value: 0},
				{ID: 2, Value: 0},
				{ID: 2, Value: 1},
			},
		},
		{
			ids: []uint16{0, 2},
			options: Options{
				{ID: 1, Value: 1},
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

func TestOptionsSet(t *testing.T) {
	datas := []struct {
		id  uint16
		val interface{}
	}{
		{id: 0, val: 0},
		{id: 1, val: 1},
		{id: 2, val: 0},
		{id: 2, val: 1},
	}
	want := Options{
		{ID: 0, Value: 0},
		{ID: 1, Value: 1},
		{ID: 2, Value: 1},
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
		{ID: 0, Value: 0},
		{ID: 1, Value: 1},
		{ID: 2, Value: 0},
		{ID: 2, Value: 1},
	}
	tests := []struct {
		id  uint16
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

func TestOptionsContain(t *testing.T) {
	options := Options{
		{ID: 0, Value: 0},
		{ID: 1, Value: 1},
		{ID: 2, Value: 0},
		{ID: 2, Value: 1},
	}
	tests := []struct {
		id uint16
		ok bool
	}{
		{id: 0, ok: true},
		{id: 1, ok: true},
		{id: 2, ok: true},
		{id: 3, ok: false},
	}
	for i, tt := range tests {
		ok := options.Contain(tt.id)
		if got, want := ok, tt.ok; got != want {
			t.Errorf("case%d: ok: %v != %v", i, got, want)
		}
	}
}

func TestOptionsWrite(t *testing.T) {
	options := Options{
		{ID: 0, Value: 0},
		{ID: 1, Value: 1},
		{ID: 2, Value: 0},
		{ID: 2, Value: 1},
	}
	s := "0: 0\r\nIf-Match: 1\r\n2: 0\r\n2: 1\r\n"

	var b bytes.Buffer
	options.Write(&b)
	if got, want := b.String(), s; got != want {
		t.Errorf("%q != %q", got, want)
	}
}

func TestOptionsSetStrings(t *testing.T) {
	want := Options{
		{ID: 0, Value: "a"},
		{ID: 0, Value: "b"},
		{ID: 0, Value: "c"},
	}
	got := Options{}
	got.SetStrings(0, []string{"a", "b", "c"})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:\n%s\nwant:\n%s\n", OptionsString(got), OptionsString(want))
	}
}

func TestOptionsGetStrings(t *testing.T) {
	options := Options{
		{ID: 0, Value: "a"},
		{ID: 0, Value: "b"},
		{ID: 0, Value: "c"},
	}
	if got, want := options.GetStrings(0), []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, want)
	}
}

func TestOptionsSetPath(t *testing.T) {
	tests := []struct {
		path string
		want Options
	}{
		{
			path: "",
			want: nil,
		},
		{
			path: "/",
			want: nil,
		},
		{
			path: "a/b/c",
			want: Options{
				{URIPath, "a"},
				{URIPath, "b"},
				{URIPath, "c"},
			},
		},
		{
			path: "/a/b/c",
			want: Options{
				{URIPath, "a"},
				{URIPath, "b"},
				{URIPath, "c"},
			},
		},
	}
	for i, tt := range tests {
		var got Options
		got.SetPath(tt.path)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("case%d:\ngot:\n%s\nwant:\n%s\n", i, OptionsString(got), OptionsString(tt.want))
		}
	}
}

func TestOptionsPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{
			path: "",
			want: "",
		},
		{
			path: "/",
			want: "",
		},
		{
			path: "a/b/c",
			want: "a/b/c",
		},
		{
			path: "/a/b/c",
			want: "a/b/c",
		},
	}
	for i, tt := range tests {
		options := Options{}
		options.SetPath(tt.path)
		if got, want := options.GetPath(), tt.want; got != want {
			t.Errorf("case%d: %q != %q", i, got, want)
		}
	}
}

func TestOptionsSetQuery(t *testing.T) {
	tests := []struct {
		query string
		want  Options
	}{
		{
			query: "",
			want:  nil,
		},
		{
			query: "a=1&b=2&c=3",
			want: Options{
				{URIQuery, "a=1"},
				{URIQuery, "b=2"},
				{URIQuery, "c=3"},
			},
		},
	}
	for i, tt := range tests {
		var got Options
		got.SetQuery(tt.query)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("case%d:\ngot:\n%s\nwant:\n%s\n", i, OptionsString(got), OptionsString(tt.want))
		}
	}
}

func TestOptionsQuery(t *testing.T) {
	tests := []string{
		"",
		"a=1&b=2&c=3",
	}
	for i, query := range tests {
		options := Options{}
		options.SetQuery(query)
		if got, want := options.GetQuery(), query; got != want {
			t.Errorf("case%d: %q != %q", i, got, want)
		}
	}
}
