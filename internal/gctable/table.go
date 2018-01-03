package gctable

import (
	"hash/crc32"
	"sync"
	"time"
)

var (
	bucketNum    = 10240
	minThreshold = 100
	gcInterval   = 10 * time.Minute
)

func SetGC(interval time.Duration) (previous time.Duration) {
	previous = gcInterval
	gcInterval = interval
	return
}

type Object interface {
	Key() string
	CanGC() bool
	ExecuteGC()
}

type Table struct {
	mu      sync.Mutex
	buckets []bucket
}

func (t *Table) Add(key string, new func() Object) Object {
	b := t.getBucket(key)
	return b.add(key, new)
}

func (t *Table) Get(key string) (Object, bool) {
	b := t.getBucket(key)
	return b.get(key)
}

func (t *Table) Remove(key string) {
	b := t.getBucket(key)
	b.remove(key)
}

func (t *Table) getBucket(key string) *bucket {
	t.mu.Lock()
	if t.buckets == nil {
		t.buckets = make([]bucket, bucketNum)
	}
	t.mu.Unlock()
	hash := crc32.ChecksumIEEE([]byte(key))
	index := hash % uint32(len(t.buckets))
	return &t.buckets[index]
}

type bucket struct {
	mu        sync.Mutex
	m         map[string]Object
	threshold int
	lastGC    time.Time
}

func (b *bucket) add(key string, alloc func() Object) Object {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.gc()
	if b.m == nil {
		b.m = make(map[string]Object)
		b.threshold = minThreshold
		b.lastGC = time.Now()
	}
	object, ok := b.m[key]
	if !ok {
		object = alloc()
		b.m[key] = object
	}
	return object
}

func (b *bucket) remove(key string) {
	b.mu.Lock()
	b.gc()
	if object, ok := b.m[key]; ok {
		delete(b.m, key)
		object.ExecuteGC()
	}
	b.mu.Unlock()
}

func (b *bucket) get(key string) (Object, bool) {
	b.mu.Lock()
	b.gc()
	object, ok := b.m[key]
	b.mu.Unlock()
	return object, ok
}

func (b *bucket) gc() {
	if len(b.m) <= b.threshold && time.Since(b.lastGC) < gcInterval {
		return
	}
	b.performGC()
	b.threshold = 2 * len(b.m)
	if b.threshold < minThreshold {
		b.threshold = minThreshold
	}
	b.lastGC = time.Now()
}

func (b *bucket) performGC() {
	for key, object := range b.m {
		if object.CanGC() {
			delete(b.m, key)
			object.ExecuteGC()
		}
	}
}
