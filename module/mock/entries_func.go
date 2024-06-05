// Code generated by mockery v2.43.2. DO NOT EDIT.

package mock

import mock "github.com/stretchr/testify/mock"

// EntriesFunc is an autogenerated mock type for the EntriesFunc type
type EntriesFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields:
func (_m *EntriesFunc) Execute() uint {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 uint
	if rf, ok := ret.Get(0).(func() uint); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint)
	}

	return r0
}

// NewEntriesFunc creates a new instance of EntriesFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEntriesFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *EntriesFunc {
	mock := &EntriesFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
