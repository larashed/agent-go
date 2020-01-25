package collectors

import (
	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/server"
)

type AppMetric struct {
	record string
}

func NewAppMetric(record string) *AppMetric {
	return &AppMetric{record}
}

func (am *AppMetric) String() string {
	return am.record
}

func (am *AppMetric) Value() interface{} {
	return am.record
}

type AppMetricCollector struct {
	socketServer server.DomainSocketServer
	bucket       *buckets.Bucket
}

func NewAppMetricCollector(socketServer server.DomainSocketServer, bucket *buckets.Bucket) *AppMetricCollector {
	return &AppMetricCollector{
		socketServer: socketServer,
		bucket:       bucket,
	}
}

func (amc *AppMetricCollector) Collect() error {
	return amc.socketServer.Start(func(record string) {
		amc.bucket.Add(NewAppMetric(record))
	})
}
