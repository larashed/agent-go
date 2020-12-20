// copied from https://github.com/docker/cli/blob/master/cli/command/container/stats.go

package container

import (
	"context"
	"sync"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

type statsOptions struct {
	all      bool
	noStream bool
	noTrunc  bool
	format   string
}

// GetStats collects docker container stats
func GetStats(dockerCli *client.Client) ([]StatsEntry, error) {
	var opts = statsOptions{
		all:      false,
		noStream: true,
		noTrunc:  false,
		format:   "",
	}
	closeChan := make(chan error)
	ctx, doneFunc := context.WithCancel(context.Background())

	// monitorContainerEvents watches for container creation and removal (only
	// used when calling `docker stats` without arguments).
	monitorContainerEvents := func(started chan<- struct{}, c chan events.Message) {
		f := filters.NewArgs()
		f.Add("type", "container")
		options := types.EventsOptions{
			Filters: f,
		}

		eventq, errq := dockerCli.Events(ctx, options)

		// Whether we successfully subscribed to eventq or not, we can now
		// unblock the main goroutine.
		close(started)

		for {
			select {
			case event := <-eventq:
				c <- event
			case err := <-errq:
				closeChan <- err
				return
			}
		}
	}

	// Get the daemonOSType if not set already
	if daemonOSType == "" {
		svctx := context.Background()
		sv, err := dockerCli.ServerVersion(svctx)
		if err != nil {
			doneFunc()

			return nil, err
		}
		daemonOSType = sv.Os
	}

	// waitFirst is a WaitGroup to wait first stat data's reach for each container
	waitFirst := &sync.WaitGroup{}

	cStats := stats{}
	// getContainerList simulates creation event for all previously existing
	// containers (only used when calling `docker stats` without arguments).
	getContainerList := func() {
		options := types.ContainerListOptions{
			All: opts.all,
		}
		cs, err := dockerCli.ContainerList(ctx, options)
		if err != nil {
			log.Trace().Err(err).Msg("getContainerList")
			closeChan <- err
		}
		for _, container := range cs {
			s := NewStats(container.ID[:12])
			if cStats.add(s) {
				waitFirst.Add(1)
				go collect(ctx, s, dockerCli, false, waitFirst)
			}
		}
	}

	started := make(chan struct{})
	eh := command.InitEventHandler()

	eh.Handle("start", func(e events.Message) {
		s := NewStats(e.ID[:12])
		if cStats.add(s) {
			waitFirst.Add(1)
			go collect(ctx, s, dockerCli, false, waitFirst)
		}
	})

	eh.Handle("die", func(e events.Message) {
		cStats.remove(e.ID[:12])
	})

	eventChan := make(chan events.Message)
	go eh.Watch(eventChan)
	go monitorContainerEvents(started, eventChan)
	defer func() {
		doneFunc()
		close(eventChan)
	}()
	<-started

	// Start a short-lived goroutine to retrieve the initial list of
	// containers.
	getContainerList()

	// make sure each container get at least one valid stat data
	waitFirst.Wait()

	ccstats := []StatsEntry{}
	cStats.mu.Lock()
	for _, c := range cStats.cs {
		ccstats = append(ccstats, c.GetStatistics())
	}
	cStats.mu.Unlock()

	return ccstats, nil
}
