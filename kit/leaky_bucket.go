package kit

import (
	"time"
)

type LeakyBucket struct {
	Size        int
	Refill      int
	Duration    time.Duration
	bucket      chan (bool)
	stopFilling chan (bool)
	ticker      *time.Ticker
}

func NewLeakyBucket(size, refill, duration int) *LeakyBucket {
	dur := time.Duration(duration) * time.Second

	return &LeakyBucket{
		Size:        size,
		Refill:      refill,
		Duration:    dur,
		bucket:      make(chan bool, size),
		stopFilling: make(chan bool),
		ticker:      time.NewTicker(dur),
	}
}

func (b *LeakyBucket) StartDripping() {
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

func (b *LeakyBucket) StopDripping() {
	go func() {
		b.stopFilling <- true
	}()
}

func (b *LeakyBucket) Available() int {
	return len(b.bucket)
}

func (b *LeakyBucket) IsEmpty() bool {
	return len(b.bucket) == 0
}

func (b *LeakyBucket) IsFull() bool {
	return len(b.bucket) >= cap(b.bucket)
}

func (b *LeakyBucket) TopUp() {
	b.fill(b.Size)
}

func (b *LeakyBucket) AddDrops() {
	b.fill(b.Refill)
}

func (b *LeakyBucket) GetDrop() {
	<-b.bucket
}

func (b *LeakyBucket) fill(amount int) {
	for i := 0; i < amount; i++ {
		select {
		case b.bucket <- true:
		default:
			// Bucket is full, just ignore the drop
		}
	}
}
