/*
Package timed_cache implements cache where items are stored for a certain period of time

BSD 2-Clause License

Copyright (c) 2021, Piotr Pszczółkowski
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
package timed_cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{})

// TimedCache implements timed cache
type TimedCache struct {
	sync.Mutex
	evictList   *list.List
	items       map[interface{}]*list.Element
	duration    int64
	onEvictCall EvictCallback
}

type entry struct {
	tm    int64
	key   interface{}
	value interface{}
}

// NewTimedCache constructs an LRU of the given size
func NewTimedCache(duration int64, onEvictCall EvictCallback) *TimedCache {
	c := &TimedCache{
		evictList:   list.New(),
		items:       make(map[interface{}]*list.Element),
		duration:    duration,
		onEvictCall: onEvictCall,
	}
	return c
}

// Purge is used to completely clear the cache.
func (c *TimedCache) Purge() {
	c.Lock()
	defer c.Unlock()

	for k, v := range c.items {
		if c.onEvictCall != nil {
			c.onEvictCall(k, v.Value.(*entry).value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *TimedCache) Add(key, value interface{}) bool {
	c.Lock()
	defer c.Unlock()

	c.purgeExpired()

	if _, ok := c.items[key]; ok {
		return false
	}

	ent := &entry{time.Now().Unix(), key, value}
	entry := c.evictList.PushFront(ent)
	c.items[key] = entry
	return true
}

// Update adds a value to the cache.  Returns true if an eviction occurred.
func (c *TimedCache) Update(key, value interface{}) bool {
	c.Lock()
	defer c.Unlock()

	c.purgeExpired()

	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		ent.Value.(*entry).tm = time.Now().Unix()
		return true
	}
	return false
}

// Get looks up a key's value from the cache.
func (c *TimedCache) Get(key interface{}) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()

	c.purgeExpired()

	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		if ent.Value.(*entry) == nil {
			return nil, false
		}
		ent.Value.(*entry).tm = time.Now().Unix()
		return ent.Value.(*entry).value, true
	}
	return nil, false
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *TimedCache) Contains(key interface{}) bool {
	c.Lock()
	defer c.Unlock()

	c.purgeExpired()

	_, ok := c.items[key]
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *TimedCache) Peek(key interface{}) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()

	c.purgeExpired()

	if ent, ok := c.items[key]; ok {
		return ent.Value.(*entry).value, true
	}
	return nil, false
}

// Remove removes the provided key from the cache
func (c *TimedCache) Remove(key interface{}) bool {
	c.Lock()
	defer c.Unlock()

	c.purgeExpired()

	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *TimedCache) Keys() []interface{} {
	c.Lock()
	defer c.Unlock()

	c.purgeExpired()

	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *TimedCache) Len() int {
	c.Lock()
	defer c.Unlock()

	c.purgeExpired()
	return c.evictList.Len()
}

// removeElement is used to remove a given timed-cache element from the cache
func (c *TimedCache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
	if c.onEvictCall != nil {
		c.onEvictCall(kv.key, kv.value)
	}
}

func (c *TimedCache) PurgeExpired() {
	c.Lock()
	defer c.Unlock()
	c.purgeExpired()
}

func (c *TimedCache) purgeExpired() {
	now := time.Now().Unix()

	for e := c.evictList.Back(); e != nil; e = e.Prev() {
		ent := e.Value.(*entry)
		if now-ent.tm > c.duration {
			c.removeElement(e)
		} else {
			// previous list elements are only newer
			// nothing to check more
			return
		}
	}
}

func (c *TimedCache) Print() {
	for e := c.evictList.Front(); e != nil; e = e.Next() {
		ent := e.Value.(*entry)
		fmt.Printf("tm: %v, key: %v, value: %v\n", time.Unix(ent.tm, 0), ent.key, ent.value)
	}
}
