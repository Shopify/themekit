package kit

import (
	"time"
)

// Foreman is a job queueing processor using a LeakyBucket throttler.
type Foreman struct {
	leakyBucket *LeakyBucket
	halt        chan bool
	JobQueue    chan AssetEvent
	WorkerQueue chan AssetEvent
	OnIdle      func()
}

// NewForeman will return a new Foreman using the bucket for throttling.
func NewForeman(leakyBucket *LeakyBucket) Foreman {
	newForeman := Foreman{
		leakyBucket: leakyBucket,
		halt:        make(chan bool),
		JobQueue:    make(chan AssetEvent),
		WorkerQueue: make(chan AssetEvent),
		OnIdle:      func() {},
	}
	newForeman.IssueWork()
	return newForeman
}

func (f Foreman) Restart() {
	f.Halt()
	f.leakyBucket.TopUp()
	f.IssueWork()
}

// IssueWork start the Foreman processing jobs that are in it's queue. It will call
// OnIdle every second when there is no jobs to process. If there are jobs in the queue
// then it will make sure there is a worker to process it from the bucket.
func (f Foreman) IssueWork() {
	f.leakyBucket.StartDripping()
	go func() {
		notifyProcessed := false
		for {
			select {
			case job := <-f.JobQueue:
				f.leakyBucket.GetDrop()
				notifyProcessed = true
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

// Halt stops the Foreman from processing jobs in its queue.
func (f Foreman) Halt() {
	f.leakyBucket.StopDripping()
	go func() {
		f.halt <- true
	}()
}
