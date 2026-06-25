// Package cache 提供 Redis 缓存封装，内置缓存三大问题完整防护。
//
// 本文件实现布隆过滤器（BloomFilter），用于缓存穿透防护。
package cache

import (
	"hash/fnv"
	"sync"
)

// BloomFilter 布隆过滤器，用于缓存穿透防护。
//
// 原理：使用 k 个哈希函数将元素映射到位数组的 k 个位置，查询时
// 如果任一位置为 0 则元素一定不存在；如果全部为 1 则元素可能存在（有小概率误判）。
//
// 适用场景：
//   - 预热所有合法 ID（如用户ID、角色ID）
//   - 请求先过布隆过滤器，不存在直接拦截，避免查 Redis 和 DB
//   - 误判率可控（m=1<<20, k=7，约 100 万元素时误判率 < 1%）
//
// 特性：
//   - 支持并发安全（内部使用 RWMutex）
//   - 内存占用固定（bitset 大小在创建时确定，不会动态增长）
//   - 不支持删除（如需删除可用 Counting Bloom Filter，本场景不需要）
type BloomFilter struct {
	m      uint         // 位数组大小（bit 数）
	k      uint         // 哈希函数个数
	bitset []bool       // 位数组（使用 bool 切片，简单直观；追求极致性能可改用 bit packing）
	mu     sync.RWMutex // 读写锁，支持并发读
	count  uint         // 已添加元素计数（近似值，用于估算误判率）
}

// NewBloomFilter 创建布隆过滤器。
//
// 参数：
//   - m: 位数组大小（推荐 1<<20 = 1,048,576 bits ≈ 128KB，支持约百万级元素）
//   - k: 哈希函数个数（推荐 7，对于 m/n=10 的场景最优）
//
// 误判率参考公式：p ≈ (1 - e^(-kn/m))^k
// 当 n=100000, m=1<<20, k=7 时，p ≈ 0.8%
func NewBloomFilter(m uint, k uint) *BloomFilter {
	if m == 0 {
		m = 1 << 20
	}
	if k == 0 {
		k = 7
	}
	return &BloomFilter{
		m:      m,
		k:      k,
		bitset: make([]bool, m),
	}
}

// locations 计算一个 key 的 k 个哈希位置。
//
// 使用双重哈希（Double Hashing）技术：通过两个独立哈希值线性组合生成 k 个位置，
// 比使用 k 个不同哈希函数更高效，且分布性良好。
//
// 哈希算法选择：FNV-1a（非加密哈希，速度快，分布均匀）。
func (bf *BloomFilter) locations(key string) []uint {
	h := fnv.New64a()
	h.Write([]byte(key))
	hash1 := h.Sum64()
	h.Reset()
	h.Write([]byte(key + "|salt"))
	hash2 := h.Sum64()
	locs := make([]uint, bf.k)
	for i := uint(0); i < bf.k; i++ {
		locs[i] = uint((hash1 + uint64(i)*hash2) % uint64(bf.m))
	}
	return locs
}

// Add 向布隆过滤器添加一个元素。
//
// 将 k 个哈希位置全部设为 true。
func (bf *BloomFilter) Add(key string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	for _, loc := range bf.locations(key) {
		bf.bitset[loc] = true
	}
	bf.count++
}

// Contains 判断元素是否可能存在。
//
// 返回值：
//   - true: 元素可能存在（存在小概率误判）
//   - false: 元素一定不存在（100% 准确，可安全拦截）
//
// 缓存穿透防护正是利用"false 一定不存在"这一特性来拦截非法请求。
func (bf *BloomFilter) Contains(key string) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	for _, loc := range bf.locations(key) {
		if !bf.bitset[loc] {
			return false
		}
	}
	return true
}

// Count 返回已添加元素的近似数量。
func (bf *BloomFilter) Count() uint {
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	return bf.count
}

// Reset 清空布隆过滤器（用于缓存全量失效后重建）。
func (bf *BloomFilter) Reset() {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	bf.bitset = make([]bool, bf.m)
	bf.count = 0
}
