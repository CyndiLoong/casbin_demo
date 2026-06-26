// Package middleware 提供 HTTP 中间件，包括 JWT 认证、Casbin 权限校验、限流、Panic 恢复等。
package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"casbin-demo/internal/repository"
)

// 限流相关 Redis key 前缀常量
const (
	// rateLimitIPPrefix IP 维度限流 key 前缀
	// key 格式: rate:ip:{ip}
	rateLimitIPPrefix = "rate:ip:"

	// rateLimitAPIPrefix 接口维度限流 key 前缀
	// key 格式: rate:api:{method}:{path}
	rateLimitAPIPrefix = "rate:api:"

	// rateLimitUserPrefix 用户维度限流 key 前缀（登录用户）
	// key 格式: rate:user:{user_id}
	rateLimitUserPrefix = "rate:user:"

	// rateLimitStatsPrefix 限流统计 key 前缀
	rateLimitStatsPrefix = "rate:stats:"
)

// 熔断器状态常量
const (
	circuitClosed   = iota // 熔断器关闭：正常使用 Redis
	circuitOpen            // 熔断器打开：降级为本地限流
	circuitHalfOpen        // 熔断器半开：尝试恢复 Redis
)

// circuitBreaker 熔断器，用于 Redis 故障时自动降级。
//
// 工作原理：
//  1. 连续失败达到阈值（failureThreshold）时，熔断器打开，直接使用本地限流
//  2. 打开状态持续 resetTimeout 后，进入半开状态，允许少量请求尝试 Redis
//  3. 半开状态下连续成功达到 successThreshold，熔断器关闭，恢复 Redis 限流
//  4. 半开状态下只要失败一次，立即重新打开熔断器
type circuitBreaker struct {
	failures        int64         // 当前连续失败次数
	successes       int64         // 半开状态下连续成功次数
	state           int32         // 熔断器状态 (atomic)
	lastFailureTime time.Time     // 上次失败时间
	failureThreshold int64        // 失败阈值（达到则打开熔断）
	successThreshold int64        // 半开时成功阈值（达到则关闭熔断）
	resetTimeout    time.Duration // 打开状态持续时间
}

// newCircuitBreaker 创建熔断器实例。
func newCircuitBreaker() *circuitBreaker {
	return &circuitBreaker{
		state:            circuitClosed,
		failureThreshold: 5,
		successThreshold: 3,
		resetTimeout:     30 * time.Second,
	}
}

// allowRedis 检查是否允许使用 Redis。
// 返回 true 表示可以尝试 Redis，false 表示应使用本地限流。
func (cb *circuitBreaker) allowRedis() bool {
	state := atomic.LoadInt32(&cb.state)
	switch state {
	case circuitClosed:
		return true
	case circuitOpen:
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			atomic.CompareAndSwapInt32(&cb.state, circuitOpen, circuitHalfOpen)
			atomic.StoreInt64(&cb.successes, 0)
			slog.Info("rate limiter circuit breaker: half-open, trying redis")
			return true
		}
		return false
	case circuitHalfOpen:
		return true
	default:
		return false
	}
}

// recordSuccess 记录 Redis 操作成功。
func (cb *circuitBreaker) recordSuccess() {
	state := atomic.LoadInt32(&cb.state)
	if state == circuitHalfOpen {
		s := atomic.AddInt64(&cb.successes, 1)
		if s >= cb.successThreshold {
			atomic.CompareAndSwapInt32(&cb.state, circuitHalfOpen, circuitClosed)
			atomic.StoreInt64(&cb.failures, 0)
			slog.Info("rate limiter circuit breaker: closed, redis recovered")
		}
	} else if state == circuitClosed {
		atomic.StoreInt64(&cb.failures, 0)
	}
}

// recordFailure 记录 Redis 操作失败。
func (cb *circuitBreaker) recordFailure() {
	atomic.AddInt64(&cb.failures, 1)
	f := atomic.LoadInt64(&cb.failures)
	state := atomic.LoadInt32(&cb.state)

	if state == circuitHalfOpen {
		atomic.StoreInt32(&cb.state, circuitOpen)
		cb.lastFailureTime = time.Now()
		atomic.StoreInt64(&cb.successes, 0)
		slog.Warn("rate limiter circuit breaker: open (half-open failure), falling back to local")
	} else if f >= cb.failureThreshold {
		if atomic.CompareAndSwapInt32(&cb.state, circuitClosed, circuitOpen) {
			cb.lastFailureTime = time.Now()
			slog.Warn("rate limiter circuit breaker: open (threshold reached), falling back to local",
				"failures", f)
		}
	}
}

// rateBucket 令牌桶，本地内存版（Redis不可用时降级使用）。
//
// 令牌桶算法说明：
//   - 桶内最多有 maxTokens 个令牌
//   - 每秒向桶内补充 refillRate 个令牌
//   - 每次请求消耗 1 个令牌，无令牌则拒绝请求
//   - 支持突发流量，令牌桶空时平滑限流
type rateBucket struct {
	tokens     float64   // 当前令牌数
	maxTokens  float64   // 桶容量（最大突发请求数）
	refillRate float64   // 每秒补充令牌数
	lastRefill time.Time // 上次补充时间
}

// RateLimitStats 限流统计数据
type RateLimitStats struct {
	TotalRequests  int64 `json:"total_requests"`  // 总请求数
	Allowed        int64 `json:"allowed"`         // 通过的请求数
	RejectedIP     int64 `json:"rejected_ip"`     // IP限流拒绝数
	RejectedAPI    int64 `json:"rejected_api"`    // 接口限流拒绝数
	RejectedUser   int64 `json:"rejected_user"`   // 用户限流拒绝数
	RedisFallbacks int64 `json:"redis_fallbacks"` // Redis降级次数
}

// RateLimiter 三层限流器：本地内存 + Redis 分布式限流 + 熔断器保护。
//
// 设计目标（风险规避）：
//  1. 网关层 IP 维度限流：防止单个 IP 恶意刷请求（防DDoS/CC）
//  2. 接口级全局限流：保护下游数据库和服务不被压垮（防雪崩）
//  3. 用户级限流：防止单个用户滥用资源（防恶意调用）
//  4. Redis 不可用时自动降级为本地内存限流（熔断器模式，自动恢复）
//  5. 白名单机制：健康检查、内部服务等特殊路径不限流
//  6. 令牌桶算法天然支持突发流量，比固定窗口/滑动窗口更平滑
//  7. 本地桶自动清理：防止内存泄漏
//  8. 限流统计：监控限流效果和系统健康状态
//  9. Key数量限制：防止不同IP生成无限key撑爆内存
//
// 限流优先级（从外到内，任意一层拒绝即返回429）：
//  1. 白名单检查
//  2. IP级限流（网关层）
//  3. 接口级限流（全局限流）
//  4. 用户级限流（登录后，JWTAuth之后）
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*rateBucket
	stopCh   chan struct{}
	once     sync.Once
	cb       *circuitBreaker // 熔断器
	stats    RateLimitStats  // 限流统计（atomic访问）

	// IP 限流配置
	ipRate     float64 // IP 每秒令牌补充速率
	ipCapacity float64 // IP 桶容量（突发上限）

	// 接口限流配置
	apiRate     float64 // 接口每秒令牌补充速率
	apiCapacity float64 // 接口桶容量（突发上限）

	// 用户限流配置
	userRate     float64 // 用户每秒令牌补充速率
	userCapacity float64 // 用户桶容量（突发上限）

	// 白名单路径前缀（不进行限流）
	whitelistPrefixes []string
	// 白名单精确路径
	whitelistPaths map[string]bool

	// 本地桶最大数量（防止内存溢出）
	maxLocalBuckets int

	// Redis 客户端（为 nil 时始终使用本地内存限流）
	redisClient *redis.Client
}

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	IPRate           float64  // IP 维度每秒请求数
	IPCapacity       float64  // IP 维度突发容量
	APIRate          float64  // 接口维度每秒请求数
	APICapacity      float64  // 接口维度突发容量
	UserRate         float64  // 用户维度每秒请求数
	UserCapacity     float64  // 用户维度突发容量
	WhitelistPaths   []string // 白名单路径（精确匹配）
	WhitelistPrefixes []string // 白名单路径前缀（前缀匹配）
	MaxLocalBuckets  int      // 本地桶最大数量，0表示使用默认值(10000)
}

// DefaultRateLimiterConfig 返回默认限流配置。
//
// 配置说明（生产级保守配置）：
//   - IP 级：30 req/s，突发 60 — 防止单 IP 恶意攻击，正常用户浏览无感知
//   - 接口级：100 req/s，突发 200 — 保护全局接口不被压垮，预留冗余
//   - 用户级：20 req/s，突发 40 — 防止单用户滥用资源，正常操作足够
//   - 白名单：/health 健康检查路径，不影响 Docker/K8s 探针
//   - 本地桶上限：10000个，防止内存溢出
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		IPRate:       30,
		IPCapacity:   60,
		APIRate:      100,
		APICapacity:  200,
		UserRate:     20,
		UserCapacity: 40,
		WhitelistPaths: []string{
			"/health",
		},
		WhitelistPrefixes: []string{
			"/ws", // WebSocket 连接有自己的频控机制
		},
		MaxLocalBuckets: 10000,
	}
}

// NewRateLimiter 创建限流器实例。
//
// 如果 Redis 可用，使用 Redis 分布式令牌桶实现多实例共享限流状态；
// 如果 Redis 不可用，自动降级为本地内存限流（单实例有效）；
// 通过熔断器自动检测 Redis 健康状态并自动恢复。
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	maxBuckets := cfg.MaxLocalBuckets
	if maxBuckets <= 0 {
		maxBuckets = 10000
	}

	rl := &RateLimiter{
		buckets:           make(map[string]*rateBucket),
		ipRate:            cfg.IPRate,
		ipCapacity:        cfg.IPCapacity,
		apiRate:           cfg.APIRate,
		apiCapacity:       cfg.APICapacity,
		userRate:          cfg.UserRate,
		userCapacity:      cfg.UserCapacity,
		whitelistPaths:    make(map[string]bool),
		whitelistPrefixes: cfg.WhitelistPrefixes,
		maxLocalBuckets:   maxBuckets,
		stopCh:            make(chan struct{}),
		redisClient:       repository.RedisClient,
		cb:                newCircuitBreaker(),
	}

	// 初始化白名单
	for _, p := range cfg.WhitelistPaths {
		rl.whitelistPaths[p] = true
	}

	// 启动后台清理协程
	go rl.cleanup()

	slog.Info("rate limiter initialized",
		"ip_rate", cfg.IPRate, "ip_cap", cfg.IPCapacity,
		"api_rate", cfg.APIRate, "api_cap", cfg.APICapacity,
		"user_rate", cfg.UserRate, "user_cap", cfg.UserCapacity,
		"redis_enabled", rl.redisClient != nil,
	)

	return rl
}

// Stop 停止限流器后台清理 goroutine，防止 goroutine 泄漏。
// 应在服务关闭时调用（通过 defer 或 fx.Lifecycle）。
func (rl *RateLimiter) Stop() {
	rl.once.Do(func() {
		close(rl.stopCh)
		slog.Info("rate limiter stopped",
			"total_requests", atomic.LoadInt64(&rl.stats.TotalRequests),
			"allowed", atomic.LoadInt64(&rl.stats.Allowed),
			"rejected_ip", atomic.LoadInt64(&rl.stats.RejectedIP),
			"rejected_api", atomic.LoadInt64(&rl.stats.RejectedAPI),
			"rejected_user", atomic.LoadInt64(&rl.stats.RejectedUser),
			"redis_fallbacks", atomic.LoadInt64(&rl.stats.RedisFallbacks),
		)
	})
}

// GetStats 获取限流统计数据快照。
func (rl *RateLimiter) GetStats() RateLimitStats {
	return RateLimitStats{
		TotalRequests:  atomic.LoadInt64(&rl.stats.TotalRequests),
		Allowed:        atomic.LoadInt64(&rl.stats.Allowed),
		RejectedIP:     atomic.LoadInt64(&rl.stats.RejectedIP),
		RejectedAPI:    atomic.LoadInt64(&rl.stats.RejectedAPI),
		RejectedUser:   atomic.LoadInt64(&rl.stats.RejectedUser),
		RedisFallbacks: atomic.LoadInt64(&rl.stats.RedisFallbacks),
	}
}

// cleanup 后台定期清理长期不活跃的本地内存桶。
//
// 清理策略：
//   - 每 5 分钟扫描一次
//   - 删除 10 分钟内无请求且令牌已满的桶（非活跃桶）
//   - 如果桶总数超过上限，优先删除最旧的非活跃桶
//
// 这样既防止内存泄漏，又保留活跃用户的限流状态。
// 注意：Redis 端限流由 TTL 自动过期，无需手动清理。
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			// 先标记过期的key
			expiredKeys := make([]string, 0)
			for key, b := range rl.buckets {
				if b.lastRefill.Before(cutoff) && b.tokens >= b.maxTokens {
					expiredKeys = append(expiredKeys, key)
				}
			}
			// 删除过期key
			for _, key := range expiredKeys {
				delete(rl.buckets, key)
			}
			// 如果桶数量仍然超限，强制清理最早的一半
			if len(rl.buckets) > rl.maxLocalBuckets {
				slog.Warn("local rate limit buckets exceeded limit, force cleaning",
					"count", len(rl.buckets), "limit", rl.maxLocalBuckets)
				// 简单策略：删除所有，让活跃key重新创建（避免复杂排序）
				// 这会短暂放通一些请求，但比内存溢出好
				rl.buckets = make(map[string]*rateBucket)
			}
			rl.mu.Unlock()
			if len(expiredKeys) > 0 {
				slog.Debug("rate limiter cleaned up expired buckets", "count", len(expiredKeys))
			}
		case <-rl.stopCh:
			slog.Debug("rate limiter cleanup goroutine stopped")
			return
		}
	}
}

// allowLocal 本地内存令牌桶限流（Redis不可用时降级使用）。
func (rl *RateLimiter) allowLocal(key string, rate, capacity float64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		// 本地桶数量上限检查，防止key爆炸
		if len(rl.buckets) >= rl.maxLocalBuckets {
			// 桶已满，放行请求（降级为不限流，保护服务可用性）
			slog.Warn("local rate limit buckets full, allowing request", "key", key)
			return true
		}
		b = &rateBucket{
			tokens:     capacity,
			maxTokens:  capacity,
			refillRate: rate,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = b
	}

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// allowRedis 使用 Redis 实现分布式令牌桶限流。
//
// 实现原理：
//  1. 使用 Redis Hash 存储令牌桶状态（tokens, last_refill）
//  2. 使用 Lua 脚本保证原子性，避免竞态条件
//  3. 设置 TTL 自动过期，避免 key 无限累积
//  4. 通过熔断器监控 Redis 健康状态
//
// 返回值：
//   - true: 允许请求，已消耗 1 个令牌
//   - false: 限流，请求应被拒绝
func (rl *RateLimiter) allowRedis(ctx context.Context, key string, rate, capacity float64, ttl time.Duration) bool {
	// 熔断器检查：Redis不可用时直接使用本地限流
	if !rl.cb.allowRedis() {
		atomic.AddInt64(&rl.stats.RedisFallbacks, 1)
		return rl.allowLocal(key, rate, capacity)
	}

	// Lua 脚本：原子性令牌桶操作
	// KEYS[1] = 限流 key
	// ARGV[1] = 每秒补充速率 (rate)
	// ARGV[2] = 桶容量 (capacity)
	// ARGV[3] = 当前时间戳（秒，浮点）
	// ARGV[4] = TTL 秒数
	script := `
	local key = KEYS[1]
	local rate = tonumber(ARGV[1])
	local capacity = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local ttl = tonumber(ARGV[4])

	local data = redis.call('HMGET', key, 'tokens', 'last_refill')
	local tokens = tonumber(data[1])
	local lastRefill = tonumber(data[2])

	if tokens == nil then
		tokens = capacity
		lastRefill = now
	else
		local elapsed = now - lastRefill
		tokens = tokens + elapsed * rate
		if tokens > capacity then
			tokens = capacity
		end
		lastRefill = now
	end

	if tokens < 1 then
		redis.call('HMSET', key, 'tokens', tokens, 'last_refill', lastRefill)
		redis.call('EXPIRE', key, ttl)
		return 0
	end

	tokens = tokens - 1
	redis.call('HMSET', key, 'tokens', tokens, 'last_refill', lastRefill)
	redis.call('EXPIRE', key, ttl)
	return 1
	`

	now := float64(time.Now().UnixNano()) / 1e9
	result, err := rl.redisClient.Eval(ctx, script, []string{key},
		fmt.Sprintf("%.6f", rate),
		fmt.Sprintf("%.6f", capacity),
		fmt.Sprintf("%.9f", now),
		int(ttl.Seconds()),
	).Result()

	if err != nil {
		rl.cb.recordFailure()
		slog.Warn("redis rate limit failed, falling back to local", "error", err, "key", key)
		atomic.AddInt64(&rl.stats.RedisFallbacks, 1)
		return rl.allowLocal(key, rate, capacity)
	}

	rl.cb.recordSuccess()
	allowed, _ := strconv.Atoi(fmt.Sprint(result))
	return allowed == 1
}

// allow 执行限流检查，自动选择 Redis 或本地实现。
// 优先使用Redis（熔断器保护），Redis不可用时自动降级本地限流。
func (rl *RateLimiter) allow(ctx context.Context, key string, rate, capacity float64) bool {
	if rl.redisClient != nil && rl.cb.allowRedis() {
		return rl.allowRedis(ctx, key, rate, capacity, 5*time.Minute)
	}
	return rl.allowLocal(key, rate, capacity)
}

// isWhitelisted 检查路径是否在白名单中（白名单路径不限流）。
// 支持精确匹配和前缀匹配。
func (rl *RateLimiter) isWhitelisted(path string) bool {
	// 精确匹配
	if rl.whitelistPaths[path] {
		return true
	}
	// 前缀匹配
	for _, prefix := range rl.whitelistPrefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// RateLimit 创建 Gin 网关层限流中间件（IP级 + 接口级）。
//
// 此中间件应作为全局中间件使用，在 JWTAuth 之前执行。
//
// 限流层级（按顺序检查，任意一层触发即拒绝）：
//  1. 白名单检查：健康检查、WebSocket等路径直接放行
//  2. IP 维度限流：防止单 IP 恶意攻击
//  3. 接口维度限流：保护下游服务
//
// 注意：用户级限流需使用 UserRateLimit 中间件，在 JWTAuth 之后使用。
//
// 限流触发时返回 HTTP 429 Too Many Requests，
// 响应头包含 Retry-After 和限流相关信息，便于客户端实现退避重试。
func RateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		atomic.AddInt64(&limiter.stats.TotalRequests, 1)

		// 白名单路径直接放行
		if limiter.isWhitelisted(path) {
			c.Next()
			return
		}

		// 第 1 层：IP 维度限流（网关防护）
		ip := c.ClientIP()
		// 防止伪造X-Forwarded-For导致的key过长，简单截断
		if len(ip) > 64 {
			ip = ip[:64]
		}
		ipKey := rateLimitIPPrefix + ip
		if !limiter.allow(c.Request.Context(), ipKey, limiter.ipRate, limiter.ipCapacity) {
			atomic.AddInt64(&limiter.stats.RejectedIP, 1)
			slog.Warn("ip rate limit exceeded", "ip", ip, "path", path)
			retryAfter := 1 // 建议1秒后重试
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", limiter.ipCapacity))
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
				"type":    "ip_rate_limit",
			})
			return
		}

		// 第 2 层：接口维度限流（method + path，全局限流）
		// 注意：对于带参数的路径（如 /api/users/1），统一归一化为 /api/users/:id
		// 避免不同ID生成不同key导致限流失效
		normalizedPath := normalizePath(path)
		apiKey := rateLimitAPIPrefix + c.Request.Method + ":" + normalizedPath
		if !limiter.allow(c.Request.Context(), apiKey, limiter.apiRate, limiter.apiCapacity) {
			atomic.AddInt64(&limiter.stats.RejectedAPI, 1)
			slog.Warn("api rate limit exceeded", "path", path, "method", c.Request.Method)
			retryAfter := 1
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", limiter.apiCapacity))
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "服务器繁忙，请稍后再试",
				"type":    "api_rate_limit",
			})
			return
		}

		atomic.AddInt64(&limiter.stats.Allowed, 1)
		c.Next()
	}
}

// UserRateLimit 用户级限流中间件。
//
// 此中间件必须在 JWTAuth 之后使用，依赖 Context 中的 "username" 字段。
// 用于防止单个用户滥用 API 资源。
//
// 限流触发时返回 HTTP 429 Too Many Requests。
func UserRateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			// 没有用户信息则跳过用户级限流（应由外层中间件处理未认证情况）
			c.Next()
			return
		}

		// 白名单路径跳过用户级限流
		path := c.Request.URL.Path
		if limiter.isWhitelisted(path) {
			c.Next()
			return
		}

		userStr, ok := username.(string)
		if !ok || userStr == "" {
			c.Next()
			return
		}

		userKey := rateLimitUserPrefix + userStr
		if !limiter.allow(c.Request.Context(), userKey, limiter.userRate, limiter.userCapacity) {
			atomic.AddInt64(&limiter.stats.RejectedUser, 1)
			slog.Warn("user rate limit exceeded", "user", userStr, "path", path)
			retryAfter := 2 // 用户级限流建议等待更长时间
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", limiter.userCapacity))
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "操作过于频繁，请稍后再试",
				"type":    "user_rate_limit",
			})
			return
		}

		c.Next()
	}
}

// normalizePath 归一化请求路径，将数字ID替换为:id占位符。
// 避免 /users/1、/users/2 等不同ID生成不同的限流key。
//
// 示例：
//
//	/api/users/123 -> /api/users/:id
//	/api/audit/applications/456 -> /api/audit/applications/:id
//	/api/resources -> /api/resources（不变）
func normalizePath(path string) string {
	// 简单的路径分段处理，将纯数字段替换为 :id
	parts := make([]byte, 0, len(path))
	segmentStart := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			segment := path[segmentStart:i]
			if segment != "" && isNumericSegment(segment) {
				parts = append(parts, ':', 'i', 'd')
			} else {
				parts = append(parts, segment...)
			}
			if i < len(path) {
				parts = append(parts, '/')
			}
			segmentStart = i + 1
		}
	}
	// 处理尾部的段
	if segmentStart < len(path) {
		segment := path[segmentStart:]
		if isNumericSegment(segment) {
			parts = append(parts, ':', 'i', 'd')
		} else {
			parts = append(parts, segment...)
		}
	}
	return string(parts)
}

// isNumericSegment 检查路径段是否为纯数字ID。
func isNumericSegment(s string) bool {
	if len(s) == 0 {
		return false
	}
	// UUID不是纯数字（包含-和字母），不替换
	// 只匹配纯数字的ID
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	// 长度至少1，最多20（int64范围内）
	return len(s) >= 1 && len(s) <= 20
}

// DefaultAPIRateLimiter 返回默认 API 限流器实例。
//
// 默认配置（三层防护 + 熔断器）：
//   - IP 级：30 req/s，突发 60 — 防止单 IP 恶意攻击
//   - 接口级：100 req/s，突发 200 — 全局限流保护下游
//   - 用户级：20 req/s，突发 40 — 防止单用户滥用资源
//
// 风险规避：
//   - Redis 不可用自动降级为本地内存限流（熔断器自动恢复）
//   - 健康检查和WebSocket路径白名单，不影响 Docker/K8s 探针
//   - 令牌桶算法平滑限流，支持合理突发流量
//   - 本地桶数量限制，防止内存溢出
//   - 路径归一化，避免不同ID绕过限流
//   - 限流统计监控，便于观测系统状态
func DefaultAPIRateLimiter() *RateLimiter {
	return NewRateLimiter(DefaultRateLimiterConfig())
}

// 确保 math 包被引用（防止未来扩展时的编译问题）
var _ = math.Ceil
