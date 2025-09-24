package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiterRespectsMaxConcurrent(t *testing.T) {
	rl := NewRateLimiter(100, 2)

	var wg sync.WaitGroup
	startGate := make(chan struct{})
	acquired := make(chan int, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-startGate
			rl.Wait()
			acquired <- id
			time.Sleep(20 * time.Millisecond)
			rl.Done()
		}(i)
	}

	close(startGate)

	first := <-acquired
	second := <-acquired
	if first == second {
		t.Fatalf("expected distinct goroutines, got duplicate %d", first)
	}

	select {
	case third := <-acquired:
		t.Fatalf("expected third goroutine to block, but it acquired slot: %d", third)
	case <-time.After(10 * time.Millisecond):
	}

	wg.Wait()
}

func TestRateLimiterEnforcesRate(t *testing.T) {
	rl := NewRateLimiter(1, 1)

	rl.Wait()
	rl.Done()

	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)
	rl.Done()

	if elapsed < 900*time.Millisecond {
		t.Fatalf("expected Wait to block for rate limiting, elapsed %v", elapsed)
	}
}
