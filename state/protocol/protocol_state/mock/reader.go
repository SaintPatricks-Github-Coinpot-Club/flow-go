// Code generated by mockery v2.21.4. DO NOT EDIT.

package mock

import (
	protocol_state "github.com/onflow/flow-go/state/protocol/protocol_state"
	mock "github.com/stretchr/testify/mock"
)

// Reader is an autogenerated mock type for the Reader type
type Reader struct {
	mock.Mock
}

// GetInvalidEpochTransitionAttempted provides a mock function with given fields:
func (_m *Reader) GetInvalidEpochTransitionAttempted() (bool, error) {
	ret := _m.Called()

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func() (bool, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetProtocolStateVersion provides a mock function with given fields:
func (_m *Reader) GetProtocolStateVersion() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetVersionUpgrade provides a mock function with given fields:
func (_m *Reader) GetVersionUpgrade() *protocol_state.ViewBasedActivator[uint64] {
	ret := _m.Called()

	var r0 *protocol_state.ViewBasedActivator[uint64]
	if rf, ok := ret.Get(0).(func() *protocol_state.ViewBasedActivator[uint64]); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*protocol_state.ViewBasedActivator[uint64])
		}
	}

	return r0
}

// VersionedEncode provides a mock function with given fields:
func (_m *Reader) VersionedEncode() (uint64, []byte, error) {
	ret := _m.Called()

	var r0 uint64
	var r1 []byte
	var r2 error
	if rf, ok := ret.Get(0).(func() (uint64, []byte, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func() []byte); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]byte)
		}
	}

	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

type mockConstructorTestingTNewReader interface {
	mock.TestingT
	Cleanup(func())
}

// NewReader creates a new instance of Reader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewReader(t mockConstructorTestingTNewReader) *Reader {
	mock := &Reader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
