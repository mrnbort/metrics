// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package storage

import (
	"github.com/umputun/metrics/metric"
	"sync"
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
// 			DeleteFunc: func(m metric.Entry) error {
// 				panic("mock out the Delete method")
// 			},
// 			WriteFunc: func(m metric.Entry) error {
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
	DeleteFunc func(m metric.Entry) error

	// WriteFunc mocks the Write method.
	WriteFunc func(m metric.Entry) error

	// calls tracks calls to the methods.
	calls struct {
		// Delete holds details about calls to the Delete method.
		Delete []struct {
			// M is the m argument value.
			M metric.Entry
		}
		// Write holds details about calls to the Write method.
		Write []struct {
			// M is the m argument value.
			M metric.Entry
		}
	}
	lockDelete sync.RWMutex
	lockWrite  sync.RWMutex
}

// Delete calls DeleteFunc.
func (mock *AccessorMock) Delete(m metric.Entry) error {
	if mock.DeleteFunc == nil {
		panic("AccessorMock.DeleteFunc: method is nil but Accessor.Delete was just called")
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
//     len(mockedAccessor.DeleteCalls())
func (mock *AccessorMock) DeleteCalls() []struct {
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

// Write calls WriteFunc.
func (mock *AccessorMock) Write(m metric.Entry) error {
	if mock.WriteFunc == nil {
		panic("AccessorMock.WriteFunc: method is nil but Accessor.Write was just called")
	}
	callInfo := struct {
		M metric.Entry
	}{
		M: m,
	}
	mock.lockWrite.Lock()
	mock.calls.Write = append(mock.calls.Write, callInfo)
	mock.lockWrite.Unlock()
	return mock.WriteFunc(m)
}

// WriteCalls gets all the calls that were made to Write.
// Check the length with:
//     len(mockedAccessor.WriteCalls())
func (mock *AccessorMock) WriteCalls() []struct {
	M metric.Entry
} {
	var calls []struct {
		M metric.Entry
	}
	mock.lockWrite.RLock()
	calls = mock.calls.Write
	mock.lockWrite.RUnlock()
	return calls
}
