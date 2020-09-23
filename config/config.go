package config

import (
	"encoding/json"
)

type Config struct {
	ApiUrl         string
	AppEnvironment string
	AppId          string
	AppKey         string
	InDocker       bool
	Hostname       string
	SocketType     string
	SocketAddress  string
	LogLevel       string
	PathProcfs     string
	PathSysfs      string
}

func (c *Config) String() string {
	j, _ := json.Marshal(c)

	return string(j)
}
