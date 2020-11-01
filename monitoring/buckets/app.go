package buckets

import (
	"strings"
	"sync"

	"github.com/larashed/agent-go/monitoring/metrics"
)

// AppMetricBucket holds application metrics
type AppMetricBucket struct {
	metrics []metrics.AppMetric
	mutex   sync.RWMutex
	Channel chan int
}

// NewAppMetricBucket creates a new `AppMetricBucket` instance
func NewAppMetricBucket() *AppMetricBucket {
	return &AppMetricBucket{
		metrics: make([]metrics.AppMetric, 0),
		mutex:   sync.RWMutex{},
		Channel: make(chan int),
	}
}

// NewBucketFromItems creates a new `AppMetricBucket` instance with given metrics
func NewBucketFromItems(items []metrics.AppMetric) *AppMetricBucket {
	return &AppMetricBucket{
		metrics: items,
		mutex:   sync.RWMutex{},
	}
}

// Merge buckets together
func (b *AppMetricBucket) Merge(bucket *AppMetricBucket) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.metrics = append(b.metrics, *bucket.All()...)
}

// Add a metric to the bucket
func (b *AppMetricBucket) Add(record *metrics.AppMetric) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	go func() {
		b.Channel <- 1
	}()

	b.metrics = append(b.metrics, *record)
}

// All returns all metrics
func (b *AppMetricBucket) All() *[]metrics.AppMetric {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return &b.metrics
}

// Count the number of metrics in the bucket
func (b *AppMetricBucket) Count() int {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return len(b.metrics)
}

// Extract a limited amount of records from the bucket
func (b *AppMetricBucket) Extract(limit int) *AppMetricBucket {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	count := len(b.metrics)
	if limit > count {
		limit = count
	}

	newBucketItems := make([]metrics.AppMetric, 0)
	newBucketItems = append(newBucketItems, b.metrics[:limit]...)

	b.metrics = append(make([]metrics.AppMetric, 0), b.metrics[count:]...)

	return NewBucketFromItems(newBucketItems)
}

// String concatenates all metrics into a single string separated by newlines
func (b *AppMetricBucket) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var strs []string

	for i := 0; i < len(b.metrics); i++ {
		strs = append(strs, b.metrics[i].String())
	}

	return strings.Join(strs, "\n")
}
