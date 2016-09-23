package kit

import (
	"time"
)

// Foreman ... TODO
type Foreman struct {
	leakyBucket *LeakyBucket
	halt        chan bool
	JobQueue    chan AssetEvent
	WorkerQueue chan AssetEvent
	OnIdle      func()
}

// NewForeman ... TODO
func NewForeman(leakyBucket *LeakyBucket) Foreman {
	return Foreman{
		leakyBucket: leakyBucket,
		halt:        make(chan bool),
		JobQueue:    make(chan AssetEvent),
		WorkerQueue: make(chan AssetEvent),
		OnIdle:      func() {},
	}
}

// IssueWork ... TODO
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

// Halt ... TODO
func (f Foreman) Halt() {
	f.leakyBucket.StopDripping()
	go func() {
		f.halt <- true
	}()
}
