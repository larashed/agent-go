package buckets

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type bucketItem struct {
	record int
}

func (am *bucketItem) String() string {
	return strconv.Itoa(am.record)
}

func (am *bucketItem) Value() interface{} {
	return am.record
}

func TestBucket_Count(t *testing.T) {
	bucket := NewBucket()

	limit := 200

	for i := 0; i < limit; i++ {
		bucket.Add(&bucketItem{record: i})
	}

	assert.Equal(t, limit, bucket.Count())
}

func TestBucket_PullAndRemove(t *testing.T) {
	bucket := NewBucket()

	limit := 200

	for i := 0; i < limit; i++ {
		bucket.Add(&bucketItem{record: i})
	}

	bucketSize := 50
	newBucket := bucket.PullAndRemove(bucketSize)

	assert.Equal(t, 50, bucket.Items()[0].Value())
	assert.Equal(t, 199, bucket.Items()[bucket.Count()-1].Value())
	assert.Equal(t, limit-bucketSize, bucket.Count())

	// test new bucket size
	assert.Equal(t, bucketSize, newBucket.Count())
}

func TestBucket_Remove(t *testing.T) {
	bucket := NewBucket()

	limit := 200

	for i := 0; i < limit; i++ {
		bucket.Add(&bucketItem{record: i})
	}

	bucket.Remove(50)

	assert.Equal(t, 150, bucket.Count())
	assert.Equal(t, 50, bucket.Items()[0].Value())
	assert.Equal(t, 199, bucket.Items()[bucket.Count()-1].Value())
}
