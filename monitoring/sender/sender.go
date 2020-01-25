package sender

import (
	"log"
	"sync"
	"time"

	"github.com/larashed/agent-go/api"
	"github.com/larashed/agent-go/monitoring/buckets"
)

type Sender struct {
	api api.Api

	appMetricBucket    *buckets.Bucket
	serverMetricBucket *buckets.Bucket

	config *Config

	sentAt time.Time
	mutex  sync.RWMutex

	stopServerMetricSend chan int
	stopAppMetricSend    chan int
}

func NewSender(
	api api.Api,
	appMetricBucket *buckets.Bucket,
	serverMetricBucket *buckets.Bucket,
	config *Config) *Sender {
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
			case <-s.appMetricBucket.Channel:
				if s.appMetricBucket.Count() >= s.config.appBucketLimit {
					s.sendAppMetrics()
				}
			case <-s.stopServerMetricSend:
				return
			}
		}
		// collects server metrics every n seconds
		// aggregates collected metrics and sends as one record
		// @TODO implement
	}()
}

func (s *Sender) SendAppMetrics() {
	go func() {
		for {
			select {
			case <-s.appMetricBucket.Channel:
				if s.appMetricBucket.Count() >= s.config.appBucketLimit {
					s.sendAppMetrics()
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
				// send data if the bucket is not empty and there hasn't been a send in 5 seconds
				if t.Sub(s.sentAt).Seconds() > 5 {
					if s.appMetricBucket.Count() > 0 {
						s.sendAppMetrics()
					}
				}
			}
		}
	}()
}

func (s *Sender) updateSentAt() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sentAt = time.Now()
}

func (s *Sender) sendAppMetrics() {
	bucket := s.appMetricBucket.PullAndRemove(s.config.appBucketLimit)

	err := s.api.SendApplicationRecords(bucket.String())
	if err != nil {
		log.Println("Failed to send bucket.", err.Error())
		s.appMetricBucket.Merge(bucket)
	}

	s.updateSentAt()
}
