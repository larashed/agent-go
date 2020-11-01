package metrics

import (
	"encoding/json"
	"time"
)

// ServiceStatusStopped used to indicate a stopped service
const ServiceStatusStopped = 0

// ServiceStatusRunning used to indicate a running service
const ServiceStatusRunning = 1

// ServiceStatusUnknown used to indicate an unknown state service
const ServiceStatusUnknown = 2

// Service represents an OS service
type Service struct {
	Status int    `json:"status"`
	Name   string `json:"name"`
}

// ServerLoad represents server load metric
type ServerLoad struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// ServerMetric represents a server metric
type ServerMetric struct {
	CPUUsedPercentage    float64     `json:"cpu_used_percentage"`
	CPUCoreCount         int         `json:"cpu_core_count"`
	Load                 ServerLoad  `json:"load"`
	MemoryTotal          uint64      `json:"memory_total"`
	MemoryUserPercentage float64     `json:"memory_used_percentage"`
	DiskTotal            uint64      `json:"disk_total"`
	DiskUsedPercentage   float64     `json:"disk_used_percentage"`
	CreatedAt            time.Time   `json:"-"`
	CreatedAtFormatted   string      `json:"created_at"`
	Services             []Service   `json:"services"`
	Containers           []Container `json:"containers"`
}

// String returns `ServerMetric` in a string format
func (sm *ServerMetric) String() string {
	sm.CreatedAtFormatted = sm.CreatedAt.Format(time.RFC3339)
	str, _ := json.Marshal(sm)

	return string(str)
}
