package buckets

import (
	"strings"
	"sync"

	"github.com/larashed/agent-go/monitoring/metrics"
)

type Metrics []metrics.AppMetric

type AppMetricBucket struct {
	metrics Metrics
	mutex   sync.RWMutex
	Channel chan int
}

func NewAppMetricBucket() *AppMetricBucket {
	return &AppMetricBucket{
		metrics: make(Metrics, 0),
		mutex:   sync.RWMutex{},
		Channel: make(chan int),
	}
}

func NewBucketFromItems(items Metrics) *AppMetricBucket {
	return &AppMetricBucket{
		metrics: items,
		mutex:   sync.RWMutex{},
	}
}

func (b *AppMetricBucket) Merge(bucket *AppMetricBucket) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.metrics = append(b.metrics, *bucket.All()...)
}

func (b *AppMetricBucket) Add(record *metrics.AppMetric) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	go func() {
		b.Channel <- 1
	}()

	b.metrics = append(b.metrics, *record)
}

func (b *AppMetricBucket) All() *Metrics {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return &b.metrics
}

func (b *AppMetricBucket) Count() int {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return len(b.metrics)
}

func (b *AppMetricBucket) Extract(limit int) *AppMetricBucket {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	count := len(b.metrics)
	if limit > count {
		limit = count
	}

	newBucketItems := make(Metrics, 0)
	newBucketItems = append(newBucketItems, b.metrics[:limit]...)

	b.metrics = append(make(Metrics, 0), b.metrics[count:]...)

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
