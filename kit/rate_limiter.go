package kit

import (
	"time"
)

type rateLimiter struct {
	nextChan chan bool
}

var (
	apiLimit       = time.Second / 2
	domainLimitMap = make(map[string]*rateLimiter)
)

func rateLimitFor(domain string) *rateLimiter {
	if _, ok := domainLimitMap[domain]; !ok {
		domainLimitMap[domain] = &rateLimiter{nextChan: make(chan bool)}
		domainLimitMap[domain].next()
	}
	return domainLimitMap[domain]
}

func (limiter *rateLimiter) next() {
	go func() {
		select {
		case <-time.Tick(apiLimit):
			limiter.nextChan <- true
		}
	}()
}

func (limiter *rateLimiter) Wait() {
	<-limiter.nextChan
	limiter.next()
}
