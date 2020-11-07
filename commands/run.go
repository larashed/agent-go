package commands

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/config"
	"github.com/larashed/agent-go/monitoring"
	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/collectors"
	"github.com/larashed/agent-go/monitoring/metrics"
	"github.com/larashed/agent-go/monitoring/sender"
	socketserver "github.com/larashed/agent-go/server"
)

// RunCommand defines the agent's run command
type RunCommand struct {
	config       *config.Config
	api          api.Api
	socketServer *socketserver.Server

	stopSocketServer    chan struct{}
	stopCollectorServer chan struct{}
	stopSenderApp       chan struct{}
	stopSenderServer    chan struct{}
	errorChan           chan error
}

// NewRunCommand creates an instance of `RunCommand`
func NewRunCommand(cfg *config.Config, apiClient api.Api, socketServer *socketserver.Server) *RunCommand {
	return &RunCommand{
		config:       cfg,
		api:          apiClient,
		socketServer: socketServer,

		stopSocketServer:    make(chan struct{}),
		stopCollectorServer: make(chan struct{}),
		stopSenderApp:       make(chan struct{}),
		stopSenderServer:    make(chan struct{}),
		errorChan:           make(chan error),
	}
}

// Run starts the agent
func (d *RunCommand) Run() error {
	log.Info().Msgf("Starting agent v%s", config.GitTag)
	log.Trace().Msgf("Config: %s", d.config.String())
	log.Info().Msgf("Agent running with PID %d", os.Getpid())

	cfg := monitoring.NewConfig(200, 10, 10)

	appMetricBucket := buckets.NewAppMetricBucket()
	serverMetricBucket := buckets.NewServerMetricBucket()

	serverMetricCollector := collectors.NewServerMetricCollector(
		serverMetricBucket,
		cfg.ServerMetricCollectionInterval,
		d.config.InDocker,
	)

	metricSender := sender.NewSender(d.api, appMetricBucket, serverMetricBucket, cfg)

	if d.config.CollectServerResources {
		go d.runServerMetricCollector(serverMetricCollector)
		go d.runServerMetricSender(metricSender)
	} else {
		log.Info().Msg("[Disabled] Server resource collection")
	}

	if d.config.CollectAppMetrics {
		go d.runSocketServer(appMetricBucket)
		go d.runAppMetricSender(metricSender)
		log.Info().Msgf("Socket address: %s://%s", d.config.SocketType, d.config.SocketAddress)
	} else {
		log.Info().Msg("[Disabled] Application metric collection")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	select {
	case err := <-d.errorChan:
		return errors.Wrap(err, "Agent exited")
	case sig := <-sigChan:
		log.Info().Msgf("Agent received exit signal: %s", sig.String())
		d.Shutdown()
		return nil
	}
}

// Shutdown stops the agent
func (d *RunCommand) Shutdown() {
	log.Info().Msg("Stopping agent")

	if d.config.CollectServerResources {
		d.stopSenderServer <- struct{}{}
		d.stopCollectorServer <- struct{}{}
	}

	if d.config.CollectAppMetrics {
		d.stopSenderApp <- struct{}{}
		d.stopSocketServer <- struct{}{}
	}

	time.Sleep(100 * time.Millisecond)

	log.Info().Msg("Agent stopped")

	os.Exit(1)
}

func (d *RunCommand) runSocketServer(bucket *buckets.AppMetricBucket) {
	go func() {
		<-d.stopSocketServer
		err := d.socketServer.Stop()
		if err != nil {
			log.Info().Msgf("Error stopping socket server: %s", err)
		}

		log.Info().Msg("Stopped socket server")
	}()

	handleSocketMessage := func(message string) {
		if message == socketserver.QuitMessage {
			d.Shutdown()

			return
		}

		bucket.Add(metrics.NewAppMetric(message))
	}

	log.Info().Msg("Starting socket server")
	if err := d.socketServer.Start(handleSocketMessage); err != socketserver.ErrServerStopped {
		d.errorChan <- err
	}
}

func (d *RunCommand) runServerMetricCollector(serverMetricCollector *collectors.ServerMetricCollector) {
	go func() {
		<-d.stopCollectorServer
		serverMetricCollector.Stop()

		log.Info().Msg("Stopped server metric collector")
	}()

	log.Info().Msg("Starting server metric collection")
	serverMetricCollector.Start()
}

func (d *RunCommand) runServerMetricSender(sender *sender.Sender) {
	go func() {
		<-d.stopSenderServer
		sender.StopSendingServerMetrics()

		log.Info().Msg("Stopped server metric sender")
	}()

	log.Info().Msg("Starting server metric sender")
	sender.SendServerMetrics()
}

func (d *RunCommand) runAppMetricSender(sender *sender.Sender) {
	go func() {
		<-d.stopSenderApp
		sender.StopSendingAppMetrics()

		log.Info().Msg("Stopped app metric sender")
	}()

	log.Info().Msg("Starting app metric sender")
	sender.SendAppMetrics()
}
