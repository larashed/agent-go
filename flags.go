package main

import (
	"github.com/urfave/cli/v2"
)

const (
	ApiURLFlagName           = "api-url"
	AppEnvFlagName           = "env"
	AppIdFlagName            = "app-id"
	AppKeyFlagName           = "app-key"
	SocketAddressOldFlagName = "socket"
	SocketTypeFlagName       = "socket-type"
	SocketAddressFlagName    = "socket-address"
	ProcPathFlagName         = "path-proc"
	SysPathFlagName          = "path-sys"
	HostnameFlagName         = "hostname"
	JsonFlagName             = "json"
	LoggingLevelFlagName     = "log-level"

	CollectServerResourcesFlagName = "collect-server-resources"
)

var (
	ApiUrlFlag = &cli.StringFlag{
		Name:  ApiURLFlagName,
		Usage: "Larashed API URL",
		Value: "https://api.larashed.com/",
	}
	AppEnvFlag = &cli.StringFlag{
		Name:    AppEnvFlagName,
		Aliases: []string{"app-env"},
		Usage:   "Application's environment name",
	}
	AppIdFlag = &cli.StringFlag{
		Name:  AppIdFlagName,
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
	JsonFlag = &cli.BoolFlag{
		Name:  JsonFlagName,
		Usage: "Output JSON",
	}
	CollectServerResourcesFlag = &cli.BoolFlag{
		Name:  CollectServerResourcesFlagName,
		Usage: "Collect server resource metrics",
		Value: true,
	}
)
