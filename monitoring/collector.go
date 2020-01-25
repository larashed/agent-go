package monitoring

import (
	"context"
	"errors"
	"log"

	"github.com/larashed/agent-go/monitoring/collectors"
)

// collects and sends metrics to the backend
type Collector struct {
	serverMetricCollector *collectors.ServerMetricCollector
	appMetricCollector    *collectors.AppMetricCollector
	ctx                   *context.Context
}

func NewCollector(
	serverMetricCollector *collectors.ServerMetricCollector,
	appMetricCollector *collectors.AppMetricCollector,
) *Collector {
	return &Collector{
		serverMetricCollector: serverMetricCollector,
		appMetricCollector:    appMetricCollector,
	}
}

func (c *Collector) Start() {
	go c.collectServerMetrics()
	go c.collectApplicationMetrics()
}

func (c *Collector) collectServerMetrics() {
	log.Println("Starting server resource collection..")

	c.serverMetricCollector.Collect()
}

func (c *Collector) collectApplicationMetrics() {
	log.Println("Starting application metric collection..")
	err := c.appMetricCollector.Collect()
	if err != nil {
		log.Fatal(errors.New("failed to start socket server"))
	}
}
