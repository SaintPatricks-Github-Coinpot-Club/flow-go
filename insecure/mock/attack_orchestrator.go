// Code generated by mockery v2.13.1. DO NOT EDIT.

package mockinsecure

import (
	insecure "github.com/onflow/flow-go/insecure"
	mock "github.com/stretchr/testify/mock"
)

// AttackOrchestrator is an autogenerated mock type for the AttackOrchestrator type
type AttackOrchestrator struct {
	mock.Mock
}

// HandleEventFromCorruptedNode provides a mock function with given fields: _a0
func (_m *AttackOrchestrator) HandleEventFromCorruptedNode(_a0 *insecure.Event) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*insecure.Event) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WithAttackNetwork provides a mock function with given fields: _a0
func (_m *AttackOrchestrator) WithAttackNetwork(_a0 insecure.AttackNetwork) {
	_m.Called(_a0)
}

type mockConstructorTestingTNewAttackOrchestrator interface {
	mock.TestingT
	Cleanup(func())
}

// NewAttackOrchestrator creates a new instance of AttackOrchestrator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAttackOrchestrator(t mockConstructorTestingTNewAttackOrchestrator) *AttackOrchestrator {
	mock := &AttackOrchestrator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
