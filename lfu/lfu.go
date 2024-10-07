package lfu

import (
	cache "cache_handler"
	"container/heap"
)

// lfu 是一個 LFU cache，不是平行處理安全的
type lfu struct {
	// 快取最大的容量，單位位元組
	// groupcache 使用的是最大儲存 entry 個數
	maxBytes int
	// 當一個 entry 從快取中移除時呼叫該回呼函數，預設為 nil
	// groupcache 中的 key 是任意的可比較類型： value 是 interface{}
	onEvicted func(key string, value interface{})

	// 已使用的位元組數，只包含值，不包含 key
	usedBytes int

	queue *queue
	cache map[string]*entry
}

func New(maxBytes int, onEvicted func(key string, value interface{})) cache.Cache {
	q := make(queue, 0, 1024)
	return &lfu{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		queue:     &q,
		cache:     make(map[string]*entry),
	}
}

func (l *lfu) Set(key string, value interface{}) {
	if e, ok := l.cache[key]; ok {
		l.usedBytes = l.usedBytes - cache.CalcLen(e.value) + cache.CalcLen(value)
		l.queue.update(e, value, e.weight+1)
		return
	}

	en := &entry{key: key, value: value}
	heap.Push(l.queue, en)
	l.cache[key] = en

	l.usedBytes += en.Len()
	if l.maxBytes > 0 && l.usedBytes > l.maxBytes {
		l.removeElement(heap.Pop(l.queue))
	}
}

// Get 从 cache 中获取 key 对应的值，nil 表示 key 不存在
func (l *lfu) Get(key string) interface{} {
	if e, ok := l.cache[key]; ok {
		l.queue.update(e, e.value, e.weight+1)
		return e.value
	}

	return nil
}

// 刪除指定的紀錄
func (l *lfu) Del(key string) {
	if e, ok := l.cache[key]; ok {
		heap.Remove(l.queue, e.index)
		l.removeElement(e)
	}
}

// 刪除最舊的紀錄
func (l *lfu) DelOldest() {
	if l.queue.Len() == 0 {
		return
	}
	l.removeElement(heap.Pop(l.queue))
}

// Len 返回目前 cache 中的記錄數量
func (l *lfu) Len() int {
	return l.queue.Len()
}

func (l *lfu) removeElement(x interface{}) {
	if x == nil {
		return
	}

	en := x.(*entry)

	delete(l.cache, en.key)

	l.usedBytes -= en.Len()

	if l.onEvicted != nil {
		l.onEvicted(en.key, en.value)
	}
}
