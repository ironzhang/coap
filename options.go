package coap

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ironzhang/coap/message"
)

type Options []message.Option

func (options *Options) clone() Options {
	cloneOptions := make(Options, len(*options))
	for i, o := range *options {
		cloneOptions[i].ID = o.ID
		cloneOptions[i].Values = make([]interface{}, len(o.Values))
		copy(cloneOptions[i].Values, o.Values)
	}
	return cloneOptions
}

func (options *Options) Add(id message.OptionID, v interface{}) {
	for i := range *options {
		o := &(*options)[i]
		if o.ID == id {
			o.Values = append(o.Values, v)
			return
		}
	}
	*options = append(*options, message.Option{
		ID:     id,
		Values: []interface{}{v}},
	)
}

func (options *Options) Set(id message.OptionID, v interface{}) {
	for i := range *options {
		o := &(*options)[i]
		if o.ID == id {
			o.Values = []interface{}{v}
			return
		}
	}
	*options = append(*options, message.Option{
		ID:     id,
		Values: []interface{}{v},
	})
}

func (options *Options) Get(id message.OptionID) interface{} {
	for _, o := range *options {
		if o.ID == id {
			if len(o.Values) <= 0 {
				return nil
			}
			return o.Values[0]
		}
	}
	return nil
}

func (options *Options) Del(id message.OptionID) {
	var results Options
	for _, o := range *options {
		if o.ID != id {
			results = append(results, o)
		}
	}
	*options = results
}

func (options *Options) GetOption(id message.OptionID) (message.Option, bool) {
	for _, o := range *options {
		if o.ID == id {
			return o, true
		}
	}
	return message.Option{}, false
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
		for _, v := range o.Values {
			s, ok := v.(string)
			if ok {
				s = headerNewlineToSpace.Replace(s)
				fmt.Fprintf(w, "%s: %s\r\n", o.ID.String(), s)
			} else {
				fmt.Fprintf(w, "%s: %v\r\n", o.ID.String(), v)
			}
		}
	}
	return nil
}

func (options *Options) SetStrings(id message.OptionID, ss []string) {
	options.Del(id)
	for _, s := range ss {
		options.Add(id, s)
	}
}

func (options *Options) SetPath(path string) {
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	options.SetStrings(message.URIPath, strings.Split(path, "/"))
}
