package collectors

import (
	"encoding/json"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"

	"github.com/larashed/agent-go/monitoring/buckets"
)

type ServerLoad struct {
	Load1  float64 `json:"1min"`
	Load5  float64 `json:"5min"`
	Load15 float64 `json:"15min"`
}

type ServerMetric struct {
	CPU          float64    `json:"cpu"`
	CPUCoreCount int        `json:"cpu_core_count"`
	Load         ServerLoad `json:"load"`
	MemoryTotal  uint64     `json:"memory_total"`
	MemoryFree   float64    `json:"memory_free"`
	DiskTotal    uint64     `json:"disk_total"`
	DiskFree     uint64     `json:"disk_free"`
}

func (sm *ServerMetric) String() string {
	str, _ := json.Marshal(sm)

	return string(str)
}

func (sm *ServerMetric) Value() interface{} {
	return sm
}

type ServerMetricCollector struct {
	bucket *buckets.Bucket
}

func NewServerMetricCollector(bucket *buckets.Bucket) *ServerMetricCollector {
	return &ServerMetricCollector{bucket}
}

func (c *ServerMetricCollector) Collect() {
	//metrics, err = c.getMetrics()
	//if err != nil {
	//	return
	//}
	//
	//return metrics, nil
}

func (c *ServerMetricCollector) getMetrics() (ServerMetric, error) {
	metrics := ServerMetric{}

	cp, err := c.CPU()
	if err == nil {
		metrics.CPU = cp
	}

	cc, err := c.CPUCoreCount()
	if err == nil {
		metrics.CPUCoreCount = cc
	}

	m, err := c.Memory()
	if err == nil {
		metrics.MemoryTotal = m.Total
		metrics.MemoryFree = m.UsedPercent
	}

	l, err := c.Load()
	if err == nil {
		metrics.Load = *l
	}

	d, err := c.Disk()
	if err == nil {
		metrics.DiskTotal = d.Total
		metrics.DiskFree = d.Free
	}

	return metrics, err
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

func (c *ServerMetricCollector) Load() (*ServerLoad, error) {
	avg, err := load.Avg()
	if err != nil {
		return nil, err
	}

	return &ServerLoad{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
	}, nil
}
