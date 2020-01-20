package buckets

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucket_GetMany(t *testing.T) {
	bucket := NewBucket()

	limit := 200

	for i := 0; i < limit; i++ {
		bucket.Add(strconv.Itoa(i))
	}

	assert.Equal(t, limit, bucket.Count())
}

func TestBucket_Get(t *testing.T) {
	bucket := NewBucket()

	limit := 200

	for i := 0; i < limit; i++ {
		bucket.Add(strconv.Itoa(i))
	}

	bucketSize := 50
	newBucket := bucket.PullAndRemove(bucketSize)

	assert.Equal(t, "50", bucket.Items()[0])
	assert.Equal(t, "199", bucket.Items()[bucket.Count()-1])
	assert.Equal(t, limit-bucketSize, bucket.Count())

	// test new bucket size
	assert.Equal(t, bucketSize, newBucket.Count())
}
