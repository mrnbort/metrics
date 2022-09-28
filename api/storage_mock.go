// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package api

import (
	"context"
	"github.com/umputun/metrics/metric"
	"sync"
	"time"
)

// Ensure, that StorageMock does implement Storage.
// If this is not the case, regenerate this file with moq.
var _ Storage = &StorageMock{}

// StorageMock is a mock implementation of Storage.
//
// 	func TestSomethingThatUsesStorage(t *testing.T) {
//
// 		// make and configure a mocked Storage
// 		mockedStorage := &StorageMock{
// 			DeleteFunc: func(ctx context.Context, m metric.Entry) error {
// 				panic("mock out the Delete method")
// 			},
// 			GetAllFunc: func(ctx context.Context, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
// 				panic("mock out the GetAll method")
// 			},
// 			UpdateFunc: func(ctx context.Context, m metric.Entry) error {
// 				panic("mock out the Update method")
// 			},
// 		}
//
// 		// use mockedStorage in code that requires Storage
// 		// and then make assertions.
//
// 	}
type StorageMock struct {
	// DeleteFunc mocks the Delete method.
	DeleteFunc func(ctx context.Context, m metric.Entry) error

	// GetAllFunc mocks the GetAll method.
	GetAllFunc func(ctx context.Context, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error)

	// UpdateFunc mocks the Update method.
	UpdateFunc func(ctx context.Context, m metric.Entry) error

	// calls tracks calls to the methods.
	calls struct {
		// Delete holds details about calls to the Delete method.
		Delete []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// M is the m argument value.
			M metric.Entry
		}
		// GetAll holds details about calls to the GetAll method.
		GetAll []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// From is the from argument value.
			From time.Time
			// To is the to argument value.
			To time.Time
			// Interval is the interval argument value.
			Interval time.Duration
		}
		// Update holds details about calls to the Update method.
		Update []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// M is the m argument value.
			M metric.Entry
		}
	}
	lockDelete sync.RWMutex
	lockGetAll sync.RWMutex
	lockUpdate sync.RWMutex
}

// Delete calls DeleteFunc.
func (mock *StorageMock) Delete(ctx context.Context, m metric.Entry) error {
	if mock.DeleteFunc == nil {
		panic("StorageMock.DeleteFunc: method is nil but Storage.Delete was just called")
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
//     len(mockedStorage.DeleteCalls())
func (mock *StorageMock) DeleteCalls() []struct {
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

// GetAll calls GetAllFunc.
func (mock *StorageMock) GetAll(ctx context.Context, from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	if mock.GetAllFunc == nil {
		panic("StorageMock.GetAllFunc: method is nil but Storage.GetAll was just called")
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
	mock.lockGetAll.Lock()
	mock.calls.GetAll = append(mock.calls.GetAll, callInfo)
	mock.lockGetAll.Unlock()
	return mock.GetAllFunc(ctx, from, to, interval)
}

// GetAllCalls gets all the calls that were made to GetAll.
// Check the length with:
//     len(mockedStorage.GetAllCalls())
func (mock *StorageMock) GetAllCalls() []struct {
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
	mock.lockGetAll.RLock()
	calls = mock.calls.GetAll
	mock.lockGetAll.RUnlock()
	return calls
}

// Update calls UpdateFunc.
func (mock *StorageMock) Update(ctx context.Context, m metric.Entry) error {
	if mock.UpdateFunc == nil {
		panic("StorageMock.UpdateFunc: method is nil but Storage.Update was just called")
	}
	callInfo := struct {
		Ctx context.Context
		M   metric.Entry
	}{
		Ctx: ctx,
		M:   m,
	}
	mock.lockUpdate.Lock()
	mock.calls.Update = append(mock.calls.Update, callInfo)
	mock.lockUpdate.Unlock()
	return mock.UpdateFunc(ctx, m)
}

// UpdateCalls gets all the calls that were made to Update.
// Check the length with:
//     len(mockedStorage.UpdateCalls())
func (mock *StorageMock) UpdateCalls() []struct {
	Ctx context.Context
	M   metric.Entry
} {
	var calls []struct {
		Ctx context.Context
		M   metric.Entry
	}
	mock.lockUpdate.RLock()
	calls = mock.calls.Update
	mock.lockUpdate.RUnlock()
	return calls
}
