package collectors

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"

	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/metrics"
)

type ServerMetricCollector struct {
	inDocker          bool
	bucket            *buckets.ServerMetricBucket
	intervalInSeconds int
	stop              chan int
}

func NewServerMetricCollector(
	bucket *buckets.ServerMetricBucket,
	intervalInSeconds int,
	inDocker bool) *ServerMetricCollector {
	return &ServerMetricCollector{
		inDocker,
		bucket,
		intervalInSeconds,
		make(chan int, 0),
	}
}

func (smc *ServerMetricCollector) Start() {
	ticker := time.NewTicker(time.Duration(smc.intervalInSeconds) * time.Second)
	defer ticker.Stop()

	// lets measure at start
	metric, err := smc.buildServerMetrics()
	if err != nil {
		log.Error().Msgf("Failed to collect server metrics: %v", err)
	} else {
		smc.bucket.Add(metric)
	}

	for {
		select {
		case <-smc.stop:
			return
		case <-ticker.C:
			metric, err := smc.buildServerMetrics()
			if err != nil {
				log.Error().Msgf("Failed to collect server metrics: %v", err)

				continue
			}
			smc.bucket.Add(metric)
		}
	}
}

func (smc *ServerMetricCollector) Stop() {
	smc.stop <- 1
}

func (smc *ServerMetricCollector) buildServerMetrics() (*metrics.ServerMetric, error) {
	metric := &metrics.ServerMetric{}

	cp, err := smc.CPU()
	if err == nil {
		metric.CPUUsedPercentage = cp
	}

	cc, err := smc.CPUCoreCount()
	if err == nil {
		metric.CPUCoreCount = cc
	}

	m, err := smc.Memory()
	if err == nil {
		metric.MemoryTotal = m.Total
		metric.MemoryUserPercentage = m.UsedPercent
	}

	l, err := smc.Load()
	if err == nil {
		metric.Load = *l
	}

	d, err := smc.Disk()
	if err == nil {
		metric.DiskTotal = d.Total
		metric.DiskUsedPercentage = d.UsedPercent
	}

	if !smc.inDocker {
		s, err := smc.Services()
		if err == nil {
			metric.Services = s
		}

		c, err := smc.DockerContainers()
		if err != nil {
			log.Trace().Err(err)
		}

		if err == nil {
			metric.Containers = c
		}
	}

	metric.CreatedAt = time.Now()

	return metric, err
}

func (smc *ServerMetricCollector) CPUCoreCount() (int, error) {
	count, err := cpu.Counts(false)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (smc *ServerMetricCollector) CPU() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}

	return percentages[0], nil
}

func (smc *ServerMetricCollector) Memory() (*mem.VirtualMemoryStat, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (smc *ServerMetricCollector) Disk() (*disk.UsageStat, error) {
	m, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (smc *ServerMetricCollector) Load() (*metrics.ServerLoad, error) {
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

func (smc *ServerMetricCollector) DockerContainers() (containers []metrics.Container, err error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return containers, err
	}

	cmd := exec.Command(dockerPath, "stats", "--no-stream", "--format", "{{ json . }}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return containers, err
	}

	list := strings.Split(
		strings.TrimSpace(string(output)),
		"\n",
	)
	for i := 0; i < len(list); i++ {
		dto := &metrics.DockerStatsDto{}

		err := json.Unmarshal([]byte(list[i]), dto)
		if err == nil {
			containers = append(containers, *dto.ToContainer())
		}
	}

	return containers, nil
}

func (smc *ServerMetricCollector) Services() (services []metrics.Service, err error) {
	cmd := exec.Command("service", "--status-all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseServiceList(string(output)), nil
}

func parseServiceList(output string) (services []metrics.Service) {
	list := strings.Split(output, "\n")
	for i := 0; i < len(list); i++ {
		r, err := regexp.Compile(`\[\s(.)\s\](.*)`)
		if err != nil {
			continue
		}
		matched := r.FindStringSubmatch(list[i])
		if len(matched) == 0 {
			continue
		}
		service := metrics.Service{}
		parsedStatus := strings.TrimSpace(matched[1])
		switch parsedStatus {
		case "+":
			service.Status = metrics.ServiceStatusRunning
			break
		case "-":
			service.Status = metrics.ServiceStatusStopped
			break
		case "?":
			service.Status = metrics.ServiceStatusUnknown
			break
		default:
			continue
		}

		service.Name = strings.TrimSpace(matched[2])

		services = append(services, service)
	}
	return services
}
