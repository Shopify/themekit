package kit

import (
	"time"
)

type rateLimiter struct {
	rate     time.Duration
	nextChan chan bool
}

func newRateLimiter(rate time.Duration) *rateLimiter {
	newLimiter := &rateLimiter{
		rate:     rate,
		nextChan: make(chan bool),
	}
	newLimiter.next()
	return newLimiter
}

func (limiter *rateLimiter) next() {
	go func() {
		select {
		case <-time.Tick(limiter.rate):
			limiter.nextChan <- true
		}
	}()
}

func (limiter *rateLimiter) Wait() {
	<-limiter.nextChan
	limiter.next()
}
