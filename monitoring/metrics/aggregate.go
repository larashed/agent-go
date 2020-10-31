package metrics

import (
	"math"
)

func AggregateServerMetrics(metrics []*ServerMetric) *ServerMetric {
	var mem = 0.0
	var disk = 0.0
	var cpu = 0.0
	var load1 = 0.0
	var load5 = 0.0
	var load15 = 0.0

	length := len(metrics)
	lastMetric := metrics[length-1]

	for i := 0; i < length; i++ {
		disk += metrics[i].DiskUsedPercentage
		mem += metrics[i].MemoryUserPercentage
		cpu += metrics[i].CPUUsedPercentage
		load1 += metrics[i].Load.Load1
		load5 += metrics[i].Load.Load5
		load15 += metrics[i].Load.Load15
	}

	metric := &ServerMetric{}
	metric.CPUUsedPercentage = round(cpu / float64(length))
	metric.CPUCoreCount = lastMetric.CPUCoreCount
	metric.CreatedAt = lastMetric.CreatedAt
	metric.Services = lastMetric.Services

	metric.MemoryTotal = lastMetric.MemoryTotal
	metric.MemoryUserPercentage = round(mem / float64(length))

	metric.DiskTotal = lastMetric.DiskTotal
	metric.DiskUsedPercentage = round(disk / float64(length))

	metric.Load.Load1 = round(load1 / float64(length))
	metric.Load.Load5 = round(load5 / float64(length))
	metric.Load.Load15 = round(load15 / float64(length))

	return metric
}

func round(a float64) float64 {
	return math.Round(a*100) / 100
}
