package coap

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ironzhang/coap/internal/stack/base"
)

func requestKey(req *Request) string {
	return fmt.Sprintf("%s %s", req.Method.String(), req.URL.String())
}

func isCacheStatus(status Code) bool {
	if status == Content {
		return true
	}
	c := status >> 5
	if c == 4 || c == 5 {
		return true
	}
	return false
}

func cloneOptionsExclude(src Options, exclude func(o base.Option) bool) Options {
	dst := make(Options, 0, len(src))
	for _, o := range src {
		if exclude(o) {
			continue
		}
		dst = append(dst, o)
	}
	return dst
}

func optionsEqual(x, y Options) bool {
	x = cloneOptionsExclude(x, func(o base.Option) bool { return base.NoCacheKey(o.ID) })
	y = cloneOptionsExclude(y, func(o base.Option) bool { return base.NoCacheKey(o.ID) })
	return reflect.DeepEqual(x, y)
}

type cvalue struct {
	req     *Request
	resp    *Response
	start   time.Time
	timeout time.Duration
}

type cache struct {
	mu     sync.Mutex
	values map[string]cvalue
}

func (c *cache) Get(req *Request) (*Response, bool) {
	key := requestKey(req)
	value, ok := c.getValue(key)
	if !ok {
		return nil, false
	}
	if !optionsEqual(req.Options, value.req.Options) {
		return nil, false
	}
	return value.resp, true
}

func (c *cache) Add(req *Request, resp *Response) {
	if isCacheStatus(resp.Status) {
		key := requestKey(req)
		age, ok := resp.Options.Get(MaxAge).(uint32)
		if !ok {
			age = 60
		}
		c.addValue(key, cvalue{
			req:     req,
			resp:    resp,
			start:   time.Now(),
			timeout: time.Duration(age) * time.Second,
		})
	}
}

func (c *cache) addValue(key string, value cvalue) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if value.timeout > 0 {
		if c.values == nil {
			c.values = make(map[string]cvalue)
		}
		c.values[key] = value
	} else {
		delete(c.values, key)
	}
}

func (c *cache) getValue(key string) (cvalue, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.values[key]
	if !ok {
		return cvalue{}, false
	}
	if time.Since(value.start) > value.timeout {
		delete(c.values, key)
		return cvalue{}, false
	}
	return value, true
}
