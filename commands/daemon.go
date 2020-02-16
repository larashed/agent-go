package commands

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

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
	log.Info().Msg("Starting daemon command")

	config := monitoring.NewConfig(200, 10, 10)

	appMetricBucket := buckets.NewAppMetricBucket()
	serverMetricBucket := buckets.NewServerMetricBucket()

	serverMetricCollector := collectors.NewServerMetricCollector(serverMetricBucket, config.ServerMetricCollectionInterval)

	metricSender := sender.NewSender(d.api, appMetricBucket, serverMetricBucket, config)

	go d.runSocketServer(appMetricBucket)
	go d.runAppMetricSender(metricSender)

	go d.runServerMetricCollector(serverMetricCollector)
	go d.runServerMetricSender(metricSender)

	log.Info().Msgf("daemon running with PID %d", os.Getpid())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	select {
	case err := <-d.errorChan:
		return errors.Wrap(err, "daemon exited")
	case sig := <-sigChan:
		log.Info().Msgf("daemon received %s signal", sig.String())
		d.Shutdown()
		return nil
	}
}

func (d *Daemon) Shutdown() {
	log.Info().Msg("stopping daemon..")

	d.stopSenderServer <- struct{}{}
	d.stopSenderApp <- struct{}{}
	d.stopCollectorServer <- struct{}{}
	d.stopSocketServer <- struct{}{}

	time.Sleep(100 * time.Millisecond)

	log.Info().Msg("daemon stopped")

	os.Exit(1)
}

func (d *Daemon) runSocketServer(bucket *buckets.AppMetricBucket) {
	go func() {
		<-d.stopSocketServer
		err := d.socketServer.Stop()
		if err != nil {
			log.Info().Msgf("error stopping socker server: %s", err)
		}

		log.Info().Msg("stopped socket service")
	}()

	handleSocketMessage := func(message string) {
		if message == socketserver.QuitMessage {
			d.Shutdown()

			return
		}

		bucket.Add(metrics.NewAppMetric(message))
	}

	log.Info().Msg("starting socket service")
	if err := d.socketServer.Start(handleSocketMessage); err != socketserver.ErrServerStopped {
		d.errorChan <- err
	}
}

func (d *Daemon) runServerMetricCollector(serverMetricCollector *collectors.ServerMetricCollector) {
	go func() {
		<-d.stopCollectorServer
		serverMetricCollector.Stop()

		log.Info().Msg("stopped server metric collector")
	}()

	log.Info().Msg("starting server metric collection")
	serverMetricCollector.Start()
}

func (d *Daemon) runServerMetricSender(sender *sender.Sender) {
	go func() {
		<-d.stopSenderServer
		sender.StopSendingServerMetrics()

		log.Info().Msg("stopped server metric sender")
	}()

	log.Info().Msg("starting server metric sender")
	sender.SendServerMetrics()
}

func (d *Daemon) runAppMetricSender(sender *sender.Sender) {
	go func() {
		<-d.stopSenderApp
		sender.StopSendingAppMetrics()

		log.Info().Msg("stopped app metric sender")
	}()

	log.Info().Msg("starting app metric sender")
	sender.SendAppMetrics()
}
