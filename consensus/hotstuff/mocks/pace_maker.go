// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	flow "github.com/onflow/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"

	model "github.com/onflow/flow-go/consensus/hotstuff/model"

	time "time"
)

// PaceMaker is an autogenerated mock type for the PaceMaker type
type PaceMaker struct {
	mock.Mock
}

// CurView provides a mock function with given fields:
func (_m *PaceMaker) CurView() uint64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CurView")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// LastViewTC provides a mock function with given fields:
func (_m *PaceMaker) LastViewTC() *flow.TimeoutCertificate {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for LastViewTC")
	}

	var r0 *flow.TimeoutCertificate
	if rf, ok := ret.Get(0).(func() *flow.TimeoutCertificate); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.TimeoutCertificate)
		}
	}

	return r0
}

// NewestQC provides a mock function with given fields:
func (_m *PaceMaker) NewestQC() *flow.QuorumCertificate {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NewestQC")
	}

	var r0 *flow.QuorumCertificate
	if rf, ok := ret.Get(0).(func() *flow.QuorumCertificate); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.QuorumCertificate)
		}
	}

	return r0
}

// ProcessQC provides a mock function with given fields: qc
func (_m *PaceMaker) ProcessQC(qc *flow.QuorumCertificate) (*model.NewViewEvent, error) {
	ret := _m.Called(qc)

	if len(ret) == 0 {
		panic("no return value specified for ProcessQC")
	}

	var r0 *model.NewViewEvent
	var r1 error
	if rf, ok := ret.Get(0).(func(*flow.QuorumCertificate) (*model.NewViewEvent, error)); ok {
		return rf(qc)
	}
	if rf, ok := ret.Get(0).(func(*flow.QuorumCertificate) *model.NewViewEvent); ok {
		r0 = rf(qc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.NewViewEvent)
		}
	}

	if rf, ok := ret.Get(1).(func(*flow.QuorumCertificate) error); ok {
		r1 = rf(qc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProcessTC provides a mock function with given fields: tc
func (_m *PaceMaker) ProcessTC(tc *flow.TimeoutCertificate) (*model.NewViewEvent, error) {
	ret := _m.Called(tc)

	if len(ret) == 0 {
		panic("no return value specified for ProcessTC")
	}

	var r0 *model.NewViewEvent
	var r1 error
	if rf, ok := ret.Get(0).(func(*flow.TimeoutCertificate) (*model.NewViewEvent, error)); ok {
		return rf(tc)
	}
	if rf, ok := ret.Get(0).(func(*flow.TimeoutCertificate) *model.NewViewEvent); ok {
		r0 = rf(tc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.NewViewEvent)
		}
	}

	if rf, ok := ret.Get(1).(func(*flow.TimeoutCertificate) error); ok {
		r1 = rf(tc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Start provides a mock function with given fields: ctx
func (_m *PaceMaker) Start(ctx context.Context) {
	_m.Called(ctx)
}

// TargetPublicationTime provides a mock function with given fields: proposalView, timeViewEntered, parentBlockId
func (_m *PaceMaker) TargetPublicationTime(proposalView uint64, timeViewEntered time.Time, parentBlockId flow.Identifier) time.Time {
	ret := _m.Called(proposalView, timeViewEntered, parentBlockId)

	if len(ret) == 0 {
		panic("no return value specified for TargetPublicationTime")
	}

	var r0 time.Time
	if rf, ok := ret.Get(0).(func(uint64, time.Time, flow.Identifier) time.Time); ok {
		r0 = rf(proposalView, timeViewEntered, parentBlockId)
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// TimeoutChannel provides a mock function with given fields:
func (_m *PaceMaker) TimeoutChannel() <-chan time.Time {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for TimeoutChannel")
	}

	var r0 <-chan time.Time
	if rf, ok := ret.Get(0).(func() <-chan time.Time); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan time.Time)
		}
	}

	return r0
}

// NewPaceMaker creates a new instance of PaceMaker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPaceMaker(t interface {
	mock.TestingT
	Cleanup(func())
}) *PaceMaker {
	mock := &PaceMaker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
