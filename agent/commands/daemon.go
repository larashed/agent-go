package commands

import (
	"log"
	"os"

	"github.com/larashed/agent-go/agent/monitoring"
	socket_server "github.com/larashed/agent-go/agent/server"
	"github.com/larashed/agent-go/api"
)

type Daemon struct {
	api          *api.Api
	socketServer *socket_server.Server
}

func NewDaemonCommand(api *api.Api, socketServer *socket_server.Server) *Daemon {
	return &Daemon{
		api:          api,
		socketServer: socketServer,
	}
}

func (d *Daemon) Run() error {
	log.Print("Starting daemon..")
	log.Printf("PID: %d", os.Getpid())

	collector := monitoring.NewCollector(d.api, d.socketServer)
	if err := collector.Start(); err != nil {
		log.Fatal(err)
	}

	return nil
}
//
//func (d *Daemon) collect() {
//	metrics, err := d.collector.Collect()
//	j, err := json.MarshalIndent(metrics, "", "  ")
//
//	if err == nil {
//		log.Println(string(j))
//	}
//}
