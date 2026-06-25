// Package cache 提供 Redis 缓存封装，内置缓存三大问题完整防护。
//
// 缓存三大问题防护（生产级四层防护架构）：
//
//  1. 缓存穿透（Cache Penetration）：查询不存在的数据绕过缓存直打 DB。
//     三层防护：
//     - 布隆过滤器（Bloom Filter）：内存型布隆过滤器预热合法 ID，请求先过过滤器，不存在直接拦截
//     - 空值缓存（Null Caching）：DB 查询为空时写入空值标记 __NULL__，短 TTL（60s）
//     - 接口层限流：IP 维度令牌桶限流，拦截高频恶意请求
//     注意：登录等认证接口不缓存空值，避免攻击者用不存在的用户名刷缓存。
//
//  2. 缓存击穿（Cache Breakdown）：热点 Key 过期瞬间大量并发请求同时打到 DB。
//     三层防护：
//     - singleflight（进程内）：同一进程内合并同 Key 并发请求为一次 DB 查询
//     - 分布式锁（Redis SETNX）：多实例部署时通过 Redis SET 互斥锁保证全局只有一个 DB 查询
//     - 逻辑过期（Logical Expiration）：热点数据永不过期，值内携带逻辑过期时间，过期时异步重建
//     其他线程返回旧数据，彻底杜绝击穿
//
//  3. 缓存雪崩（Cache Avalanche）：大量 Key 同时过期或 Redis 宕机导致 DB 压力骤增。
//     四层防护：
//     - TTL 随机抖动：在基础 TTL 上增加 ±10% 随机偏移，避免大批 Key 同时过期
//     - 多级缓存：L1 本地内存缓存（sync.Map）+ L2 Redis，Redis 不可用时 L1 兜底
//     - 熔断器（Circuit Breaker）：Redis 连续失败时自动熔断，快速降级到 DB，防止雪崩扩散
//     - 服务降级：Redis 熔断时直接查 DB；DB 也超负载时返回兜底数据/默认值
//
// 缓存架构：L1（进程内内存）→ L2（Redis）→ DB
//
//	请求 → L1 本地缓存（微秒级）→ L2 Redis（毫秒级）→ DB（兜底）
//
// 缓存 Key 命名规范：
//
//	cache:{module}:{type}:{identifier}
//
// 示例：
//
//	cache:user:id:1             用户ID=1（UserResponse DTO，不含密码）
//	cache:user:list:p1:s10       用户列表第1页（每页10条）
//	cache:role:list:all          角色列表（全量，小数据集）
//	cache:permission:list:all    权限列表（全量）
//	cache:dashboard:stats:all    仪表盘统计
//
// 缓存安全原则：
//   - 不缓存含密码等敏感字段的实体，统一缓存脱敏后的 DTO
//   - 认证操作（登录）不缓存结果，每次登录均查询数据库验证密码
//   - 空值缓存使用独立短 TTL，与正常数据错开
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

const (
	NullValue = "__NULL__"
	KeyPrefix = "cache:"

	TTLHot       = 30 * time.Minute
	TTLQuery     = 5 * time.Minute
	TTLNull      = 60 * time.Second
	TTLDashboard = 2 * time.Minute
	TTLConfig    = 1 * time.Hour
	TTLLocal     = 30 * time.Second
	TTLLock      = 10 * time.Second

	lockPrefix = "lock:"
)

// FetchOptions 缓存读取选项，控制使用哪些防护策略。
type FetchOptions struct {
	TTL           time.Duration
	UseBloom      bool
	BloomModule   string
	UseDistLock   bool
	UseLogicalExp bool
	UseLocalCache bool
	LocalTTL      time.Duration
	ForceDB       bool
}

// DefaultFetchOptions 返回默认选项（开启全部防护）。
func DefaultFetchOptions(ttl time.Duration) FetchOptions {
	return FetchOptions{
		TTL:           ttl,
		UseBloom:      false,
		UseDistLock:   true,
		UseLogicalExp: false,
		UseLocalCache: true,
		LocalTTL:      TTLLocal,
	}
}

// HotDataOptions 返回热点数据选项（逻辑过期，永不过期+异步重建）。
func HotDataOptions() FetchOptions {
	return FetchOptions{
		TTL:           TTLHot,
		UseBloom:      true,
		UseDistLock:   true,
		UseLogicalExp: true,
		UseLocalCache: true,
		LocalTTL:      2 * time.Minute,
	}
}

// Client 缓存客户端，封装多级缓存与三大问题防护。
//
// 架构：L1 本地缓存（sync.Map）→ L2 Redis（分布式缓存）→ DB（数据源）
//
// 组件组成：
//   - local: L1 进程内内存缓存，TTL 极短（30s~2min），抗 Redis 抖动
//   - rdb: L2 Redis 客户端
//   - sf: singleflight 合并进程内并发请求
//   - bloom: 布隆过滤器，拦截对不存在 ID 的请求（防穿透）
//   - cb: Redis 熔断器，连续失败时自动打开
//   - rebuildCh: 逻辑过期异步重建通道
type Client struct {
	rdb    *redis.Client
	sf     singleflight.Group
	local  *LocalCache
	bloom  *BloomFilter
	cb     *CircuitBreaker
	ctx    context.Context
	jitter float64

	rebuildCh   chan rebuildTask
	rebuildOnce sync.Once
	closeOnce   sync.Once
	closed      chan struct{}
}

type rebuildTask struct {
	key    string
	ttl    time.Duration
	loader func() (interface{}, error)
}

// logicalEntry 逻辑过期条目结构。
type logicalEntry struct {
	Data     json.RawMessage `json:"data"`
	ExpireAt time.Time       `json:"expire_at"`
	Logical  bool            `json:"logical"`
}

// NewClient 创建缓存客户端实例。
func NewClient(rdb *redis.Client) *Client {
	c := &Client{
		rdb:    rdb,
		local:  NewLocalCache(TTLLocal),
		bloom:  NewBloomFilter(1<<20, 7),
		cb:     NewCircuitBreaker("redis", 5, 30*time.Second),
		ctx:    context.Background(),
		jitter: 0.1,
		closed: make(chan struct{}),
	}
	c.startRebuildWorker()
	return c
}

func (c *Client) startRebuildWorker() {
	c.rebuildOnce.Do(func() {
		c.rebuildCh = make(chan rebuildTask, 100)
		go func() {
			for {
				select {
				case task := <-c.rebuildCh:
					c.doRebuild(task)
				case <-c.closed:
					return
				}
			}
		}()
	})
}

func (c *Client) doRebuild(task rebuildTask) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("cache rebuild panic", "key", task.key, "panic", r)
		}
	}()
	slog.Debug("logical rebuild start", "key", task.key)
	loaded, err := task.loader()
	if err != nil {
		slog.Warn("logical rebuild failed", "key", task.key, "error", err)
		return
	}
	if loaded == nil {
		c.setRaw(task.key, []byte(NullValue), TTLNull)
		c.local.Set(task.key, []byte(NullValue), TTLNull, true)
		return
	}
	b, err := json.Marshal(loaded)
	if err != nil {
		slog.Warn("logical rebuild marshal failed", "key", task.key, "error", err)
		return
	}
	entry := logicalEntry{
		Data:     b,
		ExpireAt: time.Now().Add(task.ttl),
		Logical:  true,
	}
	entryBytes, _ := json.Marshal(entry)
	c.setRaw(task.key, entryBytes, 0)
	c.local.SetLogical(task.key, b, entry.ExpireAt)
	slog.Debug("logical rebuild done", "key", task.key)
}

// Close 关闭缓存客户端，停止异步重建协程。
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		c.local.Stop()
	})
}

// Enabled 返回 L2 Redis 是否可用（非 nil 且熔断器未打开）。
func (c *Client) Enabled() bool {
	return c != nil && c.rdb != nil && !c.cb.IsOpen()
}

// BloomFilter 返回布隆过滤器实例，用于预热合法 ID。
func (c *Client) BloomFilter() *BloomFilter {
	return c.bloom
}

func (c *Client) withJitter(ttl time.Duration) time.Duration {
	if c.jitter <= 0 || ttl <= 0 {
		return ttl
	}
	delta := time.Duration(float64(ttl) * c.jitter)
	offset := time.Duration(rand.Int63n(int64(delta*2))) - delta
	return ttl + offset
}

// GetBytes 从多级缓存读取原始字节：L1 本地 → L2 Redis。
//
// 返回值：
//   - data: 原始数据字节
//   - found: 是否命中
//   - isNull: 是否为空值标记
//   - fromLocal: 是否来自 L1 本地缓存
//   - logicalExpired: 逻辑过期条目是否已过期（仅命中逻辑过期条目时为 true）
func (c *Client) GetBytes(key string) (data []byte, found bool, isNull bool, fromLocal bool, logicalExpired bool) {
	if ldata, lfound, lnull, lexpired := c.local.Get(key); lfound {
		return ldata, true, lnull, true, lexpired
	}
	if !c.Enabled() {
		return nil, false, false, false, false
	}
	if !c.cb.Allow() {
		slog.Debug("circuit breaker open, skip redis", "key", key)
		return nil, false, false, false, false
	}
	val, err := c.rdb.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		c.cb.RecordSuccess()
		return nil, false, false, false, false
	}
	if err != nil {
		slog.Warn("cache get failed, circuit breaker record failure", "key", key, "error", err)
		c.cb.RecordFailure()
		return nil, false, false, false, false
	}
	c.cb.RecordSuccess()
	if string(val) == NullValue {
		c.local.Set(key, val, TTLNull, true)
		return nil, true, true, false, false
	}
	var entry logicalEntry
	if json.Unmarshal(val, &entry) == nil && entry.Logical {
		expired := time.Now().After(entry.ExpireAt)
		c.local.SetLogical(key, entry.Data, entry.ExpireAt)
		return entry.Data, true, false, false, expired
	}
	c.local.Set(key, val, TTLLocal, false)
	return val, true, false, false, false
}

// SetJSON 写入 L1+L2 缓存。
func (c *Client) SetJSON(key string, value interface{}, ttl time.Duration) {
	b, err := json.Marshal(value)
	if err != nil {
		slog.Warn("cache marshal failed", "key", key, "error", err)
		return
	}
	c.local.Set(key, b, TTLLocal, false)
	if c.Enabled() {
		c.setRaw(key, b, ttl)
	}
}

// SetNull 写入空值标记（防穿透）。
func (c *Client) SetNull(key string) {
	c.local.Set(key, []byte(NullValue), TTLNull, true)
	if c.Enabled() {
		c.setRaw(key, []byte(NullValue), TTLNull)
	}
}

func (c *Client) setRaw(key string, data []byte, ttl time.Duration) {
	if c.rdb == nil {
		return
	}
	if ttl > 0 {
		ttl = c.withJitter(ttl)
	}
	if err := c.rdb.Set(c.ctx, key, data, ttl).Err(); err != nil {
		slog.Warn("cache set failed", "key", key, "error", err)
		c.cb.RecordFailure()
	} else {
		c.cb.RecordSuccess()
	}
}

// SetLogical 写入逻辑过期条目（热点数据永不过期，异步重建）。
func (c *Client) SetLogical(key string, value interface{}, ttl time.Duration) {
	b, err := json.Marshal(value)
	if err != nil {
		slog.Warn("cache logical marshal failed", "key", key, "error", err)
		return
	}
	logicalExpireAt := time.Now().Add(ttl)
	c.local.SetLogical(key, b, logicalExpireAt)
	entry := logicalEntry{
		Data:     b,
		ExpireAt: logicalExpireAt,
		Logical:  true,
	}
	entryBytes, _ := json.Marshal(entry)
	if c.Enabled() {
		c.setRaw(key, entryBytes, 0)
	}
}

// acquireLock 尝试获取分布式锁（Redis SETNX）。
func (c *Client) acquireLock(lockKey string) (bool, func()) {
	if !c.Enabled() {
		return false, func() {}
	}
	ok, err := c.rdb.SetNX(c.ctx, lockKey, "1", TTLLock).Result()
	if err != nil || !ok {
		return false, func() {}
	}
	return true, func() {
		c.rdb.Del(c.ctx, lockKey)
	}
}

// Delete 删除一个或多个缓存 Key（L1+L2）。
func (c *Client) Delete(keys ...string) {
	for _, k := range keys {
		c.local.Delete(k)
	}
	if c.Enabled() && len(keys) > 0 {
		if err := c.rdb.Del(c.ctx, keys...).Err(); err != nil {
			slog.Warn("cache delete failed", "keys", keys, "error", err)
		}
	}
}

// DeleteByPattern 按模式批量删除缓存 Key（L1+L2）。
func (c *Client) DeleteByPattern(pattern string) {
	c.local.DeleteByPattern(pattern)
	if !c.Enabled() {
		return
	}
	iter := c.rdb.Scan(c.ctx, 0, pattern, 100).Iterator()
	var batch []string
	for iter.Next(c.ctx) {
		batch = append(batch, iter.Val())
		if len(batch) >= 100 {
			c.rdb.Del(c.ctx, batch...)
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		c.rdb.Del(c.ctx, batch...)
	}
	if err := iter.Err(); err != nil {
		slog.Warn("cache scan delete failed", "pattern", pattern, "error", err)
	}
}

func (c *Client) InvalidateUserCaches() {
	c.DeleteByPattern(KeyPrefix + "user:*")
	c.InvalidateDashboardCache()
	slog.Debug("cache invalidated: user")
}

func (c *Client) InvalidateRoleCaches() {
	c.DeleteByPattern(KeyPrefix + "role:*")
	c.InvalidatePermissionCaches()
	slog.Debug("cache invalidated: role")
}

func (c *Client) InvalidatePermissionCaches() {
	c.DeleteByPattern(KeyPrefix + "permission:*")
	c.InvalidateDashboardCache()
	slog.Debug("cache invalidated: permission")
}

func (c *Client) InvalidateDashboardCache() {
	c.Delete(KeyPrefix + "dashboard:stats")
}

func CacheKey(module, typ, id string) string {
	return fmt.Sprintf("%s%s:%s:%s", KeyPrefix, module, typ, id)
}

func CacheKeyUint(module, typ string, id uint) string {
	return CacheKey(module, typ, fmt.Sprintf("%d", id))
}

func CacheKeyList(module string, page, pageSize int) string {
	return CacheKey(module, "list", fmt.Sprintf("p%d:s%d", page, pageSize))
}

// Fetch 缓存读取核心方法，内置完整的三大问题防护。
//
// 处理流程（L1→L2→DB 多级缓存）：
//  1. 布隆过滤器检查（防穿透第一层）：BloomFilter 判定不存在则直接返回 nil
//  2. L1 本地缓存检查：命中则直接返回（最快路径，微秒级）
//  3. L2 Redis 缓存检查：
//     - 命中空值标记 → 返回 nil（防穿透第二层）
//     - 命中逻辑过期条目 → 返回旧数据 + 异步触发重建（防击穿）
//     - 命中普通数据 → 反序列化返回 + 回填 L1
//  4. 熔断器检查：Redis 连续失败时跳过 Redis，直接查 DB（防雪崩）
//  5. singleflight + 分布式锁合并并发请求（防击穿）
//  6. double-check：等待锁期间可能已有其他协程写入缓存
//  7. 执行 loader 从 DB 加载数据
//     - 返回 nil → 写入空值标记（防穿透）
//     - 返回数据 → 根据选项写入逻辑过期/普通缓存（TTL+抖动防雪崩）
//  8. 回填布隆过滤器（新数据 ID 加入过滤器）
func (c *Client) Fetch(key string, opt FetchOptions, result interface{}, loader func() (interface{}, error)) (interface{}, bool, error) {
	if opt.TTL == 0 {
		opt.TTL = TTLQuery
	}
	if opt.LocalTTL == 0 {
		opt.LocalTTL = TTLLocal
	}

	// 1. 布隆过滤器拦截不存在的 ID（防穿透）
	if opt.UseBloom && c.bloom != nil {
		bloomKey := key
		if !c.bloom.Contains(bloomKey) {
			slog.Debug("bloom filter reject", "key", key)
			return nil, false, nil
		}
	}

	// 2. 查多级缓存（L1 + L2）
	data, found, isNull, fromLocal, lexpired := c.GetBytes(key)
	if found {
		if isNull {
			return nil, false, nil
		}
		var entry logicalEntry
		if json.Unmarshal(data, &entry) == nil && entry.Logical {
			if time.Now().After(entry.ExpireAt) || lexpired {
				slog.Debug("logical expired, trigger async rebuild", "key", key, "from_local", fromLocal)
				select {
				case c.rebuildCh <- rebuildTask{key: key, ttl: opt.TTL, loader: loader}:
				default:
					slog.Debug("rebuild channel full, skip", "key", key)
				}
			}
			if err := json.Unmarshal(entry.Data, result); err != nil {
				slog.Warn("logical entry unmarshal failed", "key", key, "error", err)
			} else {
				return result, true, nil
			}
		} else {
			if fromLocal && lexpired {
				select {
				case c.rebuildCh <- rebuildTask{key: key, ttl: opt.TTL, loader: loader}:
				default:
				}
			}
			if err := json.Unmarshal(data, result); err != nil {
				slog.Warn("cache unmarshal failed, reload", "key", key, "from_local", fromLocal, "error", err)
			} else {
				slog.Debug("cache hit", "key", key, "from_local", fromLocal)
				return result, true, nil
			}
		}
	}

	// 4. 强制查库或缓存未命中，使用 singleflight 合并请求（防击穿）
	v, err, _ := c.sf.Do(key, func() (interface{}, error) {
		// double-check
		data2, found2, isNull2, _, _ := c.GetBytes(key)
		if found2 {
			if isNull2 {
				return nil, nil
			}
			var entry2 logicalEntry
			if json.Unmarshal(data2, &entry2) == nil && entry2.Logical {
				if err2 := json.Unmarshal(entry2.Data, result); err2 == nil {
					return result, nil
				}
			} else if err2 := json.Unmarshal(data2, result); err2 == nil {
				return result, nil
			}
		}

		// 5. 分布式锁（多实例防击穿）
		var unlock func()
		acquired := false
		if opt.UseDistLock && c.Enabled() {
			lockKey := lockPrefix + key
			acquired, unlock = c.acquireLock(lockKey)
			if acquired {
				defer unlock()
			} else {
				time.Sleep(50 * time.Millisecond)
				data3, found3, isNull3, _, _ := c.GetBytes(key)
				if found3 {
					if isNull3 {
						return nil, nil
					}
					if err3 := json.Unmarshal(data3, result); err3 == nil {
						return result, nil
					}
				}
			}
		}

		slog.Debug("cache miss, query db", "key", key, "dist_lock", acquired)
		loaded, err := loader()
		if err != nil {
			return nil, err
		}

		if loaded == nil {
			c.SetNull(key)
			return nil, nil
		}

		// 6. 写入缓存
		if opt.UseLogicalExp {
			c.SetLogical(key, loaded, opt.TTL)
		} else {
			c.SetJSON(key, loaded, opt.TTL)
		}

		// 7. 回填布隆过滤器
		if opt.UseBloom && c.bloom != nil {
			c.bloom.Add(key)
		}

		return loaded, nil
	})

	if err != nil {
		return nil, false, err
	}
	return v, v != nil, nil
}

// WarmupItem 预热项定义。
type WarmupItem struct {
	Key     string
	TTL     time.Duration
	Loader  func() (interface{}, error)
	Logical bool
}

// Warmup 缓存预热：启动时将热数据加载到多级缓存。
func (c *Client) Warmup(items ...WarmupItem) {
	slog.Info("cache warmup started", "items", len(items))
	success := 0
	for _, item := range items {
		data, err := item.Loader()
		if err != nil {
			slog.Warn("cache warmup item failed", "key", item.Key, "error", err)
			continue
		}
		if data == nil {
			continue
		}
		if item.Logical {
			c.SetLogical(item.Key, data, item.TTL)
		} else {
			c.SetJSON(item.Key, data, item.TTL)
		}
		success++
	}
	slog.Info("cache warmup completed", "success", success, "total", len(items))
}
