package kit

import (
	"sync"
	"time"
)

// foreman is a job queueing processor using a leakyBucket throttler.
type foreman struct {
	leakyBucket *leakyBucket
	halt        chan bool
	JobQueue    chan AssetEvent
	WorkerQueue chan AssetEvent
	OnIdle      func()
}

// newForeman will return a new foreman using the bucket for throttling.
func newForeman(lb *leakyBucket) *foreman {
	newForeman := &foreman{
		leakyBucket: lb,
		halt:        make(chan bool),
		JobQueue:    make(chan AssetEvent),
		WorkerQueue: make(chan AssetEvent),
		OnIdle:      func() {},
	}
	newForeman.IssueWork()
	return newForeman
}

func (f *foreman) Restart() {
	f.Halt()
	f.leakyBucket.TopUp()
	f.IssueWork()
}

// IssueWork start the foreman processing jobs that are in it's queue. It will call
// OnIdle every second when there is no jobs to process. If there are jobs in the queue
// then it will make sure there is a worker to process it from the bucket.
func (f *foreman) IssueWork() {
	f.leakyBucket.StartDripping()
	go func() {
		var waitGroup sync.WaitGroup
		notifyProcessed := false
		for {
			select {
			case job, more := <-f.JobQueue:
				if !more {
					waitGroup.Wait()
					close(f.WorkerQueue)
					return
				}
				f.leakyBucket.GetDrop()
				notifyProcessed = true
				waitGroup.Add(1)
				go func(jobToAdd AssetEvent, wg *sync.WaitGroup) {
					f.WorkerQueue <- jobToAdd
					wg.Done()
				}(job, &waitGroup)
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

// Halt stops the foreman from processing jobs in its queue.
func (f *foreman) Halt() {
	f.leakyBucket.StopDripping()
	go func() {
		f.halt <- true
	}()
}
