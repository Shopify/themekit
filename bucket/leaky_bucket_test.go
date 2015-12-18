package bucket

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFillingABucket(t *testing.T) {
	bucketSize := 2
	refillAmount := 1
	refillDuration := 10
	config := Configuration{Size: bucketSize, Refill: refillAmount, Duration: time.Duration(refillDuration) * time.Millisecond}
	bucket := NewLeakyBucketWithConfiguration(config)
	assert.Equal(t, true, bucket.IsEmpty(), "The bucket should be empty")
	assert.Equal(t, 0, bucket.Available())
	assert.Equal(t, false, bucket.IsFull(), "The bucket has nothing in it. Therefore it is not full")
	bucket.AddDrops()
	assert.Equal(t, config.Refill, bucket.Available())
	assert.Equal(t, false, bucket.IsEmpty(), "The bucket should have 1 drop in it but has %d instead", bucket.Available())
	assert.Equal(t, false, bucket.IsFull(), "The bucket should not be full yet")
	bucket.AddDrops()
	assert.Equal(t, true, bucket.IsFull(), "The bucket should have 2 drops and be full but instead it has %d drops", bucket.Available())
	assert.Equal(t, config.Size, bucket.Available())
}

func TestToppingUpABucket(t *testing.T) {
	config := Configuration{Size: 4, Refill: 1, Duration: time.Duration(10) * time.Millisecond}
	bucket := NewLeakyBucketWithConfiguration(config)
	bucket.TopUp()
	assert.Equal(t, 4, bucket.Available())
	assert.Equal(t, true, bucket.IsFull())
}

func TestGrabbingADropFromTheBucket(t *testing.T) {
	config := Configuration{Size: 2, Refill: 1, Duration: time.Duration(10) * time.Millisecond}
	bucket := NewLeakyBucketWithConfiguration(config)
	assert.Equal(t, 0, bucket.Available(), "Bucket should be empty")
	bucket.TopUp()
	assert.Equal(t, bucket.Size, bucket.Available(), "Bucket should be full")
	dropsChannel := make(chan bool)

	go func() {
		bucket.GetDrop()
		dropsChannel <- true
	}()
	select {
	case <-dropsChannel:
		dropsLeft := bucket.Available()
		assert.Equal(t, 1, dropsLeft, "There should be 1 drop left in the bucket but there were %d", dropsLeft)
	case <-time.After(50 * time.Millisecond):
		// Is there a way I can achieve this without having to rely on a timeout?
		t.Log("Expected a value to be written to the dropsChannel but it did not happen")
		t.Fail()
	}
	assert.Equal(t, bucket.Size-1, bucket.Available(), "Bucket should be full")
}

func TestFillingAnAlreadyFullBucket(t *testing.T) {
	config := Configuration{Size: 2, Refill: 1, Duration: time.Duration(10) * time.Millisecond}
	bucket := NewLeakyBucketWithConfiguration(config)
	bucket.TopUp()

	assert.Equal(t, true, bucket.IsFull())
	bucket.AddDrops()
	assert.Equal(t, true, bucket.IsFull())
	assert.Equal(t, 2, bucket.Available())
}

func BenchmarkBucket(b *testing.B) {
	config := Configuration{Size: 2, Refill: 1, Duration: time.Duration(1) * time.Millisecond}
	bucket := NewLeakyBucketWithConfiguration(config)
	bucket.StartDripping()
	for i := 0; i < b.N; i++ {
		bucket.GetDrop()
	}
}
