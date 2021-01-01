package monitoring

import (
	"time"
)

// Config holds metric collection configuration
type Config struct {
	// trigger metric send if this number is reached
	AppMetricSendCount int
	// trigger metric send if this duration is reached
	AppMetricSendInterval time.Duration
	// clear metrics when this number is reached
	AppMetricOverflowLimit uint64
	// clear metrics when this number of bytes
	AppMetricOverflowLimitBytes uint64
	// sleep before adding back metrics back into the bucket on API failure
	AppMetricSleepDurationOnFailure time.Duration
	// trigger server metric collection
	ServerMetricSendInterval time.Duration
}
