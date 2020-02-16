package sender

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/monitoring"
	"github.com/larashed/agent-go/monitoring/buckets"
	"github.com/larashed/agent-go/monitoring/metrics"
)

type Sender struct {
	api api.Api

	appMetricBucket    *buckets.AppMetricBucket
	serverMetricBucket *buckets.ServerMetricBucket

	config *monitoring.Config

	sentAt time.Time
	mutex  sync.RWMutex

	stopServerMetricSend chan int
	stopAppMetricSend    chan int
}

func NewSender(
	api api.Api,
	appMetricBucket *buckets.AppMetricBucket,
	serverMetricBucket *buckets.ServerMetricBucket,
	config *monitoring.Config) *Sender {
	return &Sender{
		api,
		appMetricBucket,
		serverMetricBucket,
		config,
		time.Now(),
		sync.RWMutex{},
		make(chan int, 0),
		make(chan int, 0),
	}
}

func (s *Sender) StopSendingServerMetrics() {
	s.stopServerMetricSend <- 1
}

func (s *Sender) StopSendingAppMetrics() {
	s.stopAppMetricSend <- 1
	s.stopAppMetricSend <- 1
}

func (s *Sender) SendServerMetrics() {
	go func() {
		for {
			select {
			case <-s.serverMetricBucket.Channel:
				go s.aggregateAndSendServerMetrics()
			case <-s.stopServerMetricSend:
				return
			}
		}
	}()
}

func (s *Sender) SendAppMetrics() {
	go func() {
		for {
			select {
			case <-s.appMetricBucket.Channel:
				if count := s.appMetricBucket.Count(); count >= s.config.AppBucketLimit {
					log.Debug().Str("metric", "app").
						Int("metrics", count).
						Msg("sending from channel read")

					go s.sendAppMetrics()
				}
			case <-s.stopAppMetricSend:
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopAppMetricSend:
				ticker.Stop()
				return
			case t := <-ticker.C:
				// send data if the bucket is not empty and there hasn't been a send in n seconds
				if t.Sub(s.sentAt).Seconds() > float64(s.config.AppBucketNotFillingSeconds) {
					if count := s.appMetricBucket.Count(); count > 0 {
						log.Debug().Str("metric", "app").
							Int("count", count).
							Msg("sending from ticker read")

						go s.sendAppMetrics()
					}
				}
			}
		}
	}()
}

func (s *Sender) updateSentAt() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Debug().Msg("Updating sentAt")
	s.sentAt = time.Now()
}

func (s *Sender) aggregateAndSendServerMetrics() {
	minutes := s.serverMetricBucket.Minutes()

	// don't send if the bucket contains only one (the current) minute
	if len(minutes) <= 1 {
		return
	}

	for _, minute := range minutes {
		ms := s.serverMetricBucket.Metrics(minute)

		metric := metrics.AggregateServerMetrics(ms)

		log.Debug().
			Str("metric", "server").
			Int("minute", minute).
			Msg("sending server metrics")

		_, err := s.api.SendServerMetrics(metric.String())
		if err != nil {
			log.Warn().Msg("Failed to send server metrics: " + err.Error())
			break
		}

		log.Debug().
			Str("metric", "server").
			Int("minute", minute).
			Msg("removing minute from bucket")

		s.serverMetricBucket.Remove(minute)

		// stop sending if the bucket contains only one (the current) minute
		if len(s.serverMetricBucket.Minutes()) <= 1 {
			break
		}
	}
}

func (s *Sender) sendAppMetrics() {
	count := s.appMetricBucket.Count()
	bucket := s.appMetricBucket.Extract(s.config.AppBucketLimit)

	log.Debug().
		Int("total-size", count).
		Int("chunk-size", bucket.Count()).
		Str("metric", "app").
		Msg("sending app metrics")

	_, err := s.api.SendEnvironmentMetrics(bucket.String())
	if err != nil {
		log.Error().Msg("failed to send app metrics: " + err.Error())
		log.Debug().Msg("sleeping before adding back the metrics to the app bucket")
		time.Sleep(5 * time.Second)

		s.appMetricBucket.Merge(bucket)
		log.Debug().
			Int("count", s.appMetricBucket.Count()).
			Msg("merged app metric buckets")
		return
	}

	s.updateSentAt()
}
