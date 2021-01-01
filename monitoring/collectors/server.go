package collectors

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"

	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/metrics"
)

// ServerMetricCollector defines a server resource collector
type ServerMetricCollector struct {
	inDocker             bool
	dockerClient         *DockerClient
	bucket               *buckets.ServerMetricBucket
	serverMetricInterval time.Duration
	hostname             string
	stop                 chan int
}

// NewServerMetricCollector creates a new instance of `ServerMetricCollector`
func NewServerMetricCollector(
	bucket *buckets.ServerMetricBucket,
	serverMetricInterval time.Duration,
	hostname string,
	inDocker bool) *ServerMetricCollector {
	dockerClient, err := NewDockerClient()
	if err != nil {
		log.Trace().Err(err)
	}

	return &ServerMetricCollector{
		inDocker,
		dockerClient,
		bucket,
		serverMetricInterval,
		hostname,
		make(chan int, 0),
	}
}

// Start server metric collection
func (smc *ServerMetricCollector) Start() {
	ticker := time.NewTicker(smc.serverMetricInterval)
	defer ticker.Stop()

	// lets measure at start
	metric, err := smc.fetchServerMetrics()
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
			metric, err := smc.fetchServerMetrics()
			if err != nil {
				log.Error().Msgf("Failed to collect server metrics: %v", err)

				continue
			}
			smc.bucket.Add(metric)
		}
	}
}

// Stop server metric collection
func (smc *ServerMetricCollector) Stop() {
	smc.stop <- 1
}

func (smc *ServerMetricCollector) fetchServerMetrics() (*metrics.ServerMetric, error) {
	metric := &metrics.ServerMetric{
		RebootRequired: false,
		Hostname:       smc.hostname,
	}

	cp, err := smc.cpu()
	if err == nil {
		metric.CPUUsedPercentage = cp
	}

	cc, err := smc.processorCoreCount()
	if err == nil {
		metric.CPUCoreCount = cc
	}

	m, err := smc.memory()
	if err == nil {
		metric.MemoryTotal = m.Total
		metric.MemoryUserPercentage = m.UsedPercent
	}

	l, err := smc.load()
	if err == nil {
		metric.Load = *l
	}

	d, err := smc.disk()
	if err == nil {
		metric.DiskTotal = d.Total
		metric.DiskUsedPercentage = d.UsedPercent
	}

	if !smc.inDocker {
		s, err := smc.services()
		if err == nil {
			metric.Services = s
		} else {
			log.Trace().Err(err).Msg("Failed to fetch services")
		}

		if smc.dockerClient != nil {
			c, err := smc.dockerClient.FetchContainers()
			if err != nil {
				log.Trace().Err(err).Msg("Failed to fetch containers")
			} else {
				metric.Containers = c
			}
		}
	}

	osInfo, err := smc.os()
	if err == nil {
		metric.OS = osInfo
	} else {
		log.Trace().Err(err)
	}

	phpVersion, err := smc.phpVersion()
	if err == nil {
		metric.PHPVersion = phpVersion
	} else {
		log.Trace().Err(err).Msg("Failed to fetch PHP version")
	}

	uptime, err := host.BootTime()
	if err == nil {
		metric.BootTime = uptime
	}

	if _, err := os.Stat("/var/run/reboot-required"); err == nil {
		metric.RebootRequired = true
	}

	err = nil

	metric.CreatedAt = time.Now()

	return metric, err
}

func (smc *ServerMetricCollector) processorCoreCount() (int, error) {
	count, err := cpu.Counts(false)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (smc *ServerMetricCollector) cpu() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}

	return percentages[0], nil
}

func (smc *ServerMetricCollector) memory() (*mem.VirtualMemoryStat, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (smc *ServerMetricCollector) disk() (*disk.UsageStat, error) {
	m, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (smc *ServerMetricCollector) load() (*metrics.ServerLoad, error) {
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

func (smc *ServerMetricCollector) os() (*metrics.OS, error) {
	platform, _, version, err := host.PlatformInformation()
	if err != nil {
		return nil, err
	}

	osInfo := &metrics.OS{
		Name:    platform,
		Version: version,
	}

	return osInfo, nil
}

func (smc *ServerMetricCollector) phpVersion() (string, error) {
	phpPath, err := exec.LookPath("php")
	if err != nil {
		return "", err
	}

	cmd := exec.Command(phpPath, "-v")
	output, err := cmd.CombinedOutput()

	r, err := regexp.Compile(`^PHP\s([^\s]*)`)
	if err != nil {
		return "", err
	}
	matched := r.FindStringSubmatch(string(output))
	if len(matched) == 0 {
		return "", nil
	}

	return matched[1], nil
}

func (smc *ServerMetricCollector) services() (services []metrics.Service, err error) {
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
