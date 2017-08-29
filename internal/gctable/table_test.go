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

func MakeTestObjects(n int, timeout time.Duration) []Object {
	objects := make([]Object, 0, n)
	for i := 0; i < n; i++ {
		objects = append(objects, NewTestObject(fmt.Sprint(i), timeout))
	}
	return objects
}

func BucketAddObjects(b *bucket, objects []Object) {
	var wg sync.WaitGroup
	for _, o := range objects {
		wg.Add(1)
		go func(o Object) {
			b.add(o)
			wg.Done()
		}(o)
	}
	wg.Wait()
}

func BucketRemoveObjects(b *bucket, objects []Object) {
	var wg sync.WaitGroup
	for _, o := range objects {
		wg.Add(1)
		go func(o Object) {
			b.remove(o.Key())
			wg.Done()
		}(o)
	}
	wg.Wait()
}

func BucketGetObjects(b *bucket, objects []Object) int {
	var count int64
	var wg sync.WaitGroup
	for _, o := range objects {
		wg.Add(1)
		go func(o Object) {
			_, ok := b.get(o.Key())
			if ok {
				atomic.AddInt64(&count, 1)
			}
			wg.Done()
		}(o)
	}
	wg.Wait()
	return int(count)
}

func TestBucketAdd(t *testing.T) {
	var n = 10000
	var b bucket
	var objects = MakeTestObjects(n, time.Second)
	BucketAddObjects(&b, objects)
	if got, want := len(b.m), n; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TestBucketRemove(t *testing.T) {
	var n = 10000
	var b bucket
	var objects = MakeTestObjects(n, time.Second)
	BucketAddObjects(&b, objects)
	BucketRemoveObjects(&b, objects)
	if got, want := len(b.m), 0; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TestBucketGet(t *testing.T) {
	var n = 10000
	var b bucket
	var objects = MakeTestObjects(n, time.Second)
	BucketAddObjects(&b, objects)
	count := BucketGetObjects(&b, objects)
	if got, want := count, n; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TestBucketPerformGC(t *testing.T) {
	var n = 10000
	var b bucket
	var objects = MakeTestObjects(n, time.Second/2)
	BucketAddObjects(&b, objects)
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
	var objects = MakeTestObjects(n, time.Second/2)
	BucketAddObjects(&b, objects)
	time.Sleep(time.Second)
	b.gc()
	if got, want := len(b.m), 0; got != want {
		t.Errorf("object num: %d != %d", got, want)
	}
}

func TableAddObjects(t *Table, objects []Object) int {
	var count int64
	var wg sync.WaitGroup
	for _, o := range objects {
		wg.Add(1)
		go func(o Object) {
			if err := t.Add(o); err == nil {
				atomic.AddInt64(&count, 1)
			}
			wg.Done()
		}(o)
	}
	wg.Wait()
	return int(count)
}

func TableGetObjects(t *Table, objects []Object) int {
	var count int64
	var wg sync.WaitGroup
	for _, o := range objects {
		wg.Add(1)
		go func(o Object) {
			if _, ok := t.Get(o.Key()); ok {
				atomic.AddInt64(&count, 1)
			}
			wg.Done()
		}(o)
	}
	wg.Wait()
	return int(count)
}

func TestTable(t *testing.T) {
	previous := SetGC(time.Second)
	defer SetGC(previous)

	var tb Table
	var n = 10000
	var objects = MakeTestObjects(n, time.Second+500*time.Millisecond)
	if got, want := TableAddObjects(&tb, objects), n; got != want {
		t.Errorf("table add objects: %v != %v", got, want)
	}
	if got, want := TableGetObjects(&tb, objects), n; got != want {
		t.Errorf("table get objects: %v != %v", got, want)
	}
	time.Sleep(time.Second + 600*time.Millisecond)
	if got, want := TableGetObjects(&tb, objects), 0; got != want {
		t.Errorf("table get objects: %v != %v", got, want)
	}
}
