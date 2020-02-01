package collectors

import (
	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/metrics"
	"github.com/larashed/agent-go/server"
)

type AppMetricCollector struct {
	socketServer server.DomainSocketServer
	bucket       *buckets.AppMetricBucket
}

func NewAppMetricCollector(socketServer server.DomainSocketServer, bucket *buckets.AppMetricBucket) *AppMetricCollector {
	return &AppMetricCollector{
		socketServer: socketServer,
		bucket:       bucket,
	}
}

func (amc *AppMetricCollector) Start() error {
	return amc.socketServer.Start(func(record string) {
		amc.bucket.Add(metrics.NewAppMetric(record))
	})
}

func (amc *AppMetricCollector) Stop() error {
	return amc.socketServer.Stop()
}
