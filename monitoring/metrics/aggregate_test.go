package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAggregateServerMetrics(t *testing.T) {
	lastMetric := &ServerMetric{
		CPUUsedPercentage: 15,
		CPUCoreCount:      3,
		Load: ServerLoad{
			Load1:  8,
			Load5:  8,
			Load15: 8,
		},
		MemoryTotal:          5,
		MemoryUserPercentage: 1,
		DiskTotal:            5,
		DiskUsedPercentage:   1,
		CreatedAt:            time.Now(),
	}
	list := []*ServerMetric{
		{
			CPUUsedPercentage: 5,
			CPUCoreCount:      1,
			Load: ServerLoad{
				Load1:  2,
				Load5:  2,
				Load15: 2,
			},
			MemoryTotal:          0,
			MemoryUserPercentage: 3,
			DiskTotal:            0,
			DiskUsedPercentage:   3,
			CreatedAt:            time.Time{},
		},
		lastMetric,
	}

	metric := AggregateServerMetrics(list)

	assert.Equal(t, 3, lastMetric.CPUCoreCount)
	assert.Equal(t, uint64(5), lastMetric.DiskTotal)
	assert.Equal(t, uint64(5), lastMetric.MemoryTotal)

	assert.Equal(t, 10.0, metric.CPUUsedPercentage)

	assert.Equal(t, 5.0, metric.Load.Load1)
	assert.Equal(t, 5.0, metric.Load.Load5)
	assert.Equal(t, 5.0, metric.Load.Load15)

	assert.Equal(t, float64(2), metric.DiskUsedPercentage)
	assert.Equal(t, float64(2), metric.MemoryUserPercentage)
}
