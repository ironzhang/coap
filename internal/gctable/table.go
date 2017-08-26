package gctable

import (
	"hash/crc32"
	"sync"
)

type Object interface {
	Key() string
	CanGC() bool
	ExecuteGC()
}

type Table struct {
	mu      sync.Mutex
	buckets []bucket
}

func (t *Table) Add(object Object) {
	b := t.getBucket(object.Key())
	b.add(object)
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
		t.buckets = make([]bucket, 1024)
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
}

func (b *bucket) add(object Object) {
	b.mu.Lock()
	b.gc()
	if b.m == nil {
		b.m = make(map[string]Object)
		b.threshold = 10
	}
	b.m[object.Key()] = object
	b.mu.Unlock()
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
	if len(b.m) <= b.threshold {
		return
	}
	for key, object := range b.m {
		if object.CanGC() {
			delete(b.m, key)
			object.ExecuteGC()
		}
	}
	b.threshold = 2 * len(b.m)
	if b.threshold < 10 {
		b.threshold = 10
	}
}
