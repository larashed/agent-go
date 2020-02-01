package metrics

import (
	"encoding/json"
	"time"
)

const ServiceStatusStopped = 0
const ServiceStatusRunning = 1
const ServiceStatusUnknown = 2

type Service struct {
	Status int    `json:"status"`
	Name   string `json:"name"`
}

type ServerLoad struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type ServerMetric struct {
	Hostname             string     `json:"hostname"`
	CPUUsedPercentage    float64    `json:"cpu_used_percentage"`
	CPUCoreCount         int        `json:"cpu_core_count"`
	Load                 ServerLoad `json:"load"`
	MemoryTotal          uint64     `json:"memory_total"`
	MemoryUserPercentage float64    `json:"memory_used_percentage"`
	DiskTotal            uint64     `json:"disk_total"`
	DiskUsedPercentage   float64    `json:"disk_used_percentage"`
	CreatedAt            time.Time  `json:"-"`
	Services             []Service  `json:"services"`
	CreatedAtFormatted   string     `json:"created_at"`
}

func (sm *ServerMetric) String() string {
	sm.CreatedAtFormatted = sm.CreatedAt.Format(time.RFC3339)
	str, _ := json.Marshal(sm)

	return string(str)
}

func (sm *ServerMetric) Value() *ServerMetric {
	return sm
}
