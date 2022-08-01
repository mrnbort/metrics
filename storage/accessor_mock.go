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
	// WriteFunc mocks the Write method.
	WriteFunc func(m metric.Entry) error

	// calls tracks calls to the methods.
	calls struct {
		// Write holds details about calls to the Write method.
		Write []struct {
			// M is the m argument value.
			M metric.Entry
		}
	}
	lockWrite sync.RWMutex
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
