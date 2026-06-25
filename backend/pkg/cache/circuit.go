// Package cache 提供 Redis 缓存封装，内置缓存三大问题完整防护。
//
// 本文件实现熔断器（CircuitBreaker），用于防止 Redis 故障引发缓存雪崩。
package cache

import (
	"log/slog"
	"sync"
	"time"
)

// CircuitBreaker 熔断器，用于防止 Redis 故障引发缓存雪崩。
//
// 背景：当 Redis 连续失败（如宕机、网络分区、超时）时，如果所有请求都继续
// 尝试访问 Redis 然后失败降级到 DB，会导致：
//  1. 大量请求等待 Redis 超时（浪费连接和 goroutine）
//  2. DB 压力骤增可能引发级联故障
//
// 熔断器模式（Circuit Breaker）模拟电路保险丝：
//   - Closed（关闭）：正常放行请求，统计失败次数
//   - Open（打开）：连续失败达到阈值后"熔断"，快速失败直接降级，不访问 Redis
//   - Half-Open（半开）：熔断一段时间后放一个"探测请求"，成功则关闭，失败则重新打开
//
// 状态转换：
//
//	Closed ──(连续失败≥maxFailures)──→ Open
//	Open ──(经过resetTimeout)──→ Half-Open
//	Half-Open ──(探测成功)──→ Closed
//	Half-Open ──(探测失败)──→ Open
type CircuitBreaker struct {
	mu           sync.Mutex
	failures     int           // 当前连续失败次数
	maxFailures  int           // 熔断阈值（达到此值打开熔断器）
	open         bool          // 熔断器是否打开
	openUntil    time.Time     // 打开截止时间（半开探测时机）
	resetTimeout time.Duration // 熔断持续时间（多久后进入半开）
	semiOpen     bool          // 是否处于半开状态
	name         string        // 熔断器名称（用于日志区分）
}

// NewCircuitBreaker 创建熔断器。
//
// 参数：
//   - name: 熔断器名称（日志标识）
//   - maxFailures: 连续失败多少次后熔断
//   - resetTimeout: 熔断多久后尝试半开探测
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Allow 检查是否允许请求通过（是否应尝试访问 Redis）。
//
// 返回值：
//   - true: 允许访问 Redis（Closed 或 Half-Open 状态）
//   - false: 熔断中，跳过 Redis 直接降级
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.open {
		if time.Now().After(cb.openUntil) {
			cb.semiOpen = true
			cb.open = false
			slog.Info("circuit breaker: half-open, allowing probe", "name", cb.name)
			return true
		}
		return false
	}
	return true
}

// RecordSuccess 记录一次成功请求。
//
// - Closed 状态：重置失败计数
// - Half-Open 状态：探测成功，关闭熔断器恢复正常
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.open = false
	cb.semiOpen = false
}

// RecordFailure 记录一次失败请求。
//
// - Closed 状态：失败计数+1，达到阈值则打开熔断器
// - Half-Open 状态：探测失败，立即重新打开熔断器
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.semiOpen || cb.failures >= cb.maxFailures {
		cb.open = true
		cb.semiOpen = false
		cb.openUntil = time.Now().Add(cb.resetTimeout)
		slog.Warn("circuit breaker: OPENED", "name", cb.name, "failures", cb.failures, "reset_in", cb.resetTimeout)
	}
}

// IsOpen 返回熔断器当前是否处于打开状态。
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.open
}
