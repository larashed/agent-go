package collectors

import (
	"context"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"

	"github.com/larashed/agent-go/monitoring/metrics"
)

func DockerContainersApi() (services []metrics.Service, err error) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	for i := 0; i < len(containers); i++ {
		stats, err := cli.ContainerStats(ctx, containers[i].ID, false)
		if err != nil {
			if &stats != nil {
				body, err := ioutil.ReadAll(stats.Body)

				if err != nil {
					panic(err.Error())
				}

				log.Trace().Msg(string(body))
			}
		}
	}

	return services, nil
}

// https://github.com/docker/cli/blob/f784262d078e7275a383b96da8f24925264c11fc/cli/command/container/formatter_stats.go
func parseContainerList(output string) (containers []metrics.Container) {
	return containers
}
