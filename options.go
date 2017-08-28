package coap

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ironzhang/coap/internal/stack/base"
)

// Options OptionID-Value键值对集合
type Options []base.Option

func (p *Options) clone() Options {
	c := make(Options, len(*p))
	copy(c, *p)
	return c
}

// Add 添加OptionID-Value键值对到Options.
func (p *Options) Add(id uint16, v interface{}) {
	*p = append(*p, base.Option{ID: id, Value: v})
}

// Del 删除指定选项.
func (p *Options) Del(id uint16) {
	var res Options
	for _, o := range *p {
		if o.ID != id {
			res = append(res, o)
		}
	}
	*p = res
}

// Set 设置指定选项的值.
// 该函数会替换指定选项的任何现有值.
func (p *Options) Set(id uint16, v interface{}) {
	p.Del(id)
	p.Add(id, v)
}

// Get 返回指定选项的第一个值.
// 若不包含指定选项则返回nil.
func (p *Options) Get(id uint16) interface{} {
	for _, o := range *p {
		if o.ID == id {
			return o.Value
		}
	}
	return nil
}

// GetValues 返回指定选项的所有值.
func (p *Options) GetValues(id uint16) []interface{} {
	var values []interface{}
	for _, o := range *p {
		if o.ID == id {
			values = append(values, o.Value)
		}
	}
	return values
}

// Contain 检查是否包含指定选项.
func (p *Options) Contain(id uint16) bool {
	for _, o := range *p {
		if o.ID == id {
			return true
		}
	}
	return false
}

var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")

// Write 以特定格式输出Options.
func (p *Options) Write(w io.Writer) error {
	sort.Slice(*p, func(i, j int) bool {
		if (*p)[i].ID == (*p)[j].ID {
			return i < j
		}
		return (*p)[i].ID < (*p)[j].ID
	})

	for _, o := range *p {
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

// SetStrings 设置设置指定选项的所有值, 以字符串数组形式写入.
func (p *Options) SetStrings(id uint16, ss []string) {
	p.Del(id)
	for _, s := range ss {
		p.Add(id, s)
	}
}

// GetStrings 获取指定选项的所有值, 以字符串数组形式返回.
func (p *Options) GetStrings(id uint16) []string {
	values := p.GetValues(id)
	ss := make([]string, 0, len(values))
	for _, v := range values {
		if s, ok := v.(string); ok {
			ss = append(ss, s)
		}
	}
	return ss
}

// SetPath 设置URIPath.
func (p *Options) SetPath(path string) {
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	if len(path) > 0 {
		p.SetStrings(URIPath, strings.Split(path, "/"))
	}
}

// GetPath 获取URIPath.
func (p *Options) GetPath() string {
	paths := p.GetStrings(URIPath)
	return strings.Join(paths, "/")
}

// SetQuery 设置URIQuery.
func (p *Options) SetQuery(query string) {
	if len(query) > 0 {
		p.SetStrings(URIQuery, strings.Split(query, "&"))
	}
}

// GetQuery 获取URIQuery.
func (p *Options) GetQuery() string {
	querys := p.GetStrings(URIQuery)
	return strings.Join(querys, "&")
}
