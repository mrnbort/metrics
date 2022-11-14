package storage

import (
	"context"
	"fmt"
	"github.com/umputun/metrics/metric"
	"log"
	"sync"
	"time"
)

//go:generate moq -out accessor_mock.go . Accessor

// Service allows access to db and memory
type Service struct {
	db Accessor

	staging struct {
		sync.Mutex
		data map[string]metric.Entry
	}
}

// Accessor provides access to the db functions
type Accessor interface {
	Write(ctx context.Context, m metric.Entry) error
	Delete(ctx context.Context, m metric.Entry) error
	GetMetricsList(ctx context.Context) ([]string, error)
	FindOneMetric(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error)
	FindAll(ctx context.Context, from, to time.Time, interval time.Duration) ([]metric.Entry, error)
}

// New initiates and returns db and in-memory data
func New(db Accessor) *Service {
	result := &Service{
		db: db,
	}
	result.staging.data = make(map[string]metric.Entry)
	return result
}

// Update adds or updates a metric to the in-memory storage and
// calls Write to add the metric to the db
func (s *Service) Update(ctx context.Context, m metric.Entry) error {
	s.staging.Lock()
	defer s.staging.Unlock()

	v, ok := s.staging.data[m.Name]
	if !ok {
		// metric not found
		m.MinSinceMidnight = s.getMinSinceMidnight(m.TimeStamp)
		m.Type = 1 * time.Minute
		m.TypeStr = "1m"
		s.staging.data[m.Name] = m
		return nil
	}

	mins := s.getMinSinceMidnight(m.TimeStamp)
	if mins == v.MinSinceMidnight { // matched minute, update metric value
		v.Value += m.Value
		s.staging.data[m.Name] = v
		return nil
	}

	// new minute
	if err := s.db.Write(ctx, v); err != nil {
		return fmt.Errorf("failed to write metric %v: %w", m, err)
	}

	m.MinSinceMidnight = s.getMinSinceMidnight(m.TimeStamp)
	m.Type = 1 * time.Minute
	m.TypeStr = "1m"
	s.staging.data[m.Name] = m // set new metric to hash
	return nil
}

// Delete removes the metric from in-memory storage and db
func (s *Service) Delete(ctx context.Context, m metric.Entry) error {
	s.staging.Lock()

	_, ok := s.staging.data[m.Name]
	if ok {
		// metric found in data
		delete(s.staging.data, m.Name)
	}

	s.staging.Unlock()

	if err := s.db.Delete(ctx, m); err != nil {
		return fmt.Errorf("failed to delete metric %v: %w", m, err)
	}
	return nil
}

// GetList returns a list of all the available metrics in db
func (s *Service) GetList(ctx context.Context) ([]string, error) {
	metrics, err := s.db.GetMetricsList(ctx)
	if err != nil {
		return metrics, fmt.Errorf("failed to find all metrics: %w", err)
	}
	return metrics, nil
}

// GetOneMetric returns a list values for the requested metric during the requested interval
func (s *Service) GetOneMetric(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	metrics, err := s.db.FindOneMetric(ctx, name, from, to, interval)
	if err != nil {
		return metrics, fmt.Errorf("failed to find %v metric: %w", name, err)
	}
	return metrics, nil
}

// GetAll gets all entries for the specified timeframe and interval
func (s *Service) GetAll(ctx context.Context, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	metrics, err := s.db.FindAll(ctx, from, to, interval)
	if err != nil {
		return metrics, fmt.Errorf("failed to find metrics: %w", err)
	}
	return metrics, nil
}

// getMinSinceMidnight calculates the number of minutes since midnight
func (s *Service) getMinSinceMidnight(tm time.Time) int {
	return tm.Hour()*60 + tm.Minute()
}

// ActivateCleanup activates cleanup for the specified duration
func (s *Service) ActivateCleanup(ctx context.Context, duration time.Duration) {

	go func() {
		tick := time.NewTicker(duration)
		defer tick.Stop()

		for range tick.C {
			if err := s.doCleanup(ctx); err != nil {
				log.Printf("oh my, failed to clenaup, %v", err)
			}
		}

	}()
}

// doCleanup cleans up the in-memory data by moving entries to db
func (s *Service) doCleanup(ctx context.Context) error {
	s.staging.Lock()
	defer s.staging.Unlock()

	if len(s.staging.data) <= 0 {
		return nil
	}

	nowMins := s.getMinSinceMidnight(time.Now())

	for k, v := range s.staging.data {
		if nowMins == v.MinSinceMidnight {
			continue
		}

		if err := s.db.Write(ctx, v); err != nil {
			return fmt.Errorf("failed to add expired minute %v: %w", v, err)
		}
		delete(s.staging.data, k)
	}

	return nil
}
