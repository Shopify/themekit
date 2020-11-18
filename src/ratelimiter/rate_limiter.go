package ratelimiter

import (
	"bytes"
	"context"
	"errors"
	"golang.org/x/time/rate"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var domainLimitMap = make(map[string]*Limiter)

// Limiter keeps track of an api rate limit and wont let you pass the limit
type Limiter struct {
	perSecond rate.Limit
	rate      *rate.Limiter
	waiting   chan int
	ctx       context.Context
	cancel    context.CancelFunc
	locked    bool
}

// New creates a new call rate limiter for a single domain
func New(domain string, reqPerSec int) *Limiter {
	if _, ok := domainLimitMap[domain]; !ok {
		everySecond := rate.Every(time.Second / time.Duration(reqPerSec))
		ctx, cancel := context.WithCancel(context.Background())
		domainLimitMap[domain] = &Limiter{
			perSecond: everySecond,
			rate:      rate.NewLimiter(everySecond, reqPerSec),
			ctx:       ctx,
			cancel:    cancel,
		}
	}
	return domainLimitMap[domain]
}

// GateReq will make the http request but will force it to comply with concurrent limits,
// rate limits, and it will also retry requests that receive 429.
// When a 429 occurs, it will cancel all inflight requests and pauses, so that the requests
// dont continue to batter the server and cause bot detection
func (limiter *Limiter) GateReq(client *http.Client, origReq *http.Request, body []byte) (*http.Response, error) {
	limiter.rate.Wait(context.Background())
	req := origReq.WithContext(limiter.ctx)
	// reset the body when non-nil for every request (rewind)
	if len(body) > 0 {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusTooManyRequests {
		limiter.retryAfter(resp.Header.Get("Retry-After"))
		return limiter.GateReq(client, origReq, body)
	} else if errors.Is(err, context.Canceled) {
		<-limiter.waiting
		return limiter.GateReq(client, origReq, body)
	}
	return resp, err
}

func (limiter *Limiter) retryAfter(header string) {
	limiter.lock()
	defer limiter.unlock()
	after, _ := strconv.ParseFloat(header, 10)
	time.Sleep(time.Duration(after) * time.Second)
}

func (limiter *Limiter) lock() {
	if limiter.locked {
		return
	}
	limiter.locked = true
	limiter.waiting = make(chan int)
	limiter.rate.SetLimit(0)
	limiter.cancel()
}

func (limiter *Limiter) unlock() {
	if !limiter.locked {
		return
	}
	limiter.rate.SetLimit(limiter.perSecond)
	limiter.ctx, limiter.cancel = context.WithCancel(context.Background())
	close(limiter.waiting)
	limiter.locked = false
}
