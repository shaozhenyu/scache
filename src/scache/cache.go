package scache

import (
	"sync"
)

var (
	cache = make(map[string]*CacheTable)
	mutex sync.RWMutex
)

func Cache(tableName string) *CacheTable {
	mutex.Lock()
	table, ok := cache[tableName]
	mutex.Unlock()

	if !ok {
		table = &CacheTable{
			Name:  tableName,
			Items: map[interface{}]*CacheItem{},
		}

		mutex.Lock()
		cache[tableName] = table
		mutex.Unlock()
	}
	return table
}
