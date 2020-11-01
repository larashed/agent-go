package config

import (
	"encoding/json"
)

// Config holds agent configuration
type Config struct {
	ApiUrl         string //nolint:golint
	AppEnvironment string
	AppId          string //nolint:golint
	AppKey         string
	InDocker       bool
	Hostname       string
	SocketType     string
	SocketAddress  string
	LogLevel       string
	PathProcfs     string
	PathSysfs      string

	CollectServerResources bool
}

func (c *Config) String() string {
	j, _ := json.Marshal(c)

	return string(j)
}
