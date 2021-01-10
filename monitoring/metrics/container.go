package metrics

// Volume defines a docker volume
type Volume struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Driver      string `json:"driver"`
	Mode        string `json:"mode"`
	Size        int64  `json:"size"`
	RW          bool   `json:"rw"`
}

// Port defines a container port
type Port struct {
	IP          string `json:"ip_address"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port"`
	Type        string `json:"type"`
}

// DockerCompose defines docker-compose data
type DockerCompose struct {
	Project         string `json:"project"`
	Version         string `json:"version"`
	Directory       string `json:"directory"`
	ContainerNumber string `json:"container_number"`
}

// Container defines container data
type Container struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"state"`
	State  string `json:"status"`

	DockerCompose DockerCompose `json:"docker_compose"`

	Image         string `json:"image"`
	SizeContainer int64  `json:"size_container"`
	SizeAdded     int64  `json:"size_added"`

	Ports   []Port   `json:"ports"`
	Volumes []Volume `json:"volumes"`

	Command           string  `json:"command"`
	CreatedAt         int64   `json:"created_at"`
	CPUUsedPercentage float64 `json:"cpu_used_percentage"`

	MemoryTotal          float64 `json:"memory_total"`
	MemoryCurrent        float64 `json:"memory_current"`
	MemoryUsedPercentage float64 `json:"memory_used_percentage"`

	IPAddress string `json:"ip_address"`

	NetworkName     string  `json:"network_name"`
	NetworkInbound  float64 `json:"network_inbound"`
	NetworkOutbound float64 `json:"network_outbound"`

	PIDs uint64 `json:"pid_count"`
}
