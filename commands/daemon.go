package commands

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/monitoring"
	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/collectors"
	"github.com/larashed/agent-go/monitoring/metrics"
	"github.com/larashed/agent-go/monitoring/sender"
	socketserver "github.com/larashed/agent-go/server"
)

type Daemon struct {
	api          api.Api
	socketServer socketserver.DomainSocketServer

	stopSocketServer    chan struct{}
	stopCollectorServer chan struct{}
	stopSenderApp       chan struct{}
	stopSenderServer    chan struct{}
	errorChan           chan error
}

func NewDaemonCommand(apiClient api.Api, socketServer socketserver.DomainSocketServer) *Daemon {
	return &Daemon{
		api:          apiClient,
		socketServer: socketServer,

		stopSocketServer:    make(chan struct{}),
		stopCollectorServer: make(chan struct{}),
		stopSenderApp:       make(chan struct{}),
		stopSenderServer:    make(chan struct{}),
		errorChan:           make(chan error),
	}
}

func (d *Daemon) Run() error {
	log.Println("Starting daemon..")

	config := monitoring.NewConfig(200, 5, 10)

	appMetricBucket := buckets.NewAppMetricBucket()
	serverMetricBucket := buckets.NewServerMetricBucket()

	serverMetricCollector := collectors.NewServerMetricCollector(serverMetricBucket, config.ServerMetricCollectionInterval)

	metricSender := sender.NewSender(d.api, appMetricBucket, serverMetricBucket, config)

	go d.runSocketServer(appMetricBucket)
	go d.runAppMetricSender(metricSender)

	go d.runServerMetricCollector(serverMetricCollector)
	go d.runServerMetricSender(metricSender)

	log.Printf("Daemon running with PID %d\n", os.Getpid())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	select {
	case err := <-d.errorChan:
		return errors.Wrap(err, "daemon exited")
	case sig := <-sigChan:
		log.Printf("Daemon received %s signal", sig.String())
		d.Shutdown()
		return nil
	}
}

func (d *Daemon) Shutdown() {
	log.Printf("Stopping daemon..")

	d.stopSenderServer <- struct{}{}
	d.stopSenderApp <- struct{}{}
	d.stopCollectorServer <- struct{}{}
	d.stopSocketServer <- struct{}{}

	time.Sleep(100 * time.Millisecond)
	log.Printf("Daemon stopped")

	os.Exit(1)
}

func (d *Daemon) runSocketServer(bucket *buckets.AppMetricBucket) {
	go func() {
		<-d.stopSocketServer
		err := d.socketServer.Stop()
		if err != nil {
			log.Printf("Error stopping socker server: %s", err)
		}

		log.Println("Stopped socket service")
	}()

	handleSocketMessage := func(message string) {
		log.Printf("received %s with length %d", message, len(message))
		if message == socketserver.QuitMessage {
			d.Shutdown()

			return
		}

		bucket.Add(metrics.NewAppMetric(message))
	}

	if err := d.socketServer.Start(handleSocketMessage); err != socketserver.ErrServerStopped {
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
