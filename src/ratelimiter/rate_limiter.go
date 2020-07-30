package ratelimiter

import (
	"context"
	"golang.org/x/time/rate"
	"time"
)

var domainLimitMap = make(map[string]*Limiter)

// Limiter keeps track of an api rate limit and wont let you pass the limit
type Limiter struct {
	perSecond rate.Limit
	rate      *rate.Limiter
}

// New creates a new call rate limiter for a single domain
func New(domain string, reqPerSec int) *Limiter {
	if _, ok := domainLimitMap[domain]; !ok {
		everySecond := rate.Every(time.Second / time.Duration(reqPerSec))
		domainLimitMap[domain] = &Limiter{
			perSecond: everySecond,
			rate:      rate.NewLimiter(everySecond, 10),
		}
	}
	return domainLimitMap[domain]
}

// ResetAfter will reset the bucket to 0, wait for the amount of time until it resumes
// This will allow the rate limiter to stop all activity and restart slowly
func (limiter *Limiter) ResetAfter(after time.Duration) {
	if limiter.rate.Limit() == 0 {
		return
	}
	limiter.rate.SetLimit(0)
	go func() {
		time.Sleep(after)
		limiter.rate.SetLimit(limiter.perSecond)
	}()
}

// Wait will block until enough time has passed and the limit will not be passed
func (limiter *Limiter) Wait() {
	limiter.rate.Wait(context.Background())
}
