package themekit

import (
	"time"

	"github.com/Shopify/themekit/bucket"
)

type Foreman struct {
	leakyBucket *bucket.LeakyBucket
	halt        chan bool
	JobQueue    chan AssetEvent
	WorkerQueue chan AssetEvent
	OnIdle      func()
}

func NewForeman(leakyBucket *bucket.LeakyBucket) Foreman {
	return Foreman{
		leakyBucket: leakyBucket,
		halt:        make(chan bool),
		JobQueue:    make(chan AssetEvent),
		WorkerQueue: make(chan AssetEvent),
		OnIdle:      func() {},
	}
}

func (f Foreman) IssueWork() {

	f.leakyBucket.StartDripping()
	go func() {
		notifyProcessed := false
		for {
			select {
			case job := <-f.JobQueue:
				f.leakyBucket.GetDrop()
				notifyProcessed = true
				// TODO: this was a potential aliasing issue!
				go func(jobToAdd AssetEvent) {
					f.WorkerQueue <- jobToAdd
				}(job)
			case <-f.halt:
				return
			case <-time.Tick(1 * time.Second):
				if notifyProcessed {
					notifyProcessed = false
					f.OnIdle()
				}
			}
		}
	}()
}

func (f Foreman) Halt() {
	f.leakyBucket.StopDripping()
	go func() {
		f.halt <- true
	}()
}
