package buckets

import (
	"sort"
	"sync"

	"github.com/larashed/agent-go/monitoring/metrics"
)

type ServerMetrics map[int][]*metrics.ServerMetric

type ServerMetricBucket struct {
	metrics ServerMetrics
	mutex   sync.RWMutex
	Channel chan int
}

func NewServerMetricBucket() *ServerMetricBucket {
	return &ServerMetricBucket{
		metrics: make(ServerMetrics, 0),
		mutex:   sync.RWMutex{},
		Channel: make(chan int),
	}
}

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

func (s *ServerMetricBucket) Metrics(minute int) []*metrics.ServerMetric {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.metrics[minute]
}

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

func (s *ServerMetricBucket) Remove(minute int) {
	delete(s.metrics, minute)
}

func (s *ServerMetricBucket) Count() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.metrics)
}
