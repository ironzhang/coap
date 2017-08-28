package coap

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ironzhang/coap/internal/stack/base"
)

type Options []base.Option

func (p *Options) clone() Options {
	c := make(Options, len(*p))
	copy(c, *p)
	return c
}

func (p *Options) Add(id uint16, v interface{}) {
	*p = append(*p, base.Option{ID: id, Value: v})
}

func (p *Options) Get(id uint16) interface{} {
	for _, o := range *p {
		if o.ID == id {
			return o.Value
		}
	}
	return nil
}

func (p *Options) Del(id uint16) {
	var res Options
	for _, o := range *p {
		if o.ID != id {
			res = append(res, o)
		}
	}
	*p = res
}

func (p *Options) Set(id uint16, v interface{}) {
	p.Del(id)
	p.Add(id, v)
}

func (p *Options) HasOption(id uint16) bool {
	for _, o := range *p {
		if o.ID == uint16(id) {
			return true
		}
	}
	return false
}

func (p *Options) GetValues(id uint16) []interface{} {
	var values []interface{}
	for _, o := range *p {
		if o.ID == id {
			values = append(values, o.Value)
		}
	}
	return values
}

var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")

func (options *Options) Write(w io.Writer) error {
	sort.Slice(*options, func(i, j int) bool {
		if (*options)[i].ID == (*options)[j].ID {
			return i < j
		}
		return (*options)[i].ID < (*options)[j].ID
	})

	for _, o := range *options {
		s, ok := o.Value.(string)
		if ok {
			s = headerNewlineToSpace.Replace(s)
			fmt.Fprintf(w, "%s: %s\r\n", base.OptionName(o.ID), s)
		} else {
			fmt.Fprintf(w, "%s: %v\r\n", base.OptionName(o.ID), o.Value)
		}
	}
	return nil
}

func (options *Options) SetStrings(id uint16, ss []string) {
	options.Del(id)
	for _, s := range ss {
		options.Add(id, s)
	}
}

func (options *Options) GetStrings(id uint16) []string {
	values := options.GetValues(id)
	ss := make([]string, 0, len(values))
	for _, v := range values {
		if s, ok := v.(string); ok {
			ss = append(ss, s)
		}
	}
	return ss
}

func (options *Options) SetPath(path string) {
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if len(path) > 0 {
		options.SetStrings(URIPath, strings.Split(path, "/"))
	}
}

func (options *Options) GetPath() string {
	paths := options.GetStrings(URIPath)
	return strings.Join(paths, "/")
}

func (options *Options) SetQuery(query string) {
	if len(query) > 0 {
		options.SetStrings(URIQuery, strings.Split(query, "&"))
	}
}

func (options *Options) GetQuery() string {
	querys := options.GetStrings(URIQuery)
	return strings.Join(querys, "&")
}
