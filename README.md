# cache_handler
參考資料：《用 Go 語言完成 6 個大型專案》第五章節 + https://github.com/go-programming-tour-book/cache-example

### 常用的快取淘汰演算法
FIFO (First In First Out): 先進先出演算法
LFU (Least Frequently Used): 最少使用演算法
LRU (Least Recently Used): 最近最少使用演算法，最常被使用的演算法!

### 測試快取的平行處理效能
pkg: allegro/bigcache-bench
提升效能: 減少鎖競爭、避免GC
```
go get -u golang.org/x/perf/cmd/benchstat
```