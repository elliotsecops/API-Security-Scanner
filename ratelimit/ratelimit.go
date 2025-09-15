package ratelimit

import (
	"sync"
	"time"

	"api-security-scanner/logging"
)

// RateLimiter controls the rate of requests
type RateLimiter struct {
	// For token bucket algorithm
	mu            sync.Mutex
	tokens        int
	maxTokens     int
	tokensPerSec  int
	lastAddTokens time.Time

	// For concurrent request limiting
	concurrentLimiter chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond, maxConcurrentRequests int) *RateLimiter {
	// Apply default values if not specified
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10 // Default value
	}

	if maxConcurrentRequests <= 0 {
		maxConcurrentRequests = 5 // Default value
	}

	rl := &RateLimiter{
		tokens:            requestsPerSecond, // Start with a full bucket
		maxTokens:         requestsPerSecond,
		tokensPerSec:      requestsPerSecond,
		lastAddTokens:     time.Now(),
		concurrentLimiter: make(chan struct{}, maxConcurrentRequests),
	}

	logging.Info("Rate limiter initialized", map[string]interface{}{
		"requests_per_second":      requestsPerSecond,
		"max_concurrent_requests":  maxConcurrentRequests,
	})

	return rl
}

// Wait blocks until a token is available and a concurrent slot is available
func (rl *RateLimiter) Wait() {
	// First, wait for a concurrent slot
	rl.concurrentLimiter <- struct{}{}

	// Then, wait for a rate limit token
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Add tokens based on time passed
	now := time.Now()
	secondsPassed := now.Sub(rl.lastAddTokens).Seconds()
	tokensToAdd := int(secondsPassed * float64(rl.tokensPerSec))
	
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastAddTokens = now
	}

	// If we have tokens, consume one
	if rl.tokens > 0 {
		rl.tokens--
		return
	}

	// If no tokens, calculate wait time and sleep
	// We need to wait for the next token
	timeToWait := time.Duration(1.0/float64(rl.tokensPerSec)*float64(time.Second)) - now.Sub(rl.lastAddTokens)%time.Duration(1.0/float64(rl.tokensPerSec)*float64(time.Second))
	rl.mu.Unlock()
	time.Sleep(timeToWait)
	rl.mu.Lock()
	
	// Add tokens again after waiting
	now = time.Now()
	secondsPassed = now.Sub(rl.lastAddTokens).Seconds()
	tokensToAdd = int(secondsPassed * float64(rl.tokensPerSec))
	
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastAddTokens = now
	}
	
	// Consume one token
	if rl.tokens > 0 {
		rl.tokens--
	}
}

// Done signals that a request has completed
func (rl *RateLimiter) Done() {
	// Free up the concurrent slot
	<-rl.concurrentLimiter
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	return map[string]interface{}{
		"current_tokens":          rl.tokens,
		"max_tokens":              rl.maxTokens,
		"tokens_per_second":       rl.tokensPerSec,
		"concurrent_slots_total":  cap(rl.concurrentLimiter),
		"concurrent_slots_in_use": len(rl.concurrentLimiter),
	}
}