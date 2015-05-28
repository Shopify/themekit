package themekit

import "time"

type Foreman struct {
	bucket      *LeakyBucket
	halt        chan bool
	JobQueue    chan AssetEvent
	WorkerQueue chan AssetEvent
	OnIdle      func()
}

func NewForeman(bucket *LeakyBucket) Foreman {
	return Foreman{
		bucket:      bucket,
		halt:        make(chan bool),
		JobQueue:    make(chan AssetEvent),
		WorkerQueue: make(chan AssetEvent),
		OnIdle:      func() {},
	}
}

func (f Foreman) IssueWork() {
	f.bucket.StartDripping()
	go func() {
		notifyProcessed := false
		for {
			select {
			case job := <-f.JobQueue:
				f.bucket.GetDrop()
				notifyProcessed = true
				go func() {
					f.WorkerQueue <- job
				}()
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
	f.bucket.StopDripping()
	go func() {
		f.halt <- true
	}()
}
