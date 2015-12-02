package bucket

import (
	"sync"
	"time"
)

type LeakyBucket struct {
	size, refill, duration, available int
	bucket                            chan (bool)
	lock                              *sync.Mutex
	stopFilling                       chan (bool)
}

func NewLeakyBucket(size, refill, duration int) *LeakyBucket {
	b := &LeakyBucket{size: size, refill: refill, duration: duration}
	b.bucket = make(chan bool, size)
	b.stopFilling = make(chan bool)
	b.lock = &sync.Mutex{}
	b.available = 0
	return b
}

func (b *LeakyBucket) StartDripping() {
	go func() {
		for {
			select {
			case <-b.stopFilling:
				return
			default:
				b.AddDrops()
			}
		}
	}()
}

func (b *LeakyBucket) StopDripping() {
	go func() {
		b.stopFilling <- true
	}()
}

func (b *LeakyBucket) UnsafeAvailable() int {
	return b.available
}

func (b *LeakyBucket) IsEmpty() (success bool) {
	result := b.locked(func() interface{} {
		return b.available <= 0
	})
	success, _ = result.(bool)
	return
}

func (b *LeakyBucket) IsFull() (success bool) {
	result := b.locked(func() interface{} {
		return b.available == b.size
	})
	success, _ = result.(bool)
	return
}

func (b *LeakyBucket) TopUp() {
	b.fillBucket(true)
}

func (b *LeakyBucket) AddDrops() {
	time.Sleep(time.Duration(b.duration) * time.Second)
	b.fillBucket(false)
}

func (b *LeakyBucket) GetDrop() {
	<-b.bucket
	b.locked(func() interface{} {
		b.available--
		return nil
	})
}

type lockedFunc func() interface{}

func (b *LeakyBucket) locked(code lockedFunc) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()
	return code()
}

func (b *LeakyBucket) fillBucket(full bool) {
	amount := b.size - b.available
	if !full && amount > b.refill {
		amount = b.refill
	}
	for i := 0; i < amount; i++ {
		b.bucket <- true
		b.locked(func() interface{} {
			b.available++
			return nil
		})
	}
}
