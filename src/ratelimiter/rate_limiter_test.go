package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterForDomain(t *testing.T) {
	limiter1 := New("domain.com", 0)
	limiter2 := New("domain.com", 0)
	limiter3 := New("otherdomain.com", 0)
	assert.Equal(t, limiter1, limiter2)
	assert.NotEqual(t, limiter2, limiter3)
}

func TestRateLimiterHalts(t *testing.T) {
	received, timeout := checkTime(time.Second)
	assert.Equal(t, true, timeout)
	assert.Equal(t, false, received)
}

func TestRateLimiterCanGoAfterTimeout(t *testing.T) {
	received, timeout := checkTime(time.Millisecond)
	assert.Equal(t, false, timeout)
	assert.Equal(t, true, received)
}

func TestWait(t *testing.T) {
	domainLimitMap = make(map[string]*Limiter)
	limiter := New("domain.com", time.Nanosecond)
	limiter.Wait()
}

func checkTime(dur time.Duration) (received, timeout bool) {
	domainLimitMap = make(map[string]*Limiter)
	limiter := New("domain.com", dur)
	select {
	case <-limiter.nextChan:
		received = true
	case <-time.Tick(3 * time.Millisecond):
		timeout = true
	}
	return
}
