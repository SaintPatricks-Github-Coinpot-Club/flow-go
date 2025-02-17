// Code generated by mockery v2.13.1. DO NOT EDIT.

package mock

import (
	common "github.com/onflow/cadence/runtime/common"

	mock "github.com/stretchr/testify/mock"
)

// AccountFreezer is an autogenerated mock type for the AccountFreezer type
type AccountFreezer struct {
	mock.Mock
}

// SetAccountFrozen provides a mock function with given fields: _a0, _a1
func (_m *AccountFreezer) SetAccountFrozen(_a0 common.Address, _a1 bool) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(common.Address, bool) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewAccountFreezer interface {
	mock.TestingT
	Cleanup(func())
}

// NewAccountFreezer creates a new instance of AccountFreezer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAccountFreezer(t mockConstructorTestingTNewAccountFreezer) *AccountFreezer {
	mock := &AccountFreezer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
