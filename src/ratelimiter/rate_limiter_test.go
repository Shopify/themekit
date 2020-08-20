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

func TestRateLimiterResetAfter(t *testing.T) {
	limiter := New("domain.com", 1)
	assert.Equal(t, limiter.rate.Limit(), rate.Limit(1))
	limiter.ResetAfter(time.Millisecond)
	assert.Equal(t, limiter.rate.Limit(), rate.Limit(0))
	time.Sleep(2 * time.Millisecond)
	assert.Equal(t, limiter.rate.Limit(), rate.Limit(1))
}
