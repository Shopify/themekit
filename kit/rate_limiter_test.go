package kit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterForDomain(t *testing.T) {
	limiter1 := rateLimitFor("domain.com")
	limiter2 := rateLimitFor("domain.com")
	limiter3 := rateLimitFor("otherdomain.com")
	assert.Equal(t, limiter1, limiter2)
	assert.NotEqual(t, limiter2, limiter3)
}

func TestRateLimiterHalts(t *testing.T) {
	limiter := rateLimitFor("domain.com")
	timeout := false
	received := false
	select {
	case <-limiter.nextChan:
		received = true
	case <-time.Tick(time.Millisecond / 2):
		timeout = true
	}
	assert.Equal(t, true, timeout)
	assert.Equal(t, false, received)
}

func TestRateLimiterCanGoAfterTimeout(t *testing.T) {
	timeout := false
	received := false
	limiter := &rateLimiter{nextChan: make(chan bool)}
	limiter.next()
	select {
	case <-limiter.nextChan:
		received = true
	case <-time.Tick(3 * time.Second):
		timeout = true
	}
	assert.Equal(t, false, timeout)
	assert.Equal(t, true, received)
}
