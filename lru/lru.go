package lru

import (
	"container/list"
)

// 最近最少使用，相对于仅考虑时间因素的FIFO和仅考虑访问频率的LFU，LRU算法可以认为是相对平衡的一种淘汰算法。
// LRU: 如果数据最近被访问过，那么将来被访问的概率也会更高。
// 算法的实现
// 1.维护一个队列，如果某条记录被访问了，则移动到队尾，
// 2.队首则是最近最少访问的数据，淘汰该条记录即可。

// Cache 核心结构体
type Cache struct {
	maxBytes int                      // 最大内存
	nBytes   int                      // 目前内存
	ll       *list.List               // 维护的队列
	cache    map[string]*list.Element // 缓存内的数据

	OnEvicted func(key string, value string)
}

// entry 双向链表结点
// 正常情况下在key里面存数据就可以了
// 淘汰队首节点时候,需要用key从字典中删除对应的映射
type entry struct {
	key   string
	value string
}

// NewCache 实例化Cache
func NewCache(maxBytes int, OnEvicted func(string, string)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

// Get 查询key所对应的value
// 1.找到对应双向链表的节点
// 2.将该节点移动到队尾,并且返回查找到的值
func (c *Cache) Get(key string) (value string, ok bool) {
	if data, ok := c.cache[key]; ok {
		// 将该节点移动到队尾
		c.ll.MoveToFront(data)
		// 返回k,value
		kv := data.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 缓存淘汰
func (c *Cache) RemoveOldest() {
	// 取到队首节点
	ele := c.ll.Back()
	if ele != nil {
		// 删除队首节点
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 删除cache中该节点的映射关系
		delete(c.cache, kv.key)
		// 更新缓存当前的字节数
		c.nBytes -= len(kv.key) + len(kv.value)
		// 若回调函数不为nil,缓存淘汰的时候调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value string) {
	// 如果key存在
	if ele, ok := c.cache[key]; ok {
		// 将该节点移动到队尾
		c.ll.MoveToFront(ele)
		// 获取entry节点
		kv := ele.Value.(*entry)
		c.nBytes += len(value) - len(kv.value)
		// 更新对应节点的值
		kv.value = value
	} else { // key不存在
		// 向队尾添加节点
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += len(key) + len(value)
	}
	// 判断是否超出内存
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
