package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/commands"
	"github.com/larashed/agent-go/config"
	"github.com/larashed/agent-go/log"
	socket_server "github.com/larashed/agent-go/server"
)

type App struct {
	app *cli.App
}

func NewApp() *App {
	app := &cli.App{
		Name:        "Larashed",
		Usage:       "Monitoring agent",
		Description: "Monitoring agent for https://larashed.com",
		Commands: []*cli.Command{
			{
				Name:    "run",
				Usage:   "Starts server monitoring & socket server",
				Aliases: []string{"daemon"},
				Action: func(c *cli.Context) error {
					cfg := newConfig(c)
					setEnvVariables(cfg)

					// validate required flags and output error message with help
					if !validateConfig(cfg.SocketAddress, SocketAddressFlagName) ||
						!validateConfig(cfg.AppId, AppIdFlagName) ||
						!validateConfig(cfg.AppKey, AppKeyFlagName) ||
						!validateConfig(cfg.AppEnvironment, AppEnvFlagName) {
						return cli.ShowCommandHelp(c, "run")
					}

					log.Bootstrap(log.ParseLoggingLevel(cfg.LogLevel))

					apiClient := api.NewClient(
						cfg.ApiUrl,
						cfg.AppEnvironment,
						cfg.AppId,
						cfg.AppKey,
						cfg.Hostname,
					)

					server := socket_server.NewServer(cfg.SocketType, cfg.SocketAddress)

					return commands.NewRunCommand(cfg, apiClient, server).Run()
				},
				Flags: []cli.Flag{
					SocketTypeFlag,
					SocketAddressFlag,
					OldSocketAddressFlag,
					ApiUrlFlag,
					AppEnvFlag,
					AppIdFlag,
					AppKeyFlag,
					ProcPathFlag,
					SysPathFlag,
					HostnameFlag,
					LoggingLevelFlag,
					CollectServerResourcesFlag,
				},
			},
			{
				Name:  "version",
				Usage: "print agent version",
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

func newConfig(c *cli.Context) *config.Config {
	cfg := &config.Config{
		ApiUrl: c.String(ApiURLFlagName),

		PathProcfs: c.String(ProcPathFlagName),
		PathSysfs:  c.String(SysPathFlagName),

		Hostname: c.String(HostnameFlagName),
		InDocker: os.Getenv("DOCKER_BUILD") == "1",

		LogLevel: c.String(LoggingLevelFlagName),

		AppEnvironment: c.String(AppEnvFlagName),
		AppId:          c.String(AppIdFlagName),
		AppKey:         c.String(AppKeyFlagName),

		SocketAddress: c.String(SocketAddressFlagName),
		SocketType:    c.String(SocketTypeFlagName),

		CollectServerResources: c.Bool(CollectServerResourcesFlagName),
	}

	if len(cfg.SocketAddress) == 0 {
		cfg.SocketAddress = c.String(SocketAddressOldFlagName)
	}

	if len(cfg.Hostname) == 0 {
		hostname, err := os.Hostname()
		if err == nil {
			cfg.Hostname = hostname
		}
	}

	return cfg
}

func validateConfig(value, flag string) bool {
	if len(value) == 0 {
		println("Incorrect Usage: --" + flag + " is required\n")

		return false
	}

	return true
}

// used by github.com/shirou/gopsutil and internal code
func setEnvVariables(cfg *config.Config) {
	os.Setenv("HOST_PROC", cfg.PathProcfs)
	os.Setenv("HOST_SYS", cfg.PathSysfs)
}
