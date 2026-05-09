package ratelimit

import (
	"sync"
	"time"

	"api-security-scanner/logging"
)

// RateLimiter controls the rate of requests using a token bucket algorithm
type RateLimiter struct {
	mu                sync.Mutex
	tokens            float64
	maxTokens         float64
	tokensPerSecond   float64
	lastRefill        time.Time
	concurrentLimiter chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond, maxConcurrentRequests int) *RateLimiter {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10
	}

	if maxConcurrentRequests <= 0 {
		maxConcurrentRequests = 5
	}

	rl := &RateLimiter{
		tokens:            float64(requestsPerSecond),
		maxTokens:         float64(requestsPerSecond),
		tokensPerSecond:   float64(requestsPerSecond),
		lastRefill:        time.Now(),
		concurrentLimiter: make(chan struct{}, maxConcurrentRequests),
	}

	logging.Info("Rate limiter initialized", map[string]interface{}{
		"requests_per_second":     requestsPerSecond,
		"max_concurrent_requests": maxConcurrentRequests,
	})

	return rl
}

// refill adds tokens based on elapsed time
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.tokensPerSecond
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now
}

// Wait blocks until a token is available and a concurrent slot is available
func (rl *RateLimiter) Wait() {
	// First, wait for a concurrent slot
	rl.concurrentLimiter <- struct{}{}

	// Then, wait for a rate limit token
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1 {
		rl.tokens--
		return
	}

	// Calculate wait time for next token
	tokensNeeded := 1 - rl.tokens
	waitDuration := time.Duration(tokensNeeded / rl.tokensPerSecond * float64(time.Second))
	if waitDuration < time.Millisecond {
		waitDuration = time.Millisecond
	}

	// Use a timer instead of sleeping with lock held
	timer := time.NewTimer(waitDuration)
	defer timer.Stop()

	// Release lock while waiting
	rl.mu.Unlock()
	<-timer.C
	rl.mu.Lock()

	// Refill again after waiting
	rl.refill()
	if rl.tokens >= 1 {
		rl.tokens--
	}
}

// Done signals that a request has completed
func (rl *RateLimiter) Done() {
	<-rl.concurrentLimiter
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	return map[string]interface{}{
		"current_tokens":          rl.tokens,
		"max_tokens":              rl.maxTokens,
		"tokens_per_second":       rl.tokensPerSecond,
		"concurrent_slots_total":  cap(rl.concurrentLimiter),
		"concurrent_slots_in_use": len(rl.concurrentLimiter),
	}
}
