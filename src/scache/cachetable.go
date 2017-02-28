package scache

import (
	"log"
	"sync"
	"time"
)

type CacheTable struct {
	sync.RWMutex
	Name            string
	Items           map[interface{}]*CacheItem
	CleanupTimer    *time.Timer
	CleanupInterval time.Duration
	Logger          *log.Logger
}

func (table *CacheTable) expirationCheck() {
	table.Lock()
	if table.CleanupTimer != nil {
		table.CleanupTimer.Stop()
	}

	if table.CleanupInterval > 0 {
		table.log("cleanupInterval : ", table.CleanupInterval)
	}

	items := table.Items
	table.Unlock()

	now := time.Now()
	smallestDuration := 0 * time.Second
	for k, v := range items {

		table.RLock()
		lifeSpan := v.LifeSpan
		createdAt := v.CreatedAt
		table.RUnlock()

		if lifeSpan == 0 {
			continue
		}

		if now.Sub(createdAt) >= lifeSpan {
			err := table.Delete(k)
			if err != nil {
				table.log("delete item error : ", v.Key, err)
			}
		}
		if smallestDuration == 0 || (lifeSpan-now.Sub(createdAt) < smallestDuration) {
			smallestDuration = lifeSpan - now.Sub(createdAt)
		}
	}

	table.Lock()
	table.CleanupInterval = smallestDuration
	if smallestDuration != 0 {
		table.CleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
	table.Unlock()

}

func (table *CacheTable) addInterval(item *CacheItem) {
	table.log("add item key: ", item.Key, " lifeSpan : ", item.LifeSpan, " to table : ", table.Name)
	table.Items[item.Key] = item

	cleanupInterval := table.CleanupInterval
	table.Unlock()

	if item.LifeSpan > 0 && (cleanupInterval == 0 || item.LifeSpan < cleanupInterval) {
		table.expirationCheck()
	}
}

func (table *CacheTable) Add(key interface{}, lifeSpan time.Duration, value interface{}) (*CacheItem, error) {

	table.RLock()
	if _, ok := table.Items[key]; ok {
		table.log(key, " has exist, add error")
		table.RUnlock()
		return nil, ErrKeyIsExist
	}
	table.RUnlock()

	item := NewItem(key, lifeSpan, value)

	table.Lock()
	table.addInterval(item)

	return item, nil
}

func (table *CacheTable) Delete(key interface{}) error {
	table.log("delete item : ", key)
	table.Lock()
	defer table.Unlock()
	_, ok := table.Items[key]
	if !ok {
		return ErrKeyNotFound
	}
	delete(table.Items, key)
	return nil
}

func (table *CacheTable) Value(key interface{}) (*CacheItem, error) {
	table.RLock()
	item, ok := table.Items[key]
	table.RUnlock()

	if !ok {
		return nil, ErrKeyNotFound
	}

	item.KeepAlive()
	return item, nil
}

func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.Items[key]
	return ok
}

func (table *CacheTable) Flush() {
	table.Lock()
	defer table.Unlock()

	table.Items = map[interface{}]*CacheItem{}
	table.CleanupInterval = 0
	if table.CleanupTimer != nil {
		table.CleanupTimer.Stop()
	}
}

func (table *CacheTable) Foreach(trans func(key interface{}, itme *CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	for k, v := range table.Items {
		trans(k, v)
	}
}

func (table *CacheTable) SetLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()
	table.Logger = logger
}

func (table *CacheTable) log(v ...interface{}) {
	if table.Logger == nil {
		return
	}
	table.Logger.Println(v)
}
