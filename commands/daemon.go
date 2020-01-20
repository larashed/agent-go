package commands

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/larashed/agent-go/api"
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

func (d *Daemon) Run() {
	log.Print("Starting daemon..")
	log.Printf("PID: %d", os.Getpid())
	log.Printf("Starting2...")

	config := sender.NewConfig(200, 5, 15*time.Second, time.Minute)

	appMetricBucket := buckets.NewBucket()
	serverMetricBucket := buckets.NewBucket()

	appMetricCollector := collectors.NewAppMetricCollector(d.socketServer, appMetricBucket)
	serverMetricCollector := collectors.NewServerMetricCollector(serverMetricBucket)

	metricSender := sender.NewSender(d.api, appMetricBucket, serverMetricBucket, config)

	//closeChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	errorChan := make(chan error, 1)

	go d.runAppMetricCollector(appMetricCollector, errorChan, sigChan)
	go d.runAppMetricSender(metricSender, errorChan, sigChan)

	go d.runServerMetricCollector(serverMetricCollector, errorChan, sigChan)
	go d.runServerMetricSender(metricSender, errorChan, sigChan)

	err := <-errorChan
	log.Printf("Agent exited with %s", err)
	sigChan <- os.Kill
}

func (d *Daemon) runAppMetricCollector(appMetricCollector *collectors.AppMetricCollector, errorChan chan error, sigChan chan os.Signal) {
	go func() {
		<-sigChan
		err := appMetricCollector.Stop()
		if err != nil {
			log.Printf("Error stopping app collector: %s", err)
			//errorChan <- err
		}

		log.Println("Stopped app metric collector")
	}()

	errorChan <- appMetricCollector.Start()
}

func (d *Daemon) runServerMetricCollector(serverMetricCollector *collectors.ServerMetricCollector, errorChan chan error, sigChan chan os.Signal) {
	go func() {
		<-sigChan
		serverMetricCollector.Stop()

		log.Println("Stopped server metric collector")
	}()

	serverMetricCollector.Start()

	errorChan <- nil
}

func (d *Daemon) runServerMetricSender(sender *sender.Sender, errorChan chan error, sigChan chan os.Signal) {
	go func() {
		<-sigChan
		sender.StopSendingServerMetrics()

		log.Println("Stopped server metric sender")
	}()

	sender.SendServerMetrics()

	errorChan <- nil
}

func (d *Daemon) runAppMetricSender(sender *sender.Sender, errorChan chan error, sigChan chan os.Signal) {
	go func() {
		<-sigChan
		sender.StopSendingAppMetrics()

		log.Println("Stopped app metric sender")
	}()

	sender.SendAppMetrics()

	errorChan <- nil
}
