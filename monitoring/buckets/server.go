package buckets

import (
	"sort"
	"sync"

	"github.com/larashed/agent-go/monitoring/metrics"
)

// ServerMetrics holds server metrics grouped by minute
type ServerMetrics map[int][]*metrics.ServerMetric

// ServerMetricBucket holds server metrics
type ServerMetricBucket struct {
	metrics ServerMetrics
	mutex   sync.RWMutex
	Channel chan int
}

// NewServerMetricBucket returns a new instance of `ServerMetricBucket`
func NewServerMetricBucket() *ServerMetricBucket {
	return &ServerMetricBucket{
		metrics: make(ServerMetrics, 0),
		mutex:   sync.RWMutex{},
		Channel: make(chan int),
	}
}

// Add a server metric to the bucket
func (s *ServerMetricBucket) Add(record *metrics.ServerMetric) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	go func() {
		s.Channel <- 1
	}()

	// store records in minute groups
	t := record.CreatedAt.Minute()

	s.metrics[t] = append(s.metrics[t], record)
}

// Metrics returns server metrics for a selected minute
func (s *ServerMetricBucket) Metrics(minute int) []*metrics.ServerMetric {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.metrics[minute]
}

// Minutes returns all minutes with collected metrics
func (s *ServerMetricBucket) Minutes() []int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	keys := make([]int, 0, len(s.metrics))
	for key := range s.metrics {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	return keys
}

// Remove a minute with its metrics
func (s *ServerMetricBucket) Remove(minute int) {
	delete(s.metrics, minute)
}
