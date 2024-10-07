package lru

import (
	"container/list"

	cache "cache_handler"
)

//核心概念與 FIFO 相同，差別在每次存取時會將該 value 移動到最後面避免被刪除

// lru 是一個 LRU cache，不是平行處理安全的
type lru struct {
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
	return &lru{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

func (l *lru) Set(key string, value interface{}) {
	if e, ok := l.cache[key]; ok {
		l.ll.MoveToBack(e) // 這裡!!!差別在這裡!!!
		en := e.Value.(*entry)
		l.usedBytes = l.usedBytes - cache.CalcLen(en.value) + cache.CalcLen(value)
		en.value = value
		return
	}

	en := &entry{key, value}
	e := l.ll.PushBack(en)
	l.cache[key] = e

	l.usedBytes += en.Len()
	if l.maxBytes > 0 && l.usedBytes > l.maxBytes {
		l.DelOldest()
	}
}

func (l *lru) Get(key string) interface{} {
	if e, ok := l.cache[key]; ok {
		l.ll.MoveToBack(e)
		return e.Value.(*entry).value
	}

	return nil
}

// 刪除指定的紀錄
func (l *lru) Del(key string) {
	if e, ok := l.cache[key]; ok {
		l.removeElement(e)
	}
}

// 刪除最舊的紀錄
func (l *lru) DelOldest() {
	l.removeElement(l.ll.Front())
}

func (l *lru) Len() int {
	return l.ll.Len()
}

func (l *lru) removeElement(e *list.Element) {
	if e == nil {
		return
	}

	l.ll.Remove(e)
	en := e.Value.(*entry)
	l.usedBytes -= en.Len()
	delete(l.cache, en.key)

	if l.onEvicted != nil {
		l.onEvicted(en.key, en.value)
	}
}
