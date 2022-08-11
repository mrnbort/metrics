// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package api

import (
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
// 			DeleteFunc: func(m metric.Entry) error {
// 				panic("mock out the Delete method")
// 			},
// 			GetFunc: func(from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
// 				panic("mock out the Get method")
// 			},
// 			UpdateFunc: func(m metric.Entry) error {
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
	DeleteFunc func(m metric.Entry) error

	// GetFunc mocks the Get method.
	GetFunc func(from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error)

	// UpdateFunc mocks the Update method.
	UpdateFunc func(m metric.Entry) error

	// calls tracks calls to the methods.
	calls struct {
		// Delete holds details about calls to the Delete method.
		Delete []struct {
			// M is the m argument value.
			M metric.Entry
		}
		// Get holds details about calls to the Get method.
		Get []struct {
			// From is the from argument value.
			From time.Time
			// To is the to argument value.
			To time.Time
			// Interval is the interval argument value.
			Interval time.Duration
		}
		// Update holds details about calls to the Update method.
		Update []struct {
			// M is the m argument value.
			M metric.Entry
		}
	}
	lockDelete sync.RWMutex
	lockGet    sync.RWMutex
	lockUpdate sync.RWMutex
}

// Delete calls DeleteFunc.
func (mock *StorageMock) Delete(m metric.Entry) error {
	if mock.DeleteFunc == nil {
		panic("StorageMock.DeleteFunc: method is nil but Storage.Delete was just called")
	}
	callInfo := struct {
		M metric.Entry
	}{
		M: m,
	}
	mock.lockDelete.Lock()
	mock.calls.Delete = append(mock.calls.Delete, callInfo)
	mock.lockDelete.Unlock()
	return mock.DeleteFunc(m)
}

// DeleteCalls gets all the calls that were made to Delete.
// Check the length with:
//     len(mockedStorage.DeleteCalls())
func (mock *StorageMock) DeleteCalls() []struct {
	M metric.Entry
} {
	var calls []struct {
		M metric.Entry
	}
	mock.lockDelete.RLock()
	calls = mock.calls.Delete
	mock.lockDelete.RUnlock()
	return calls
}

// Get calls GetFunc.
func (mock *StorageMock) Get(from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	if mock.GetFunc == nil {
		panic("StorageMock.GetFunc: method is nil but Storage.Get was just called")
	}
	callInfo := struct {
		From     time.Time
		To       time.Time
		Interval time.Duration
	}{
		From:     from,
		To:       to,
		Interval: interval,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(from, to, interval)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//     len(mockedStorage.GetCalls())
func (mock *StorageMock) GetCalls() []struct {
	From     time.Time
	To       time.Time
	Interval time.Duration
} {
	var calls []struct {
		From     time.Time
		To       time.Time
		Interval time.Duration
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// Update calls UpdateFunc.
func (mock *StorageMock) Update(m metric.Entry) error {
	if mock.UpdateFunc == nil {
		panic("StorageMock.UpdateFunc: method is nil but Storage.Update was just called")
	}
	callInfo := struct {
		M metric.Entry
	}{
		M: m,
	}
	mock.lockUpdate.Lock()
	mock.calls.Update = append(mock.calls.Update, callInfo)
	mock.lockUpdate.Unlock()
	return mock.UpdateFunc(m)
}

// UpdateCalls gets all the calls that were made to Update.
// Check the length with:
//     len(mockedStorage.UpdateCalls())
func (mock *StorageMock) UpdateCalls() []struct {
	M metric.Entry
} {
	var calls []struct {
		M metric.Entry
	}
	mock.lockUpdate.RLock()
	calls = mock.calls.Update
	mock.lockUpdate.RUnlock()
	return calls
}
