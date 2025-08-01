// Code generated by mockery v2.53.3. DO NOT EDIT.

package mock

import (
	flow "github.com/onflow/flow-go/model/flow"
	mock "github.com/stretchr/testify/mock"
)

// SnapshotExecutionSubset is an autogenerated mock type for the SnapshotExecutionSubset type
type SnapshotExecutionSubset struct {
	mock.Mock
}

// RandomSource provides a mock function with no fields
func (_m *SnapshotExecutionSubset) RandomSource() ([]byte, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RandomSource")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]byte, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// VersionBeacon provides a mock function with no fields
func (_m *SnapshotExecutionSubset) VersionBeacon() (*flow.SealedVersionBeacon, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for VersionBeacon")
	}

	var r0 *flow.SealedVersionBeacon
	var r1 error
	if rf, ok := ret.Get(0).(func() (*flow.SealedVersionBeacon, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *flow.SealedVersionBeacon); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.SealedVersionBeacon)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewSnapshotExecutionSubset creates a new instance of SnapshotExecutionSubset. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSnapshotExecutionSubset(t interface {
	mock.TestingT
	Cleanup(func())
}) *SnapshotExecutionSubset {
	mock := &SnapshotExecutionSubset{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
