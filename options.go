package coap

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type Options []Option

func (options *Options) clone() Options {
	cloneOptions := make(Options, len(*options))
	for i, o := range *options {
		cloneOptions[i].ID = o.ID
		cloneOptions[i].Values = make([]interface{}, len(o.Values))
		copy(cloneOptions[i].Values, o.Values)
	}
	return cloneOptions
}

func (options *Options) Add(id OptionID, v interface{}) {
	for i := range *options {
		o := &(*options)[i]
		if o.ID == id {
			o.Values = append(o.Values, v)
			return
		}
	}
	*options = append(*options, Option{
		ID:     id,
		Values: []interface{}{v}},
	)
}

func (options *Options) Set(id OptionID, v interface{}) {
	for i := range *options {
		o := &(*options)[i]
		if o.ID == id {
			o.Values = []interface{}{v}
			return
		}
	}
	*options = append(*options, Option{
		ID:     id,
		Values: []interface{}{v},
	})
}

func (options *Options) Get(id OptionID) interface{} {
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

func (options *Options) Del(id OptionID) {
	var results Options
	for _, o := range *options {
		if o.ID != id {
			results = append(results, o)
		}
	}
	*options = results
}

func (options *Options) GetOption(id OptionID) (Option, bool) {
	for _, o := range *options {
		if o.ID == id {
			return o, true
		}
	}
	return Option{}, false
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
