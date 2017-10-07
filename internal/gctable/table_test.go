package gctable

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type TestObject struct {
	key     string
	time    time.Time
	timeout time.Duration
	gc      bool
}

func NewTestObject(key string, timeout time.Duration) *TestObject {
	return &TestObject{
		key:     key,
		time:    time.Now(),
		timeout: timeout,
	}
}

func (o *TestObject) Key() string {
	return o.key
}

func (o *TestObject) CanGC() bool {
	return time.Since(o.time) > o.timeout
}

func (o *TestObject) ExecuteGC() {
	o.gc = true
}

func MakeTestKeys(n int) []string {
	keys := make([]string, 0, n)
	for i := 0; i < n; i++ {
		keys = append(keys, fmt.Sprint(i))
	}
	return keys
}

func BucketAddObjects(b *bucket, keys []string, timeout time.Duration) {
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			b.add(key, func() Object { return NewTestObject(key, timeout) })
			wg.Done()
		}(key)
	}
	wg.Wait()
}

func BucketRemoveObjects(b *bucket, keys []string) {
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			b.remove(key)
			wg.Done()
		}(key)
	}
	wg.Wait()
}

func BucketGetObjects(b *bucket, keys []string) int {
	var count int64
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			_, ok := b.get(key)
			if ok {
				atomic.AddInt64(&count, 1)
			}
			wg.Done()
		}(key)
	}
	wg.Wait()
	return int(count)
}

func TestBucketAdd(t *testing.T) {
	var n = 10000
	var b bucket
	var keys = MakeTestKeys(n)
	BucketAddObjects(&b, keys, time.Second)
	if got, want := len(b.m), n; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TestBucketRemove(t *testing.T) {
	var n = 10000
	var b bucket
	var keys = MakeTestKeys(n)
	BucketAddObjects(&b, keys, time.Second)
	BucketRemoveObjects(&b, keys)
	if got, want := len(b.m), 0; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TestBucketGet(t *testing.T) {
	var n = 10000
	var b bucket
	var keys = MakeTestKeys(n)
	BucketAddObjects(&b, keys, time.Second)
	count := BucketGetObjects(&b, keys)
	if got, want := count, n; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TestBucketPerformGC(t *testing.T) {
	var n = 10000
	var b bucket
	var keys = MakeTestKeys(n)
	BucketAddObjects(&b, keys, time.Second/2)
	time.Sleep(time.Second)
	b.performGC()
	if got, want := len(b.m), 0; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TestBucketGC(t *testing.T) {
	previous := SetGC(time.Second / 2)
	defer SetGC(previous)

	var n = 10000
	var b bucket
	var keys = MakeTestKeys(n)
	BucketAddObjects(&b, keys, time.Second/2)
	time.Sleep(time.Second)
	b.gc()
	if got, want := len(b.m), 0; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TableAddObjects(t *Table, keys []string, timeout time.Duration) int {
	var count int64
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			t.Add(key, func() Object { atomic.AddInt64(&count, 1); return NewTestObject(key, timeout) })
			wg.Done()
		}(key)
	}
	wg.Wait()
	return int(count)
}

func TableGetObjects(t *Table, keys []string) int {
	var count int64
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			if _, ok := t.Get(key); ok {
				atomic.AddInt64(&count, 1)
			}
			wg.Done()
		}(key)
	}
	wg.Wait()
	return int(count)
}

func TestTable(t *testing.T) {
	previous := SetGC(time.Second)
	defer SetGC(previous)

	var tb Table
	var n = 10000
	var keys = MakeTestKeys(n)
	if got, want := TableAddObjects(&tb, keys, time.Second+500*time.Millisecond), n; got != want {
		t.Errorf("table add objects: %v != %v", got, want)
	}
	if got, want := TableGetObjects(&tb, keys), n; got != want {
		t.Errorf("table get objects: %v != %v", got, want)
	}
	time.Sleep(time.Second + 600*time.Millisecond)
	if got, want := TableGetObjects(&tb, keys), 0; got != want {
		t.Errorf("table get objects: %v != %v", got, want)
	}
}
