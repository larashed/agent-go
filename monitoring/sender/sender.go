package sender

import (
	"math"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/monitoring"
	"github.com/larashed/agent-go/monitoring/buckets"
)

// InternalMetrics ...
type InternalMetrics struct {
	AppMetricsReceived uint64
	AppMetricsSent     uint64
	APICallsSuccess    uint64
	APICallsFail       uint64
	APICallsEmpty      uint64
	DiscardedItems     uint64
}

// Sender defines a metric sender
type Sender struct {
	api api.Api

	appMetricBucket    *buckets.AppMetricBucket
	serverMetricBucket *buckets.ServerMetricBucket

	config *monitoring.Config

	sentAt time.Time
	mutex  sync.RWMutex

	stopServerMetricSend chan int
	stopAppMetricSend    chan int

	appMetricFails int

	internalMetrics *InternalMetrics
}

// NewSender creates an instance of `Sender`
func NewSender(
	api api.Api,
	appMetricBucket *buckets.AppMetricBucket,
	serverMetricBucket *buckets.ServerMetricBucket,
	config *monitoring.Config,
	inspect bool) *Sender {
	sender := &Sender{
		api,
		appMetricBucket,
		serverMetricBucket,
		config,
		time.Now(),
		sync.RWMutex{},
		make(chan int, 0),
		make(chan int, 0),
		0,
		nil,
	}

	if inspect {
		sender.internalMetrics = &InternalMetrics{
			AppMetricsReceived: 0,
			AppMetricsSent:     0,
			APICallsSuccess:    0,
			APICallsFail:       0,
			APICallsEmpty:      0,
			DiscardedItems:     0,
		}
	}

	return sender
}

// StopSendingServerMetrics stops sending server metrics
func (s *Sender) StopSendingServerMetrics() {
	s.stopServerMetricSend <- 1
}

// GetInternalMetrics returns internal sender metrics
func (s *Sender) GetInternalMetrics() *InternalMetrics {
	return s.internalMetrics
}

// StopSendingAppMetrics stops sending app metrics
func (s *Sender) StopSendingAppMetrics() {
	s.stopAppMetricSend <- 1
	s.stopAppMetricSend <- 1
	s.stopAppMetricSend <- 1
}

// StartServerMetricSend sends collected server metrics
func (s *Sender) StartServerMetricSend() {
	go func() {
		for {
			select {
			case metric := <-s.serverMetricBucket.Channel:
				log.Trace().
					Str("metric", "server").
					Msgf("Server metrics: %s", metric.String())

				_, err := s.api.SendServerMetrics(metric.String())
				if err != nil {
					log.Err(err).Msg("Failed to send server metrics")
				}
			case <-s.stopServerMetricSend:
				return
			}
		}
	}()
}

// StartAppMetricSend sends collected app metrics
func (s *Sender) StartAppMetricSend() {
	go s.sendOnBucketFill()
	go s.sendPeriodically()
	go s.clearOverflowingMetrics()
}

func (s *Sender) clearOverflowingMetrics() {
	ticker := time.NewTicker(s.config.AppMetricSendInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopAppMetricSend:
			ticker.Stop()
			return
		case t := <-ticker.C:
			if t.Sub(s.sentAt) > s.config.AppMetricSendInterval {
				count := s.appMetricBucket.Count()

				if uint64(count) >= s.config.AppMetricOverflowLimit {
					if s.internalMetrics != nil {
						s.internalMetrics.DiscardedItems += s.config.AppMetricOverflowLimit
					}

					log.Debug().
						Int("total app metrics", count).
						Uint64("discarding", s.config.AppMetricOverflowLimit).
						Msg("discarding")

					s.appMetricBucket.Discard(int(s.config.AppMetricOverflowLimit))
				}
			}
		}
	}
}

func (s *Sender) sendPeriodically() {
	ticker := time.NewTicker(s.config.AppMetricSendInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopAppMetricSend:
			ticker.Stop()
			return
		case t := <-ticker.C:
			// send data if the bucket is not empty and there hasn't been a send in n seconds
			if t.Sub(s.sentAt) > s.config.AppMetricSendInterval {
				if count := s.appMetricBucket.Count(); count > 0 {
					rounds := int(math.Ceil(float64(count) / float64(s.config.AppMetricSendCount)))

					log.Debug().
						Int("app metrics", count).
						Int("rounds", rounds).
						Msg("sending periodic metrics")

					for i := 0; i < rounds; i++ {
						go s.sendAppMetrics(
							s.appMetricBucket.Extract(s.config.AppMetricSendCount), "periodic",
						)
					}
				}
			}
		}
	}
}

func (s *Sender) sendOnBucketFill() {
	for {
		select {
		case <-s.appMetricBucket.Channel:
			if s.internalMetrics != nil {
				s.internalMetrics.AppMetricsReceived++
			}

			if count := s.appMetricBucket.Count(); count >= s.config.AppMetricSendCount {
				log.Debug().
					Int("app metrics", count).
					Msg("sending filled bucket metrics")

				go s.sendAppMetrics(
					s.appMetricBucket.Extract(s.config.AppMetricSendCount), "fill",
				)
			}
		case <-s.stopAppMetricSend:
			return
		}
	}
}

func (s *Sender) sendAppMetrics(bkt *buckets.AppMetricBucket, ctx string) {
	// not sure if it's the best idea
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if bkt.Count() == 0 {
		if s.internalMetrics != nil {
			s.internalMetrics.APICallsEmpty++
		}

		log.Error().
			Str("context", ctx).
			Msg("making empty API call")

		return
	}

	_, err := s.api.SendAppMetrics(bkt.String())
	if err != nil {
		go func() {
			time.Sleep(s.config.AppMetricSleepDurationOnFailure)
			s.appMetricBucket.Merge(bkt)

			log.Debug().
				Int("total", s.appMetricBucket.Count()).
				Int("bucket", bkt.Count()).
				Str("context", ctx).
				Err(err).
				Msg("returned failed to send metrics to bucket")
		}()

		if s.internalMetrics != nil {
			s.internalMetrics.APICallsFail++
		}

		return
	}

	if s.internalMetrics != nil {
		s.internalMetrics.APICallsSuccess++
		s.internalMetrics.AppMetricsSent += uint64(bkt.Count())
	}

	s.sentAt = time.Now()
}