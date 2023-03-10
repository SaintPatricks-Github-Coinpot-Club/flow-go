// Code generated by mockery v2.21.4. DO NOT EDIT.

package mockp2p

import (
	p2p "github.com/onflow/flow-go/network/p2p"
	mock "github.com/stretchr/testify/mock"
)

// RateLimiterOpt is an autogenerated mock type for the RateLimiterOpt type
type RateLimiterOpt struct {
	mock.Mock
}

// Execute provides a mock function with given fields: limiter
func (_m *RateLimiterOpt) Execute(limiter p2p.RateLimiter) {
	_m.Called(limiter)
}

type mockConstructorTestingTNewRateLimiterOpt interface {
	mock.TestingT
	Cleanup(func())
}

// NewRateLimiterOpt creates a new instance of RateLimiterOpt. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRateLimiterOpt(t mockConstructorTestingTNewRateLimiterOpt) *RateLimiterOpt {
	mock := &RateLimiterOpt{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
