package scache

import (
	"bytes"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	k = "testkey"
	v = "testvalue"
)

func TestCache(t *testing.T) {
	table := Cache("testCache")
	_, err := table.Add(k+"_1", 0*time.Second, v)
	if err != nil {
		t.Error("Error add key", err)
	}
	_, err = table.Add(k+"_2", 1*time.Second, v)
	if err != nil {
		t.Error("Error add key", err)
	}

	p, err := table.Value(k + "_1")
	if err != nil || p == nil || p.Value.(string) != v {
		t.Error("Error retrieving non expiring data from cache", err)
	}
	p, err = table.Value(k + "_2")
	if err != nil || p == nil || p.Value.(string) != v {
		t.Error("Error retrieving data from cache", err)
	}

	if p.AccessdTimes != 1 {
		t.Error("Error getting correct access count")
	}
	if p.LifeSpan != 1*time.Second {
		t.Error("Error getting correct lifespan")
	}
	if p.CreatedAt.UnixNano() == 0 {
		t.Error("Error getting correct created")
	}
	if p.AccessdAt.UnixNano() == 0 {
		t.Error("Error getting correct accessat")
	}
}

func TestCacheExpire(t *testing.T) {
	table := Cache("testCache")

	_, err := table.Add(k+"_3", 100*time.Millisecond, v)
	if err != nil {
		t.Error("Error add key", err)
	}

	time.Sleep(75 * time.Millisecond)

	_, err = table.Value(k + "_3")
	if err != nil {
		t.Error("Error retrieving value from cache:", err)
	}

	time.Sleep(75 * time.Millisecond)

	_, err = table.Value(k + "_3")
	if err == nil {
		t.Error("Found key which should have been expired by now")
	}
}

func TestExists(t *testing.T) {
	table := Cache("testExists")
	table.Add(k, 0, v)
	if !table.Exists(k) {
		t.Error("Error verifying existing data in cache")
	}
}

func TestAddExists(t *testing.T) {
	table := Cache("testaddexits")

	var finish sync.WaitGroup
	var added int32
	var idle int32

	fn := func(id int) {
		for i := 0; i < 100; i++ {
			_, err := table.Add(i, 0, i)
			if err == nil {
				atomic.AddInt32(&added, 1)
			} else {
				atomic.AddInt32(&idle, 1)
			}
			time.Sleep(0)
		}
		finish.Done()
	}

	finish.Add(10)
	go fn(0x0000)
	go fn(0x1100)
	go fn(0x2200)
	go fn(0x3300)
	go fn(0x4400)
	go fn(0x5500)
	go fn(0x6600)
	go fn(0x7700)
	go fn(0x8800)
	go fn(0x9900)
	finish.Wait()

	t.Log(added, idle)
	table.Foreach(func(key interface{}, item *CacheItem) {
		t.Logf("%02x  %04x\n", key.(int), item.Value.(int))
	})
}

func TestDelete(t *testing.T) {
	table := Cache("testDelete")
	table.Add(k, 0, v)
	p, err := table.Value(k)
	if err != nil || p == nil || p.Value.(string) != v {
		t.Error("Error retrieving data from cache", err)
	}

	table.Delete(k)
	p, err = table.Value(k)
	if err == nil || p != nil {
		t.Error("Error deleting data")
	}

	err = table.Delete(k)
	if err == nil {
		t.Error("Expected error deleting item")
	}
}

func TestFlush(t *testing.T) {
	table := Cache("testFlush")
	table.Add(k, 10*time.Second, v)
	table.Flush()

	p, err := table.Value(k)
	if err == nil || p != nil {
		t.Error("Error flushing table")
	}

	if len(table.Items) != 0 {
		t.Error("Error verifying count of flushed table")
	}
}

func TestCount(t *testing.T) {
	table := Cache("testCount")
	count := 100000
	for i := 0; i < count; i++ {
		key := k + strconv.Itoa(i)
		table.Add(key, 10*time.Second, v)
	}

	for i := 0; i < count; i++ {
		key := k + strconv.Itoa(i)
		p, err := table.Value(key)
		if err != nil || p == nil || p.Value.(string) != v {
			t.Error("Error retrieving data")
		}
	}

	if len(table.Items) != count {
		t.Error("Data count mismatch")
	}
}

func TestAccessCount(t *testing.T) {
	count := 100
	table := Cache("testAccessCount")
	for i := 0; i < count; i++ {
		table.Add(i, 10*time.Second, v)
	}

	for i := 0; i < count; i++ {
		for j := 0; j < i; j++ {
			table.Value(i)
		}
	}

	for i := 0; i < count; i++ {
		if item, ok := table.Items[i]; ok {
			if item.AccessdTimes != int64(i) {
				t.Error("Error item accessedTimes")
			}
		} else {
			t.Error("Error get item")
		}
	}
}

func TestLogger(t *testing.T) {
	out := new(bytes.Buffer)
	l := log.New(out, "scache ", log.Ldate|log.Ltime)

	table := Cache("testLogger")
	table.SetLogger(l)
	table.Add(k, 1*time.Second, v)

	if out.Len() == 0 {
		t.Error("Logger is empty")
	}
}
