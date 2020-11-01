package metrics

import (
	"strconv"
	"strings"

	"code.cloudfoundry.org/bytefmt"
)

// DockerStatsDto holds `docker stats` command JSON DTO
type DockerStatsDto struct {
	BlockIO   string `json:"BlockIO"`
	CPUPerc   string `json:"CPUPerc"`
	Container string `json:"Container"`
	ID        string `json:"ID"`
	MemPerc   string `json:"MemPerc"`
	MemUsage  string `json:"MemUsage"`
	Name      string `json:"Name"`
	NetIO     string `json:"NetIO"`
	PIDs      string `json:"PIDs"`
}

// Container defines container data
type Container struct {
	Name                 string  `json:"name"`
	ID                   string  `json:"id"`
	Type                 string  `json:"type"`
	CPUUsedPercentage    float64 `json:"cpu_used_percentage"`
	MemoryTotal          uint64  `json:"memory_total"`
	MemoryUserPercentage float64 `json:"memory_used_percentage"`
}

// ToContainer map `docker stats` DTO to internal `Container` type
func (dto *DockerStatsDto) ToContainer() *Container {
	c := &Container{
		ID:                   dto.ID,
		Name:                 dto.Name,
		CPUUsedPercentage:    0,
		MemoryTotal:          0,
		MemoryUserPercentage: 0,
		Type:                 "docker",
	}

	if s, err := strconv.ParseFloat(strings.Replace(dto.CPUPerc, "%", "", 1), 64); err == nil {
		c.CPUUsedPercentage = s
	}

	if s, err := strconv.ParseFloat(strings.Replace(dto.MemPerc, "%", "", 1), 64); err == nil {
		c.MemoryUserPercentage = s
	}

	memUsage := strings.Split(dto.MemUsage, "/")
	if len(memUsage) == 2 {
		total, err := bytefmt.ToBytes(memUsage[1])
		if err == nil {
			c.MemoryTotal = total
		}
	}

	return c
}
