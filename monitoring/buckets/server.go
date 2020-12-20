package buckets

import (
	"sync"

	"github.com/larashed/agent-go/monitoring/metrics"
)

// ServerMetricBucket holds server metrics
type ServerMetricBucket struct {
	mutex   sync.RWMutex
	Channel chan metrics.ServerMetric
}

// NewServerMetricBucket returns a new instance of `ServerMetricBucket`
func NewServerMetricBucket() *ServerMetricBucket {
	return &ServerMetricBucket{
		mutex:   sync.RWMutex{},
		Channel: make(chan metrics.ServerMetric),
	}
}

// Add a server metric to the bucket
func (s *ServerMetricBucket) Add(record *metrics.ServerMetric) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Channel <- *record
}
