package main

import (
	"github.com/urfave/cli/v2"
)

const (
	ApiUrlFlagName           = "api-url" //nolint:golint
	AppEnvFlagName           = "env"
	AppIDFlagName            = "app-id"
	AppKeyFlagName           = "app-key"
	SocketAddressOldFlagName = "socket"
	SocketTypeFlagName       = "socket-type"
	SocketAddressFlagName    = "socket-address"
	ProcPathFlagName         = "path-proc"
	SysPathFlagName          = "path-sys"
	HostnameFlagName         = "hostname"
	JSONFlagName             = "json"
	LoggingLevelFlagName     = "log-level"

	CollectServerResourcesFlagName    = "collect-server-resources"
	CollectApplicationMetricsFlagName = "collect-application-metrics"
)

var (
	ApiUrlFlag = &cli.StringFlag{ //nolint:golint
		Name:  ApiUrlFlagName,
		Usage: "Larashed API URL",
		Value: "https://api.larashed.com/",
	}
	AppEnvFlag = &cli.StringFlag{
		Name:    AppEnvFlagName,
		Aliases: []string{"app-env"},
		Usage:   "Application's environment name",
	}
	AppIDFlag = &cli.StringFlag{
		Name:  AppIDFlagName,
		Usage: "Your application's ID",
	}
	AppKeyFlag = &cli.StringFlag{
		Name:  AppKeyFlagName,
		Usage: "Your application's secret key",
	}
	SocketTypeFlag = &cli.StringFlag{
		Name:  SocketTypeFlagName,
		Usage: "Socket type (unix, tcp)",
		Value: "unix",
	}
	SocketAddressFlag = &cli.StringFlag{
		Name:  SocketAddressFlagName,
		Usage: "Socket address",
	}
	OldSocketAddressFlag = &cli.StringFlag{
		Name:  SocketAddressOldFlagName,
		Usage: "Socket address (deprecated, use --socket-address instead)",
	}
	LoggingLevelFlag = &cli.StringFlag{
		Name:  LoggingLevelFlagName,
		Usage: "Logging level (info, debug, trace)",
		Value: "debug",
	}
	ProcPathFlag = &cli.StringFlag{
		Name:  ProcPathFlagName,
		Usage: "Kernel & process file path",
		Value: "/proc",
	}
	SysPathFlag = &cli.StringFlag{
		Name:  SysPathFlagName,
		Usage: "System component file path",
		Value: "/sys",
	}
	HostnameFlag = &cli.StringFlag{
		Name:  HostnameFlagName,
		Usage: "Hostname",
	}
	JSONFlag = &cli.BoolFlag{
		Name:  JSONFlagName,
		Usage: "Output JSON",
	}
	CollectServerResourcesFlag = &cli.BoolFlag{
		Name:  CollectServerResourcesFlagName,
		Usage: "Collect server resource metrics",
		Value: true,
	}
	CollectApplicationMetricsFlag = &cli.BoolFlag{
		Name:  CollectApplicationMetricsFlagName,
		Usage: "Collect application metrics",
		Value: true,
	}
)
