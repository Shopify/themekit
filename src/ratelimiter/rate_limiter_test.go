package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiterForDomain(t *testing.T) {
	limiter1 := New("domain.com", 1)
	limiter2 := New("domain.com", 2)
	limiter3 := New("otherdomain.com", 2)
	assert.Equal(t, limiter1, limiter2)
	assert.NotEqual(t, limiter2, limiter3)
}

func TestRateLimiterLockUnlock(t *testing.T) {
	limiter := New("domain.com", 1)
	assert.Equal(t, limiter.rate.Limit(), rate.Limit(1))
	limiter.lock()
	<-limiter.ctx.Done()
	assert.Equal(t, limiter.rate.Limit(), rate.Limit(0))
	limiter.unlock()
	assert.Equal(t, limiter.rate.Limit(), rate.Limit(1))
	assert.Nil(t, limiter.ctx.Err())
	limiter.unlock()
}

func TestRateLimiterRetryAfter(t *testing.T) {
	limiter := New("domain.com", 1)
	expected := time.Now().Add(2 * time.Second)
	limiter.retryAfter("2.0")
	after := time.Now()
	assert.True(t, after.After(expected) || after.Equal(expected))
}
