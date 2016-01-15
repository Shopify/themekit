package bucket

import (
	"time"
)

// LeakyBucketConfiguration configures the size and refill rate of the bucket.
type Configuration struct {
	Size, Refill int
	Duration     time.Duration
}

type LeakyBucket struct {
	Configuration
	bucket      chan (bool)
	stopFilling chan (bool)
	ticker      *time.Ticker
}

func NewLeakyBucket(size, refill, duration int) *LeakyBucket {
	return NewLeakyBucketWithConfiguration(Configuration{Size: size, Refill: refill, Duration: time.Duration(duration) * time.Second})
}

func NewLeakyBucketWithConfiguration(configuration Configuration) *LeakyBucket {
	b := &LeakyBucket{Configuration: configuration}
	b.bucket = make(chan bool, b.Size)
	b.stopFilling = make(chan bool)
	b.ticker = time.NewTicker(b.Duration)
	return b
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
