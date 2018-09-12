package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterForDomain(t *testing.T) {
	limiter1 := New("domain.com", time.Nanosecond)
	limiter2 := New("domain.com", time.Nanosecond)
	limiter3 := New("otherdomain.com", time.Nanosecond)
	assert.Equal(t, limiter1, limiter2)
	assert.NotEqual(t, limiter2, limiter3)
}

func TestRateLimiterHalts(t *testing.T) {
	received, timeout := checkTime("testone.com", time.Second)
	assert.Equal(t, true, timeout)
	assert.Equal(t, false, received)
}

func TestRateLimiterCanGoAfterTimeout(t *testing.T) {
	received, timeout := checkTime("testtwo.com", time.Millisecond)
	assert.Equal(t, false, timeout)
	assert.Equal(t, true, received)
}

func TestWait(t *testing.T) {
	domainLimitMap = make(map[string]*Limiter)
	limiter := New("domain.com", time.Nanosecond)
	limiter.Wait()
}

func checkTime(domain string, dur time.Duration) (received, timeout bool) {
	domainLimitMap = make(map[string]*Limiter)
	limiter := New(domain, dur)
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	select {
	case <-limiter.nextChan:
		received = true
	case <-ticker.C:
		timeout = true
	}
	return
}
