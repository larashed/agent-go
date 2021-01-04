package collectors

import (
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"

	"github.com/larashed/agent-go/monitoring/metrics"
)

// Services returns systemd services
func Services() ([]metrics.Service, error) {
	con, err := dbus.New()

	if err != nil {
		return nil, err
	}

	defer con.Close()

	services, err := con.ListUnitsByPatterns([]string{"running"}, []string{"*.service"})
	if err != nil {
		return nil, err
	}

	srvs := make([]metrics.Service, 0)
	for i := 0; i < len(services); i++ {
		srvs = append(srvs, metrics.Service{
			Name:        strings.Replace(services[i].Name, ".service", "", 1),
			Description: services[i].Description,
			LoadState:   services[i].LoadState,
			ActiveState: services[i].ActiveState,
			SubState:    services[i].SubState,
		})
	}

	return srvs, nil
}
