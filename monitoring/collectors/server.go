package collectors

import (
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"

	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/metrics"
)

type ServerMetricCollector struct {
	bucket            *buckets.ServerMetricBucket
	intervalInSeconds int
	stop              chan int
}

func NewServerMetricCollector(bucket *buckets.ServerMetricBucket, intervalInSeconds int) *ServerMetricCollector {
	return &ServerMetricCollector{
		bucket,
		intervalInSeconds,
		make(chan int, 0),
	}
}

func (c *ServerMetricCollector) Start() {
	ticker := time.NewTicker(time.Duration(c.intervalInSeconds) * time.Second)
	defer ticker.Stop()

	// lets measure at start
	metric, _ := c.buildServerMetrics()
	c.bucket.Add(metric)

	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			metric, err := c.buildServerMetrics()
			if err != nil {
				continue
			}
			c.bucket.Add(metric)
		}
	}
}

func (c *ServerMetricCollector) Stop() {
	c.stop <- 1
}

func (c *ServerMetricCollector) buildServerMetrics() (*metrics.ServerMetric, error) {
	metric := &metrics.ServerMetric{}

	cp, err := c.CPU()
	if err == nil {
		metric.CPUUsedPercentage = cp
	}

	cc, err := c.CPUCoreCount()
	if err == nil {
		metric.CPUCoreCount = cc
	}

	m, err := c.Memory()
	if err == nil {
		metric.MemoryTotal = m.Total
		metric.MemoryUserPercentage = m.UsedPercent
	}

	l, err := c.Load()
	if err == nil {
		metric.Load = *l
	}

	d, err := c.Disk()
	if err == nil {
		metric.DiskTotal = d.Total
		metric.DiskUsedPercentage = d.UsedPercent
	}

	metric.CreatedAt = time.Now()

	hostname, err := os.Hostname()
	if err == nil {
		metric.Hostname = hostname
	}

	return metric, err
}

func (c *ServerMetricCollector) CPUCoreCount() (int, error) {
	count, err := cpu.Counts(false)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (c *ServerMetricCollector) CPU() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}

	return percentages[0], nil
}

func (c *ServerMetricCollector) Memory() (*mem.VirtualMemoryStat, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (c *ServerMetricCollector) Disk() (*disk.UsageStat, error) {
	m, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (c *ServerMetricCollector) Load() (*metrics.ServerLoad, error) {
	avg, err := load.Avg()
	if err != nil {
		return nil, err
	}

	return &metrics.ServerLoad{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
	}, nil
}
