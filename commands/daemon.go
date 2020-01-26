package commands

import (
	"log"
	"os"
	"time"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/collectors"
	"github.com/larashed/agent-go/monitoring/sender"
	socketserver "github.com/larashed/agent-go/server"
	"github.com/pkg/errors"
)

type Daemon struct {
	api          api.Api
	socketServer socketserver.DomainSocketServer

	stopCollectorApp    chan struct{}
	stopCollectorServer chan struct{}
	stopSenderApp       chan struct{}
	stopSenderServer    chan struct{}
	errorChan           chan error
}

func NewDaemonCommand(apiClient api.Api, socketServer socketserver.DomainSocketServer) *Daemon {
	return &Daemon{
		api:          apiClient,
		socketServer: socketServer,

		stopCollectorApp:    make(chan struct{}),
		stopCollectorServer: make(chan struct{}),
		stopSenderApp:       make(chan struct{}),
		stopSenderServer:    make(chan struct{}),
		errorChan:           make(chan error),
	}
}

func (d *Daemon) Run() error {
	log.Println("Starting daemon..")

	config := sender.NewConfig(200, 5, 15*time.Second, time.Minute)

	appMetricBucket := buckets.NewBucket()
	serverMetricBucket := buckets.NewBucket()

	appMetricCollector := collectors.NewAppMetricCollector(d.socketServer, appMetricBucket)
	serverMetricCollector := collectors.NewServerMetricCollector(serverMetricBucket)

	metricSender := sender.NewSender(d.api, appMetricBucket, serverMetricBucket, config)

	go d.runAppMetricCollector(appMetricCollector)
	go d.runAppMetricSender(metricSender)

	go d.runServerMetricCollector(serverMetricCollector)
	go d.runServerMetricSender(metricSender)

	log.Printf("Daemon running with PID %d\n", os.Getpid())
	err := <-d.errorChan
	if err != nil {
		return errors.Wrap(err, "daemon exited")
	}
	return nil
}

func (d *Daemon) Shutdown() {
	d.stopSenderServer <- struct{}{}
	d.stopSenderApp <- struct{}{}
	d.stopCollectorServer <- struct{}{}
	d.stopCollectorApp <- struct{}{}
}

func (d *Daemon) runAppMetricCollector(appMetricCollector *collectors.AppMetricCollector) {
	go func() {
		<-d.stopCollectorApp
		err := appMetricCollector.Stop()
		if err != nil {
			log.Printf("Error stopping app collector: %s", err)
		}

		log.Println("Stopped app metric collector")
	}()

	if err := appMetricCollector.Start(); err != socketserver.ErrServerStopped {
		d.errorChan <- err
	}
}

func (d *Daemon) runServerMetricCollector(serverMetricCollector *collectors.ServerMetricCollector) {
	go func() {
		<-d.stopCollectorServer
		serverMetricCollector.Stop()

		log.Println("Stopped server metric collector")
	}()

	serverMetricCollector.Start()
}

func (d *Daemon) runServerMetricSender(sender *sender.Sender) {
	go func() {
		<-d.stopSenderServer
		sender.StopSendingServerMetrics()

		log.Println("Stopped server metric sender")
	}()

	sender.SendServerMetrics()
}

func (d *Daemon) runAppMetricSender(sender *sender.Sender) {
	go func() {
		<-d.stopSenderApp
		sender.StopSendingAppMetrics()

		log.Println("Stopped app metric sender")
	}()

	sender.SendAppMetrics()
}
