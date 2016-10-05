package kit

import (
	"time"
)

type leakyBucket struct {
	Size        int
	Refill      int
	Duration    time.Duration
	bucket      chan (bool)
	stopFilling chan (bool)
	ticker      *time.Ticker
}

func newLeakyBucket(size, refill, duration int) *leakyBucket {
	dur := time.Duration(duration) * time.Second
	newBucket := &leakyBucket{
		Size:        size,
		Refill:      refill,
		Duration:    dur,
		bucket:      make(chan bool, size),
		stopFilling: make(chan bool),
		ticker:      time.NewTicker(dur),
	}
	return newBucket
}

func (b *leakyBucket) StartDripping() {
	go func() {
		for {
			select {
			case <-b.stopFilling:
				return
			case <-b.ticker.C:
				b.fill(b.Refill)
			}
		}
	}()
}

func (b *leakyBucket) StopDripping() {
	go func() {
		b.stopFilling <- true
	}()
}

func (b *leakyBucket) Available() int {
	return len(b.bucket)
}

func (b *leakyBucket) IsEmpty() bool {
	return len(b.bucket) == 0
}

func (b *leakyBucket) IsFull() bool {
	return len(b.bucket) >= cap(b.bucket)
}

func (b *leakyBucket) TopUp() {
	b.fill(b.Size)
}

func (b *leakyBucket) AddDrops() {
	b.fill(b.Refill)
}

func (b *leakyBucket) GetDrop() {
	<-b.bucket
}

func (b *leakyBucket) fill(amount int) {
	for i := 0; i < amount; i++ {
		select {
		case b.bucket <- true:
		default:
			// Bucket is full, just ignore the drop
		}
	}
}
