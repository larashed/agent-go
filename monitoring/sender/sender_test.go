package sender

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/monitoring"
	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/metrics"
)

type apiClient struct {
	returnError        bool
	appMetricCallsMade uint64
}

func (ac *apiClient) SendServerMetrics(data string) (*api.Response, error) {
	if ac.returnError {
		return nil, errors.New("error")
	}

	return nil, nil
}

func (ac *apiClient) SendAppMetrics(data string) (*api.Response, error) {
	if ac.returnError {
		ac.appMetricCallsMade++

		return nil, errors.New("error")
	}
	ac.appMetricCallsMade++
	return nil, nil
}

// this test will fail at some point
// refactor to mock timers...
func TestSender_SendAppMetrics(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	apc := &apiClient{returnError: false}
	appBucket := buckets.NewAppMetricBucket()
	serverBucket := buckets.NewServerMetricBucket()
	cfg := &monitoring.Config{
		AppMetricSendCount:              1000,
		AppMetricSendInterval:           time.Millisecond * 200,
		AppMetricSleepDurationOnFailure: time.Millisecond * 50,
		AppMetricOverflowLimit:          8000,
		AppMetricOverflowLimitBytes:     0,
		ServerMetricSendInterval:        0,
	}
	s := NewSender(
		apc,
		appBucket,
		serverBucket,
		cfg,
		true,
	)
	s.StartAppMetricSend()

	for i := 0; i < 10; i++ {
		go fillBucket(appBucket, 1000)
		time.Sleep(time.Millisecond * 5)
	}

	time.Sleep(time.Second * 3)

	im := s.GetInternalMetrics()

	limit := uint64(10000)

	assert.Equal(t, limit, im.AppMetricsReceived, "bucket items received")
	assert.Equal(t, limit, im.AppMetricsSent, "bucket items sent")

	assert.Equal(t, limit/uint64(cfg.AppMetricSendCount), im.APICallsSuccess, "API calls made by sender")
	assert.Equal(t, limit/uint64(cfg.AppMetricSendCount), apc.appMetricCallsMade, "mock API calls made")
	assert.Equal(t, uint64(0), im.APICallsEmpty, "empty API calls made")

	s.StopSendingAppMetrics()

	spew.Dump(im)
}

func fillBucket(bucket *buckets.AppMetricBucket, limit int) {
	for i := 0; i < limit; i++ {
		bucket.Add(&metrics.AppMetric{})
	}
}
