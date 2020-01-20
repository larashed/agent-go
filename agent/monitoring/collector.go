package monitoring

import (
	"encoding/json"
	"log"
	"time"

	"github.com/larashed/agent-go/agent/monitoring/buckets"
	"github.com/larashed/agent-go/agent/monitoring/collectors"
	"github.com/larashed/agent-go/agent/server"
	"github.com/larashed/agent-go/api"
)

type Collector struct {
	api                        *api.Api
	serverResourceCollector    *collectors.ServerResourceCollector
	applicationMetricCollector *collectors.ApplicationMetricCollector
	metricBucket               *buckets.Bucket
}

func NewCollector(api *api.Api, socketServer *server.Server) *Collector {
	return &Collector{
		api:                        api,
		serverResourceCollector:    collectors.NewServerResourceCollector(),
		applicationMetricCollector: collectors.NewApplicationMetricCollector(socketServer),
		metricBucket:               buckets.NewBucket(),
	}
}

func (c *Collector) Start() error {
	log.Println("Starting server resource collection..")
	c.collectServerResources()
	log.Println("Starting application metric collection..")
	err := c.collectApplicationMetrics()
	if err != nil {
		return err
	}

	return nil
}

func (c *Collector) collectServerResources() {
	//quit := make(chan struct{})

	// refactor this to be stoppable
	ticker := time.NewTicker(20 * time.Second)
	collectMetrics := func() {
		metrics, err := c.serverResourceCollector.Collect()
		if err == nil {
			j, err := json.MarshalIndent(metrics, "", "  ")

			if err == nil {
				log.Println(string(j))
			}
		} else {
			log.Fatal("error", err)
		}
	}

	go func() {
		collectMetrics()
		for {
			select {
			case <-ticker.C:
				collectMetrics()
			}
		}
	}()
}

func (c *Collector) collectApplicationMetrics() error {
	err := c.applicationMetricCollector.Collect(func(record string) {
		c.metricBucket.Add(record)
	})

	return err
}
