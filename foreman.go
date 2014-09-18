package phoenix

type Foreman struct {
	bucket      *LeakyBucket
	halt        chan bool
	JobQueue    chan AssetEvent
	WorkerQueue chan AssetEvent
}

func NewForeman(bucket *LeakyBucket) Foreman {
	return Foreman{
		bucket:      bucket,
		halt:        make(chan bool),
		JobQueue:    make(chan AssetEvent),
		WorkerQueue: make(chan AssetEvent),
	}
}

func (f Foreman) IssueWork() {
	f.bucket.StartDripping()
	go func() {
		for {
			select {
			case job := <-f.JobQueue:
				f.bucket.GetDrop()
				go func() {
					f.WorkerQueue <- job
				}()
			case <-f.halt:
				return
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
