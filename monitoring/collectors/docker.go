package collectors

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	docker "github.com/larashed/agent-go/docker"
	"github.com/larashed/agent-go/monitoring/metrics"
)

// DockerClient holds the docker API client
type DockerClient struct {
	client *client.Client
}

// NewDockerClient creates a docker API client instance
func NewDockerClient() (*DockerClient, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Docker API client")
	}

	return &DockerClient{
		client: apiClient,
	}, nil
}

// FetchContainers fetches docker containers with metrics and volumes
func (dc *DockerClient) FetchContainers() (collectedContainers []metrics.Container, err error) {
	containersWithStats, err := docker.GetStats(dc.client)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch item stats")
	}

	containerList, err := dc.client.ContainerList(context.Background(), types.ContainerListOptions{Size: true})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch item list")
	}

	for _, item := range containerList {
		cont := metrics.Container{}
		cont.ID = item.ID
		cont.Name = strings.Join(item.Names, ";")
		cont.Type = "docker"
		cont.Command = item.Command
		cont.Image = item.Image
		cont.CreatedAt = item.Created
		cont.State = item.State
		cont.Status = item.Status

		cont.SizeContainer = item.SizeRootFs
		cont.SizeAdded = item.SizeRw

		cont.NetworkName = item.HostConfig.NetworkMode
		if item.NetworkSettings.Networks[cont.NetworkName] != nil {
			cont.IPAddress = item.NetworkSettings.Networks[cont.NetworkName].IPAddress
		}

		for _, mount := range item.Mounts {
			cont.Volumes = append(cont.Volumes, metrics.Volume{
				Type:        string(mount.Type),
				Name:        mount.Name,
				Source:      mount.Source,
				Destination: mount.Destination,
				Driver:      mount.Driver,
				Mode:        mount.Mode,
				RW:          mount.RW,
			})
		}

		for _, port := range item.Ports {
			cont.Ports = append(cont.Ports, metrics.Port{
				IP:          port.IP,
				PrivatePort: port.PrivatePort,
				PublicPort:  port.PublicPort,
				Type:        port.Type,
			})
		}

		for _, cws := range containersWithStats {
			if cont.ID != cws.ID {
				continue
			}

			cont.PIDs = cws.PidsCurrent
			cont.CPUUsedPercentage = cws.CPUPercentage

			cont.MemoryCurrent = cws.Memory
			cont.MemoryUsedPercentage = cws.MemoryPercentage
			cont.MemoryTotal = cws.MemoryLimit

			cont.NetworkInbound = cws.NetworkRx
			cont.NetworkOutbound = cws.NetworkTx
		}

		collectedContainers = append(collectedContainers, cont)
	}

	return collectedContainers, nil
}
