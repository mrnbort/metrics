package storage

import (
	"fmt"
	"github.com/umputun/metrics/metric"
	"log"
	"time"
)

//go:generate moq -out accessor_mock.go . Accessor

type Service struct {
	db   Accessor
	data map[string]metric.Entry
}

type Accessor interface {
	Write(m metric.Entry) error
	Delete(m metric.Entry) error
}

func New(db Accessor) *Service {
	result := &Service{
		data: make(map[string]metric.Entry),
		db:   db,
	}
	return result
}

func (s *Service) Update(m metric.Entry) error {

	v, ok := s.data[m.Name]
	if !ok {
		// metric not found
		m.MinSinceMidnight = s.getMinSinceMidnight(m.TimeStamp)
		s.data[m.Name] = m
		return nil
	}

	mins := s.getMinSinceMidnight(m.TimeStamp)
	if mins == v.MinSinceMidnight { // matched minute, update metric value
		v.Value += m.Value
		s.data[m.Name] = v
		return nil
	}

	// new minute
	if err := s.db.Write(v); err != nil {
		return fmt.Errorf("failed to write metric %v: %w", m, err)
	}

	m.MinSinceMidnight = s.getMinSinceMidnight(m.TimeStamp)
	s.data[m.Name] = m // set new metric to hash
	return nil
}

func (s *Service) Delete(m metric.Entry) error {
	_, ok := s.data[m.Name]
	if ok {
		// metric found in data
		delete(s.data, m.Name)
	}
	if err := s.db.Delete(m); err != nil {
		return fmt.Errorf("failed to delete metric %v: %w", m, err)
	}
	return nil
}

//func (s *Service) Find() (metric.Entry, error) {
//
//}

func (s *Service) getMinSinceMidnight(tm time.Time) int {
	return tm.Hour()*60 + tm.Minute()
}

func (s *Service) ActivateCleanup(duration time.Duration) {

	go func() {
		tick := time.NewTicker(duration)
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				if err := s.doCleanup(); err != nil {
					log.Printf("oh my, failed to clenaup, %v", err)
				}
			}
		}
	}()
}

func (s *Service) doCleanup() error {
	if len(s.data) <= 0 {
		return nil
	}

	nowMins := s.getMinSinceMidnight(time.Now())

	for k, v := range s.data {
		if nowMins == v.MinSinceMidnight {
			continue
		}

		if err := s.db.Write(v); err != nil {
			return fmt.Errorf("failed to add expired minute %v: %w", v, err)
		}
		delete(s.data, k)
	}

	return nil
}
