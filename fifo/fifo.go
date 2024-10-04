package fifo

import (
	cache "cache_handler"
	"container/list"
)

// fifo 是一個 FIFO cache，不是平行處理安全的
type fifo struct {
	// 快取最大的容量，單位位元組
	// groupcache 使用的是最大儲存 entry 個數
	maxBytes int
	// 當一個 entry 從快取中移除時呼叫該回呼函數，預設為 nil
	// groupcache 中的 key 是任意的可比較類型： value 是 interface{}
	onEvicted func(key string, value interface{})

	// 已使用的位元組數，只包含值，不包含 key
	usedBytes int

	ll    *list.List
	cache map[string]*list.Element
}

type entry struct {
	key   string
	value interface{}
}

func (e *entry) Len() int {
	return cache.CalcLen(e.value)
}

func New(maxBytes int, onEvicted func(key string, value interface{})) cache.Cache {
	return &fifo{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

func (f *fifo) Set(key string, value interface{}) {
	if e, ok := f.cache[key]; ok {
		f.ll.MoveToBack(e)
		en := e.Value.(*entry)
		f.usedBytes = f.usedBytes - cache.CalcLen(en.value) + cache.CalcLen(value)
		en.value = value
		return
	}

	en := &entry{key, value}
	e := f.ll.PushBack(en)
	f.cache[key] = e

	f.usedBytes += en.Len()
	if f.maxBytes > 0 && f.usedBytes > f.maxBytes {
		f.DelOldest()
	}
}

func (f *fifo) Get(key string) interface{} {
	if e, ok := f.cache[key]; ok {
		return e.Value.(*entry).value
	}

	return nil
}

// 刪除指定的紀錄
func (f *fifo) Del(key string) {
	if e, ok := f.cache[key]; ok {
		f.removeElement(e)
	}
}

// 刪除最舊的紀錄
func (f *fifo) DelOldest() {
	f.removeElement(f.ll.Front())
}

func (f *fifo) removeElement(e *list.Element) {
	if e == nil {
		return
	}

	f.ll.Remove(e)
	en := e.Value.(*entry)
	f.usedBytes -= en.Len()
	delete(f.cache, en.key)

	if f.onEvicted != nil {
		f.onEvicted(en.key, en.value)
	}
}

// 取得快取紀錄數
func (f *fifo) Len() int {
	return f.ll.Len()
}
