package submitter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/j-vizcaino/datadog-smartctl/metric"
)

type statsdGauge interface {
	Gauge(string, float64, []string, float64) error
}

type errorHandler func(err error)

type Submitter struct {
	metricUpdates chan metric.DeviceMetrics
	metricStore   []metric.DeviceMetrics

	statsdClient statsdGauge
	errorHandler errorHandler

	stop    chan bool
	running sync.WaitGroup
}

func New(statsdClient statsdGauge, errorHandler errorHandler) *Submitter {

	return &Submitter{
		metricUpdates: make(chan metric.DeviceMetrics, 16),
		stop:          make(chan bool),
		statsdClient:  statsdClient,
		errorHandler:  errorHandler,
	}
}

func (s *Submitter) Run(period time.Duration) {
	s.running.Add(1)
	go s.periodicSubmit(period)
}

func (s *Submitter) Stop() {
	close(s.stop)
	s.running.Wait()
}

func (s *Submitter) Update(ctx context.Context, metrics metric.DeviceMetrics) {
	select {
	case s.metricUpdates <- metrics:
	case <-ctx.Done():
	}
}

func (s *Submitter) periodicSubmit(period time.Duration) {
	run := true
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for run {
		select {
		case <-ticker.C:
			s.submitMetrics()

		case update := <-s.metricUpdates:
			s.saveMetrics(update)

		case <-s.stop:
			run = false
		}
	}
	s.running.Done()
}

func (s *Submitter) saveMetrics(updated metric.DeviceMetrics) {
	for idx, existing := range s.metricStore {
		if updated.DeviceName == existing.DeviceName {
			s.metricStore[idx] = updated
			return
		}
	}

	s.metricStore = append(s.metricStore, updated)
}

func (s *Submitter) submitMetrics() {
	var sampleErr error
	errCount := 0
	for _, device := range s.metricStore {
		for _, metricData := range device.Entries {
			err := s.statsdClient.Gauge(
				metricData.Name,
				float64(metricData.Value),
				device.CommonTags,
				1.0,
			)
			if err != nil {
				errCount++
				if sampleErr == nil {
					sampleErr = err
				}
			}
		}
	}

	if sampleErr != nil {
		err := fmt.Errorf(
			"statsd submission failed %d times, sample error: %w",
			errCount,
			sampleErr,
		)
		s.errorHandler(err)
	}
}
