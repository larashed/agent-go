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
	stop   chan int
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
	}
}

func (s *Sender) Start() {
	// sends server metrics in a go routine indefinitely
	s.sendServerMetrics()

	// send server metrics via 2 go routines indefinitely
	s.sendApplicationMetrics()
}

func (s *Sender) Stop() {
	s.stop <- 1
	s.stop <- 1
}

func (s *Sender) sendServerMetrics() {
	go func() {
		// collects server metrics every n seconds
		// aggregates collected metrics and sends as one record
		// @TODO implement
	}()
}

func (s *Sender) sendApplicationMetrics() {
	go func() {
		for {
			select {
			case <-s.appMetricBucket.Channel:
				if s.appMetricBucket.Count() >= s.config.appBucketLimit {
					s.sendAppMetrics()
				}
			case <-s.stop:
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-s.stop:
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
