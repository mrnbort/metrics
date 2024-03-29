// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package storage

import (
	"context"
	"github.com/umputun/metrics/metric"
	"sync"
	"time"
)

// Ensure, that AccessorMock does implement Accessor.
// If this is not the case, regenerate this file with moq.
var _ Accessor = &AccessorMock{}

// AccessorMock is a mock implementation of Accessor.
//
// 	func TestSomethingThatUsesAccessor(t *testing.T) {
//
// 		// make and configure a mocked Accessor
// 		mockedAccessor := &AccessorMock{
// 			DeleteFunc: func(ctx context.Context, m metric.Entry) error {
// 				panic("mock out the Delete method")
// 			},
// 			FindAllFunc: func(ctx context.Context, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
// 				panic("mock out the FindAll method")
// 			},
// 			FindOneMetricFunc: func(ctx context.Context, name string, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
// 				panic("mock out the FindOneMetric method")
// 			},
// 			GetMetricsListFunc: func(ctx context.Context) ([]string, error) {
// 				panic("mock out the GetMetricsList method")
// 			},
// 			WriteFunc: func(ctx context.Context, m metric.Entry) error {
// 				panic("mock out the Write method")
// 			},
// 		}
//
// 		// use mockedAccessor in code that requires Accessor
// 		// and then make assertions.
//
// 	}
type AccessorMock struct {
	// DeleteFunc mocks the Delete method.
	DeleteFunc func(ctx context.Context, m metric.Entry) error

	// FindAllFunc mocks the FindAll method.
	FindAllFunc func(ctx context.Context, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error)

	// FindOneMetricFunc mocks the FindOneMetric method.
	FindOneMetricFunc func(ctx context.Context, name string, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error)

	// GetMetricsListFunc mocks the GetMetricsList method.
	GetMetricsListFunc func(ctx context.Context) ([]string, error)

	// WriteFunc mocks the Write method.
	WriteFunc func(ctx context.Context, m metric.Entry) error

	// calls tracks calls to the methods.
	calls struct {
		// Delete holds details about calls to the Delete method.
		Delete []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// M is the m argument value.
			M metric.Entry
		}
		// FindAll holds details about calls to the FindAll method.
		FindAll []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// From is the from argument value.
			From time.Time
			// To is the to argument value.
			To time.Time
			// Interval is the interval argument value.
			Interval time.Duration
		}
		// FindOneMetric holds details about calls to the FindOneMetric method.
		FindOneMetric []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Name is the name argument value.
			Name string
			// From is the from argument value.
			From time.Time
			// To is the to argument value.
			To time.Time
			// Interval is the interval argument value.
			Interval time.Duration
		}
		// GetMetricsList holds details about calls to the GetMetricsList method.
		GetMetricsList []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Write holds details about calls to the Write method.
		Write []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// M is the m argument value.
			M metric.Entry
		}
	}
	lockDelete         sync.RWMutex
	lockFindAll        sync.RWMutex
	lockFindOneMetric  sync.RWMutex
	lockGetMetricsList sync.RWMutex
	lockWrite          sync.RWMutex
}

// Delete calls DeleteFunc.
func (mock *AccessorMock) Delete(ctx context.Context, m metric.Entry) error {
	if mock.DeleteFunc == nil {
		panic("AccessorMock.DeleteFunc: method is nil but Accessor.Delete was just called")
	}
	callInfo := struct {
		Ctx context.Context
		M   metric.Entry
	}{
		Ctx: ctx,
		M:   m,
	}
	mock.lockDelete.Lock()
	mock.calls.Delete = append(mock.calls.Delete, callInfo)
	mock.lockDelete.Unlock()
	return mock.DeleteFunc(ctx, m)
}

// DeleteCalls gets all the calls that were made to Delete.
// Check the length with:
//     len(mockedAccessor.DeleteCalls())
func (mock *AccessorMock) DeleteCalls() []struct {
	Ctx context.Context
	M   metric.Entry
} {
	var calls []struct {
		Ctx context.Context
		M   metric.Entry
	}
	mock.lockDelete.RLock()
	calls = mock.calls.Delete
	mock.lockDelete.RUnlock()
	return calls
}

// FindAll calls FindAllFunc.
func (mock *AccessorMock) FindAll(ctx context.Context, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	if mock.FindAllFunc == nil {
		panic("AccessorMock.FindAllFunc: method is nil but Accessor.FindAll was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		From     time.Time
		To       time.Time
		Interval time.Duration
	}{
		Ctx:      ctx,
		From:     from,
		To:       to,
		Interval: interval,
	}
	mock.lockFindAll.Lock()
	mock.calls.FindAll = append(mock.calls.FindAll, callInfo)
	mock.lockFindAll.Unlock()
	return mock.FindAllFunc(ctx, from, to, interval)
}

// FindAllCalls gets all the calls that were made to FindAll.
// Check the length with:
//     len(mockedAccessor.FindAllCalls())
func (mock *AccessorMock) FindAllCalls() []struct {
	Ctx      context.Context
	From     time.Time
	To       time.Time
	Interval time.Duration
} {
	var calls []struct {
		Ctx      context.Context
		From     time.Time
		To       time.Time
		Interval time.Duration
	}
	mock.lockFindAll.RLock()
	calls = mock.calls.FindAll
	mock.lockFindAll.RUnlock()
	return calls
}

// FindOneMetric calls FindOneMetricFunc.
func (mock *AccessorMock) FindOneMetric(ctx context.Context, name string, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	if mock.FindOneMetricFunc == nil {
		panic("AccessorMock.FindOneMetricFunc: method is nil but Accessor.FindOneMetric was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Name     string
		From     time.Time
		To       time.Time
		Interval time.Duration
	}{
		Ctx:      ctx,
		Name:     name,
		From:     from,
		To:       to,
		Interval: interval,
	}
	mock.lockFindOneMetric.Lock()
	mock.calls.FindOneMetric = append(mock.calls.FindOneMetric, callInfo)
	mock.lockFindOneMetric.Unlock()
	return mock.FindOneMetricFunc(ctx, name, from, to, interval)
}

// FindOneMetricCalls gets all the calls that were made to FindOneMetric.
// Check the length with:
//     len(mockedAccessor.FindOneMetricCalls())
func (mock *AccessorMock) FindOneMetricCalls() []struct {
	Ctx      context.Context
	Name     string
	From     time.Time
	To       time.Time
	Interval time.Duration
} {
	var calls []struct {
		Ctx      context.Context
		Name     string
		From     time.Time
		To       time.Time
		Interval time.Duration
	}
	mock.lockFindOneMetric.RLock()
	calls = mock.calls.FindOneMetric
	mock.lockFindOneMetric.RUnlock()
	return calls
}

// GetMetricsList calls GetMetricsListFunc.
func (mock *AccessorMock) GetMetricsList(ctx context.Context) ([]string, error) {
	if mock.GetMetricsListFunc == nil {
		panic("AccessorMock.GetMetricsListFunc: method is nil but Accessor.GetMetricsList was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockGetMetricsList.Lock()
	mock.calls.GetMetricsList = append(mock.calls.GetMetricsList, callInfo)
	mock.lockGetMetricsList.Unlock()
	return mock.GetMetricsListFunc(ctx)
}

// GetMetricsListCalls gets all the calls that were made to GetMetricsList.
// Check the length with:
//     len(mockedAccessor.GetMetricsListCalls())
func (mock *AccessorMock) GetMetricsListCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockGetMetricsList.RLock()
	calls = mock.calls.GetMetricsList
	mock.lockGetMetricsList.RUnlock()
	return calls
}

// Write calls WriteFunc.
func (mock *AccessorMock) Write(ctx context.Context, m metric.Entry) error {
	if mock.WriteFunc == nil {
		panic("AccessorMock.WriteFunc: method is nil but Accessor.Write was just called")
	}
	callInfo := struct {
		Ctx context.Context
		M   metric.Entry
	}{
		Ctx: ctx,
		M:   m,
	}
	mock.lockWrite.Lock()
	mock.calls.Write = append(mock.calls.Write, callInfo)
	mock.lockWrite.Unlock()
	return mock.WriteFunc(ctx, m)
}

// WriteCalls gets all the calls that were made to Write.
// Check the length with:
//     len(mockedAccessor.WriteCalls())
func (mock *AccessorMock) WriteCalls() []struct {
	Ctx context.Context
	M   metric.Entry
} {
	var calls []struct {
		Ctx context.Context
		M   metric.Entry
	}
	mock.lockWrite.RLock()
	calls = mock.calls.Write
	mock.lockWrite.RUnlock()
	return calls
}
