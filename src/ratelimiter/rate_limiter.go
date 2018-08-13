package ratelimiter

import "time"

var domainLimitMap = make(map[string]*Limiter)

// Limiter keeps track of an api rate limit and wont let you pass the limit
type Limiter struct {
	nextChan chan bool
	apiLimit time.Duration
}

// New creates a new call rate limiter for a single domain
func New(domain string, apiLimit time.Duration) *Limiter {
	if _, ok := domainLimitMap[domain]; !ok {
		domainLimitMap[domain] = &Limiter{
			nextChan: make(chan bool),
			apiLimit: apiLimit,
		}
		domainLimitMap[domain].next()
	}
	return domainLimitMap[domain]
}

func (limiter *Limiter) next() {
	go func(l *Limiter) {
		ticker := time.NewTimer(l.apiLimit)
		<-ticker.C
		ticker.Stop()
		l.nextChan <- true
	}(limiter)
}

// Wait will block until enough time has passed and the limit will not be passed
func (limiter *Limiter) Wait() {
	<-limiter.nextChan
	limiter.next()
}
