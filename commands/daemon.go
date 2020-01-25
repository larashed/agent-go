package commands

import (
	"log"
	"os"
	"time"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/monitoring"
	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/collectors"
	"github.com/larashed/agent-go/monitoring/sender"
	socketserver "github.com/larashed/agent-go/server"
)

type Daemon struct {
	api          api.Api
	socketServer socketserver.DomainSocketServer
}

func NewDaemonCommand(api api.Api, socketServer socketserver.DomainSocketServer) *Daemon {
	return &Daemon{
		api:          api,
		socketServer: socketServer,
	}
}

func (d *Daemon) Run() error {
	log.Print("Starting daemon..")
	log.Printf("PID: %d", os.Getpid())

	config := sender.NewConfig(200, 5, time.Second*15, time.Minute)

	appMetricBucket := buckets.NewBucket()
	serverMetricBucket := buckets.NewBucket()

	appMetricCollector := collectors.NewAppMetricCollector(d.socketServer, appMetricBucket)
	serverMetricCollector := collectors.NewServerMetricCollector(serverMetricBucket)

	collector := monitoring.NewCollector(serverMetricCollector, appMetricCollector)
	collector.Start() // non-blocking, starts goroutines

	s := sender.NewSender(d.api, appMetricBucket, serverMetricBucket, config)
	s.Start() // non-blocking, stars goroutines

	// should block main() from exiting

	return nil
}
