package buckets

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/larashed/agent-go/monitoring/metrics"
)

func TestAdd(t *testing.T) {
	limit := 5000
	bucket := newBucket(limit)

	assert.Equal(t, limit, bucket.Count())
}

func TestMerge(t *testing.T) {
	firstBucket := newBucket(100)
	secondBucket := newBucket(150)

	firstBucket.Merge(secondBucket)

	assert.Equal(t, 250, firstBucket.Count())
}

func TestDiscard(t *testing.T) {
	firstBucket := newBucket(100)
	firstBucket.Discard(100)

	assert.Equal(t, 0, firstBucket.Count())

	firstBucket = newBucket(100)
	firstBucket.Discard(200)

	assert.Equal(t, 0, firstBucket.Count())

	firstBucket = newBucket(100)
	firstBucket.Discard(40)

	assert.Equal(t, 60, firstBucket.Count())

	firstBucket = newBucket(120)
	firstBucket.Discard(40)

	assert.Equal(t, 80, firstBucket.Count())
	assert.Equal(t, newBucketRange(40, 120).String(), firstBucket.String())
}

func TestExtract(t *testing.T) {
	firstBucket := newBucket(100)
	secondBucket := firstBucket.Extract(50)

	assert.Equal(t, 50, firstBucket.Count())
	assert.Equal(t, 50, secondBucket.Count())

	firstBucket = newBucket(0)
	secondBucket = firstBucket.Extract(10)

	assert.Equal(t, 0, firstBucket.Count())
	assert.Equal(t, 0, secondBucket.Count())

	firstBucket = newBucket(100)
	secondBucket = firstBucket.Extract(110)

	assert.Equal(t, 0, firstBucket.Count())
	assert.Equal(t, 100, secondBucket.Count())
}

func newBucket(limit int) *AppMetricBucket {
	bucket := NewAppMetricBucket()

	for i := 0; i < limit; i++ {
		bucket.Add(metrics.NewAppMetric(strconv.Itoa(i)))
	}

	return bucket
}

func newBucketRange(from, to int) *AppMetricBucket {
	bucket := NewAppMetricBucket()

	for i := from; i < to; i++ {
		bucket.Add(metrics.NewAppMetric(strconv.Itoa(i)))
	}

	return bucket
}
