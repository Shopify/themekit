package themekit

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFillingABucket(t *testing.T) {
	bucketSize := 2
	refillAmount := 1
	refillDuration := 0
	bucket := NewLeakyBucket(bucketSize, refillAmount, refillDuration)
	assert.Equal(t, true, bucket.IsEmpty(), "The bucket should be empty")
	assert.Equal(t, false, bucket.IsFull(), "The bucket has nothing in it. Therefore it is not full")
	bucket.AddDrops()
	assert.Equal(t, false, bucket.IsEmpty(), "The bucket should have 1 drop in it but has %d instead", bucket.UnsafeAvailable())
	assert.Equal(t, false, bucket.IsFull(), "The bucket should not be full yet")
	bucket.AddDrops()
	assert.Equal(t, true, bucket.IsFull(), "The bucket should have 2 drops and be full but instead it has %d drops", bucket.UnsafeAvailable())
}

func TestToppingUpABucket(t *testing.T) {
	bucket := NewLeakyBucket(4, 0, 0)
	bucket.TopUp()
	assert.Equal(t, 4, bucket.UnsafeAvailable())
	assert.Equal(t, true, bucket.IsFull())
}

func TestGrabbingADropFromTheBucket(t *testing.T) {
	bucketSize := 2
	bucket := NewLeakyBucket(bucketSize, 0, 0)
	bucket.TopUp()
	dropsChannel := make(chan bool)

	go func() {
		bucket.GetDrop()
		dropsChannel <- true
	}()
	select {
	case <-dropsChannel:
		dropsLeft := bucket.UnsafeAvailable()
		assert.Equal(t, 1, dropsLeft, "There should be 1 drop left in the bucket but there were %d", dropsLeft)
	case <-time.After(50 * time.Millisecond):
		// Is there a way I can achieve this without having to rely on a timeout?
		t.Log("Expected a value to be written to the dropsChannel but it did not happen")
		t.Fail()
	}
}

func TestFillingAnAlreadyFullBucket(t *testing.T) {
	bucketSize := 2
	bucket := NewLeakyBucket(bucketSize, 0, 0)
	bucket.TopUp()

	assert.Equal(t, true, bucket.IsFull())
	bucket.AddDrops()
	assert.Equal(t, true, bucket.IsFull())
	assert.Equal(t, 2, bucket.UnsafeAvailable())
}
