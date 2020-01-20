package agent

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/larashed/agent-go/agent/commands"
	socket_server "github.com/larashed/agent-go/agent/server"
	"github.com/larashed/agent-go/api"
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
)

var (
	ApiUrlFlag = &cli.StringFlag{
		Name:        ApiURLFlagName,
		Usage:       "Larashed API URL",
		DefaultText: "https://api.larashed.com/",
	}
	AppEnvFlag = &cli.StringFlag{
		Name:        AppEnvFlagName,
		Usage:       "Application environment",
		DefaultText: "https://api.larashed.com/",
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
		Usage: "location of the unix socket",
	}
)

func NewApp() *App {
	flags := []cli.Flag{
		ApiUrlFlag,
		AppEnvFlag,
		AppIdFlag,
		AppKeyFlag,
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "daemon",
				Usage: "run agent in daemon mode",
				Action: func(c *cli.Context) error {
					api := api.NewApi(
						c.String(ApiURLFlagName),
						c.String(AppEnvFlagName),
						c.String(AppIdFlagName),
						c.String(AppKeyFlagName),
					)

					server := socket_server.NewServer(c.String(SocketFlagName))

					return commands.NewDaemonCommand(
						api,
						server,
					).Run()
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
				Name:  "diagnostics",
				Usage: "agent diagnostics",
				Action: func(c *cli.Context) error {
					return commands.RunDiagnostics()
				},
				Flags: flags,
			},
		},
	}

	return &App{app}
}

func (a *App) Run() error {
	return a.app.Run(os.Args)
}
