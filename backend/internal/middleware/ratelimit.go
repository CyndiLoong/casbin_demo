// Package middleware 提供 HTTP 中间件，包括 JWT 认证、Casbin 权限校验、限流、Panic 恢复等。
package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateBucket 令牌桶，单个 IP 的限流状态。
//
// 令牌桶算法说明：
//   - 桶内最多有 maxTokens 个令牌
//   - 每秒向桶内补充 refillRate 个令牌
//   - 每次请求消耗 1 个令牌，无令牌则拒绝请求
type rateBucket struct {
	tokens     float64   // 当前令牌数
	maxTokens  float64   // 桶容量（最大突发请求数）
	refillRate float64   // 每秒补充令牌数
	lastRefill time.Time // 上次补充时间
}

// RateLimiter IP 维度令牌桶限流器。
//
// 设计目的：作为缓存穿透的第一道防线，限制单个 IP 的请求频率，
// 防止恶意攻击者用大量不存在的 ID 穿透缓存直打数据库。
//
// 内存管理：后台定期清理长期无请求的 IP 桶，避免内存无限增长。
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*rateBucket
	rate     float64       // 每秒补充令牌数
	capacity float64       // 桶容量（突发上限）
	window   time.Duration // 清理周期
	stopCh   chan struct{} // 停止信号，用于优雅退出 cleanup goroutine
	once     sync.Once     // 确保 Stop 只执行一次
}

// NewRateLimiter 创建令牌桶限流器。
//
// 参数：
//   - ratePerSec: 每秒允许的平均请求数（令牌补充速率）
//   - capacity: 突发请求上限（桶容量），瞬间最多允许 capacity 个请求
func NewRateLimiter(ratePerSec float64, capacity float64) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*rateBucket),
		rate:     ratePerSec,
		capacity: capacity,
		window:   time.Minute,
		stopCh:   make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Stop 停止限流器后台清理 goroutine，防止 goroutine 泄漏。
// 应在服务关闭时调用（通过 defer）。
func (rl *RateLimiter) Stop() {
	rl.once.Do(func() {
		close(rl.stopCh)
	})
}

// cleanup 后台定期清理长期不活跃的 IP 桶。
//
// 清理策略：每 5 分钟扫描一次，删除 10 分钟内无请求且令牌已满的桶。
// 这样既防止内存泄漏，又保留活跃用户的限流状态。
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for ip, b := range rl.buckets {
				if b.lastRefill.Before(cutoff) && b.tokens >= b.maxTokens {
					delete(rl.buckets, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCh:
			slog.Debug("rate limiter cleanup goroutine stopped")
			return
		}
	}
}

// Allow 检查指定 IP 是否允许请求（消耗一个令牌）。
//
// 返回值：
//   - true: 允许请求，已消耗 1 个令牌
//   - false: 限流，请求应被拒绝
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[ip]
	if !ok {
		b = &rateBucket{
			tokens:     rl.capacity,
			maxTokens:  rl.capacity,
			refillRate: rl.rate,
			lastRefill: time.Now(),
		}
		rl.buckets[ip] = b
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

// RateLimit 创建 Gin 限流中间件。
//
// 限流触发时返回 HTTP 429 Too Many Requests，body 包含标准化错误信息。
func RateLimit(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			slog.Warn("rate limit exceeded", "ip", ip, "path", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// DefaultAPIRateLimiter 返回默认 API 限流器配置：20 req/s，突发 50。
//
// 配置说明：
//   - 平均每秒允许 20 个请求（足够正常业务使用）
//   - 瞬间最多允许 50 个并发请求（应对突发流量）
func DefaultAPIRateLimiter() *RateLimiter {
	return NewRateLimiter(20, 50)
}
