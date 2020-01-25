package sender

import (
	"time"
)

type Config struct {
	appBucketLimit    int
	serverBucketLimit int

	appMetricSendingInterval    time.Duration
	serverMetricSendingInterval time.Duration
}

func NewConfig(
	appBucketLimit int,
	serverBucketLimit int,
	appMetricSendingInterval time.Duration,
	serverMetricSendingInterval time.Duration) *Config {
	return &Config{
		appBucketLimit,
		serverBucketLimit,
		appMetricSendingInterval,
		serverMetricSendingInterval,
	}
}
