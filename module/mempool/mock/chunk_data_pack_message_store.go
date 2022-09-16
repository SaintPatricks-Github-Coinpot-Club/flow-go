// Code generated by mockery v2.13.1. DO NOT EDIT.

package mempool

import (
	engine "github.com/onflow/flow-go/engine"

	mock "github.com/stretchr/testify/mock"
)

// ChunkDataPackMessageStore is an autogenerated mock type for the ChunkDataPackMessageStore type
type ChunkDataPackMessageStore struct {
	mock.Mock
}

// Get provides a mock function with given fields:
func (_m *ChunkDataPackMessageStore) Get() (*engine.Message, bool) {
	ret := _m.Called()

	var r0 *engine.Message
	if rf, ok := ret.Get(0).(func() *engine.Message); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*engine.Message)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func() bool); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// Put provides a mock function with given fields: _a0
func (_m *ChunkDataPackMessageStore) Put(_a0 *engine.Message) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*engine.Message) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Size provides a mock function with given fields:
func (_m *ChunkDataPackMessageStore) Size() uint {
	ret := _m.Called()

	var r0 uint
	if rf, ok := ret.Get(0).(func() uint); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint)
	}

	return r0
}

type mockConstructorTestingTNewChunkDataPackMessageStore interface {
	mock.TestingT
	Cleanup(func())
}

// NewChunkDataPackMessageStore creates a new instance of ChunkDataPackMessageStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewChunkDataPackMessageStore(t mockConstructorTestingTNewChunkDataPackMessageStore) *ChunkDataPackMessageStore {
	mock := &ChunkDataPackMessageStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
