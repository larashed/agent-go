package metrics

import (
	"encoding/json"
	"time"
)

// Service represents a systemd service
type Service struct {
	Name        string `json:"name"`         // The primary unit name as string
	Description string `json:"description"`  // The human readable description string
	LoadState   string `json:"load_state"`   // The load state (i.e. whether the unit file has been loaded successfully)
	ActiveState string `json:"active_state"` // The active state (i.e. whether the unit is currently started or not)
	SubState    string `json:"sub_state"`    // The sub state (a more fine-grained version of the active state that is specific to the
	// unit type, which the active state is not)
}

// ServerLoad represents server load metric
type ServerLoad struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// OS represents the underlying OS information
type OS struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerMetric represents a server metric
type ServerMetric struct {
	Hostname             string      `json:"hostname"`
	CPUUsedPercentage    float64     `json:"cpu_used_percentage"`
	CPUCoreCount         int         `json:"cpu_core_count"`
	Load                 ServerLoad  `json:"load"`
	MemoryTotal          uint64      `json:"memory_total"`
	MemoryUserPercentage float64     `json:"memory_used_percentage"`
	DiskTotal            uint64      `json:"disk_total"`
	DiskUsedPercentage   float64     `json:"disk_used_percentage"`
	CreatedAt            time.Time   `json:"-"`
	CreatedAtFormatted   string      `json:"created_at"`
	OS                   *OS         `json:"os"`
	BootTime             uint64      `json:"boot_time"`
	RebootRequired       bool        `json:"reboot_required"`
	Services             []Service   `json:"services"`
	Containers           []Container `json:"containers"`
	PHPVersion           string      `json:"php_version"`
}

// String returns `ServerMetric` in a string format
func (sm *ServerMetric) String() string {
	sm.CreatedAtFormatted = sm.CreatedAt.Format(time.RFC3339)
	str, _ := json.Marshal(sm)

	return string(str)
}
