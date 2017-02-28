package scache

import (
	"sync"
	"time"
)

type CacheItem struct {
	sync.RWMutex
	Key          interface{}
	Value        interface{}
	LifeSpan     time.Duration
	CreatedAt    time.Time
	AccessdAt    time.Time
	AccessdTimes int64
	//TODO callback when time expire
}

func NewItem(key interface{}, lifeSpan time.Duration, value interface{}) *CacheItem {
	now := time.Now()
	item := &CacheItem{
		Key:          key,
		Value:        value,
		LifeSpan:     lifeSpan,
		CreatedAt:    now,
		AccessdAt:    now,
		AccessdTimes: 0,
	}
	return item
}

func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.AccessdAt = time.Now()
	item.AccessdTimes += 1
}
