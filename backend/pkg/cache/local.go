// Package cache 提供 Redis 缓存封装，内置缓存三大问题完整防护。
//
// 本文件实现 L1 进程内内存缓存（LocalCache），作为多级缓存的第一层。
package cache

import (
	"sync"
	"time"
)

// localItem 本地缓存条目。
//
// 支持两种模式：
//  1. 普通 TTL 模式：expireAt 为绝对过期时间，过期后删除
//  2. 逻辑过期模式：logicalExp=true，logicalTime 为逻辑过期时间，
//     过期后不删除，返回旧数据并触发异步重建（热点数据防击穿）
type localItem struct {
	data        []byte    // 缓存数据（JSON 序列化后的字节）
	expireAt    time.Time // 普通模式：绝对过期时间
	isNull      bool      // 是否是空值标记（防穿透的 __NULL__）
	logicalExp  bool      // 是否为逻辑过期条目
	logicalTime time.Time // 逻辑过期时间
}

// LocalCache L1 进程内内存缓存，作为多级缓存的第一层（最快路径）。
//
// 设计目的：
//   - 微秒级访问（内存 map + RWMutex），比 Redis（毫秒级网络往返）快 1000 倍
//   - 抗 Redis 抖动：Redis 短暂不可用时 L1 继续提供服务
//   - TTL 极短（默认 30s~2min），保证数据一致性
//   - 后台 gc goroutine 定期清理过期条目，避免内存泄漏
//
// 注意：本地缓存是进程级的，多实例部署时各实例独立，
// 不保证一致性（短 TTL 可接受最终一致性）。
type LocalCache struct {
	mu     sync.RWMutex
	items  map[string]*localItem
	ttl    time.Duration // 默认 TTL
	stopCh chan struct{} // gc 停止信号
	once   sync.Once     // 确保 Stop 只执行一次
}

// NewLocalCache 创建本地缓存实例并启动后台 GC。
//
// 参数 defaultTTL: 默认过期时间，Set 时未指定 TTL 则使用此值。
func NewLocalCache(defaultTTL time.Duration) *LocalCache {
	if defaultTTL == 0 {
		defaultTTL = 1 * time.Minute
	}
	lc := &LocalCache{
		items:  make(map[string]*localItem),
		ttl:    defaultTTL,
		stopCh: make(chan struct{}),
	}
	go lc.gc()
	return lc
}

// Stop 停止后台 GC goroutine，防止 goroutine 泄漏。
func (lc *LocalCache) Stop() {
	lc.once.Do(func() {
		close(lc.stopCh)
	})
}

// gc 后台定期清理过期条目（仅清理普通 TTL 条目，逻辑过期条目不清理）。
//
// 清理频率：每分钟一次。逻辑过期条目永不过期（由异步重建更新），
// 直到被 Delete/DeleteByPattern 显式删除或被新值覆盖。
func (lc *LocalCache) gc() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lc.mu.Lock()
			now := time.Now()
			for k, v := range lc.items {
				if !v.logicalExp && now.After(v.expireAt) {
					delete(lc.items, k)
				}
			}
			lc.mu.Unlock()
		case <-lc.stopCh:
			return
		}
	}
}

// Get 从本地缓存读取数据。
//
// 返回值：
//   - data: 数据字节
//   - found: 是否命中
//   - isNull: 是否是空值标记（__NULL__）
//   - logicalExpired: 逻辑过期条目是否已过期（仅 logicalExp=true 时有意义）
//
// 注意：普通 TTL 条目过期时会在此处立即删除（惰性删除）；
// 逻辑过期条目过期时不删除，返回 (data, true, false, true)，由调用方触发异步重建。
func (lc *LocalCache) Get(key string) (data []byte, found bool, isNull bool, logicalExpired bool) {
	lc.mu.RLock()
	item, ok := lc.items[key]
	lc.mu.RUnlock()
	if !ok {
		return nil, false, false, false
	}
	if item.logicalExp {
		expired := time.Now().After(item.logicalTime)
		return item.data, true, item.isNull, expired
	}
	if time.Now().After(item.expireAt) {
		lc.mu.Lock()
		delete(lc.items, key)
		lc.mu.Unlock()
		return nil, false, false, false
	}
	return item.data, true, item.isNull, false
}

// Set 写入普通 TTL 缓存条目。
func (lc *LocalCache) Set(key string, data []byte, ttl time.Duration, isNull bool) {
	if ttl <= 0 {
		ttl = lc.ttl
	}
	lc.mu.Lock()
	lc.items[key] = &localItem{
		data:     data,
		expireAt: time.Now().Add(ttl),
		isNull:   isNull,
	}
	lc.mu.Unlock()
}

// SetLogical 写入逻辑过期缓存条目（热点数据）。
//
// 逻辑过期条目不会被 GC 清理，也不会在 Get 时删除；
// 过期后返回旧数据并由上层触发异步重建。
func (lc *LocalCache) SetLogical(key string, data []byte, logicalExpireAt time.Time) {
	lc.mu.Lock()
	lc.items[key] = &localItem{
		data:        data,
		logicalExp:  true,
		logicalTime: logicalExpireAt,
	}
	lc.mu.Unlock()
}

// Delete 删除单个缓存 Key。
func (lc *LocalCache) Delete(key string) {
	lc.mu.Lock()
	delete(lc.items, key)
	lc.mu.Unlock()
}

// DeleteByPattern 按前缀批量删除缓存 Key。
//
// 注意：当前实现为前缀匹配（prefix match），与 Redis SCAN 模式保持一致的行为。
func (lc *LocalCache) DeleteByPattern(prefix string) {
	lc.mu.Lock()
	for k := range lc.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(lc.items, k)
		}
	}
	lc.mu.Unlock()
}

// Clear 清空所有本地缓存（用于测试或全量失效场景）。
func (lc *LocalCache) Clear() {
	lc.mu.Lock()
	lc.items = make(map[string]*localItem)
	lc.mu.Unlock()
}
