package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/commands"
	socket_server "github.com/larashed/agent-go/server"
)

type App struct {
	app *cli.App
}

const (
	ApiURLFlagName = "api-url"
	AppEnvFlagName = "env"
	AppIdFlagName  = "app-id"
	AppKeyFlagName = "app-key"
	SocketFlagName = "socket"
	JsonFlagName   = "json"
)

var (
	ApiUrlFlag = &cli.StringFlag{
		Name:  ApiURLFlagName,
		Usage: "Larashed API URL",
		Value: "https://api.larashed.com/",
	}
	AppEnvFlag = &cli.StringFlag{
		Name:  AppEnvFlagName,
		Usage: "Application environment",
	}
	AppIdFlag = &cli.StringFlag{
		Name:  AppIdFlagName,
		Usage: "Your application API ID token",
	}
	AppKeyFlag = &cli.StringFlag{
		Name:  AppKeyFlagName,
		Usage: "Your application API secret key",
	}
	SocketFlag = &cli.StringFlag{
		Name:  SocketFlagName,
		Usage: "Location of the unix socket",
	}
	JsonFlag = &cli.BoolFlag{
		Name:  JsonFlagName,
		Usage: "Output JSON",
	}
)

func NewApp() *App {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "daemon",
				Usage: "run agent in daemon mode",
				Action: func(c *cli.Context) error {
					apiClient := api.NewClient(
						c.String(ApiURLFlagName),
						c.String(AppEnvFlagName),
						c.String(AppIdFlagName),
						c.String(AppKeyFlagName),
					)
					server := socket_server.NewServer(c.String(SocketFlagName))

					return commands.NewDaemonCommand(apiClient, server).Run()
				},
				Flags: []cli.Flag{
					ApiUrlFlag,
					AppEnvFlag,
					AppIdFlag,
					AppKeyFlag,
					SocketFlag,
				},
			},
			{
				Name:  "version",
				Usage: "agent version",
				Action: func(c *cli.Context) error {
					commands.NewVersionCommand(c.Bool(JsonFlagName))

					return nil
				},
				Flags: []cli.Flag{
					JsonFlag,
				},
			},
		},
	}

	return &App{app}
}

func (a *App) Run() error {
	return a.app.Run(os.Args)
}
