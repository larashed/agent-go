package collectors

import (
	"github.com/larashed/agent-go/agent/server"
)

type ApplicationMetricCollector struct {
	socketServer *server.Server
}

func NewApplicationMetricCollector(socketServer *server.Server) *ApplicationMetricCollector {
	return &ApplicationMetricCollector{
		socketServer: socketServer,
	}
}

func (amc *ApplicationMetricCollector) Collect(callback func(record string)) error {
	return amc.socketServer.Start(callback)
}
