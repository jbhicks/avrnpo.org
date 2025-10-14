package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	ip := "192.168.1.1"

	for i := 0; i < 3; i++ {
		if !rl.Allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	if rl.Allow(ip) {
		t.Error("Request 4 should be blocked (limit exceeded)")
	}
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	if !rl.Allow(ip1) {
		t.Error("First request from IP1 should be allowed")
	}
	if !rl.Allow(ip1) {
		t.Error("Second request from IP1 should be allowed")
	}

	if !rl.Allow(ip2) {
		t.Error("First request from IP2 should be allowed")
	}
	if !rl.Allow(ip2) {
		t.Error("Second request from IP2 should be allowed")
	}

	if rl.Allow(ip1) {
		t.Error("Third request from IP1 should be blocked")
	}
	if rl.Allow(ip2) {
		t.Error("Third request from IP2 should be blocked")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond)
	ip := "192.168.1.1"

	if !rl.Allow(ip) {
		t.Error("First request should be allowed")
	}
	if !rl.Allow(ip) {
		t.Error("Second request should be allowed")
	}
	if rl.Allow(ip) {
		t.Error("Third request should be blocked")
	}

	time.Sleep(150 * time.Millisecond)

	if !rl.Allow(ip) {
		t.Error("Request after window reset should be allowed")
	}
}

func TestRateLimiter_ZeroRequests(t *testing.T) {
	rl := NewRateLimiter(0, time.Minute)
	ip := "192.168.1.1"

	if rl.Allow(ip) {
		t.Error("No requests should be allowed with rate limit of 0")
	}
}

func TestRateLimiter_HighLoad(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	ip := "192.168.1.1"

	allowed := 0
	for i := 0; i < 150; i++ {
		if rl.Allow(ip) {
			allowed++
		}
	}

	if allowed != 100 {
		t.Errorf("Expected exactly 100 requests to be allowed, got %d", allowed)
	}
}
