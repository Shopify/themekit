package kit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterHalts(t *testing.T) {
	limiter := newRateLimiter(time.Millisecond)
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
	limiter := newRateLimiter(time.Millisecond)
	select {
	case <-limiter.nextChan:
		received = true
	case <-time.Tick(2 * time.Millisecond):
		timeout = true
	}
	assert.Equal(t, false, timeout)
	assert.Equal(t, true, received)
}
