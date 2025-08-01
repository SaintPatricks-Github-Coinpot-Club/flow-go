// Code generated by mockery v2.53.3. DO NOT EDIT.

package mock

import (
	context "context"

	execution "github.com/onflow/flow/protobuf/go/flow/execution"

	flow "github.com/onflow/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"
)

// TxErrorMessageProvider is an autogenerated mock type for the TxErrorMessageProvider type
type TxErrorMessageProvider struct {
	mock.Mock
}

// ErrorMessageByBlockIDFromAnyEN provides a mock function with given fields: ctx, execNodes, req
func (_m *TxErrorMessageProvider) ErrorMessageByBlockIDFromAnyEN(ctx context.Context, execNodes flow.GenericIdentityList[flow.IdentitySkeleton], req *execution.GetTransactionErrorMessagesByBlockIDRequest) ([]*execution.GetTransactionErrorMessagesResponse_Result, *flow.IdentitySkeleton, error) {
	ret := _m.Called(ctx, execNodes, req)

	if len(ret) == 0 {
		panic("no return value specified for ErrorMessageByBlockIDFromAnyEN")
	}

	var r0 []*execution.GetTransactionErrorMessagesResponse_Result
	var r1 *flow.IdentitySkeleton
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessagesByBlockIDRequest) ([]*execution.GetTransactionErrorMessagesResponse_Result, *flow.IdentitySkeleton, error)); ok {
		return rf(ctx, execNodes, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessagesByBlockIDRequest) []*execution.GetTransactionErrorMessagesResponse_Result); ok {
		r0 = rf(ctx, execNodes, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*execution.GetTransactionErrorMessagesResponse_Result)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessagesByBlockIDRequest) *flow.IdentitySkeleton); ok {
		r1 = rf(ctx, execNodes, req)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*flow.IdentitySkeleton)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessagesByBlockIDRequest) error); ok {
		r2 = rf(ctx, execNodes, req)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ErrorMessageByIndex provides a mock function with given fields: ctx, blockID, height, index
func (_m *TxErrorMessageProvider) ErrorMessageByIndex(ctx context.Context, blockID flow.Identifier, height uint64, index uint32) (string, error) {
	ret := _m.Called(ctx, blockID, height, index)

	if len(ret) == 0 {
		panic("no return value specified for ErrorMessageByIndex")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, flow.Identifier, uint64, uint32) (string, error)); ok {
		return rf(ctx, blockID, height, index)
	}
	if rf, ok := ret.Get(0).(func(context.Context, flow.Identifier, uint64, uint32) string); ok {
		r0 = rf(ctx, blockID, height, index)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, flow.Identifier, uint64, uint32) error); ok {
		r1 = rf(ctx, blockID, height, index)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ErrorMessageByIndexFromAnyEN provides a mock function with given fields: ctx, execNodes, req
func (_m *TxErrorMessageProvider) ErrorMessageByIndexFromAnyEN(ctx context.Context, execNodes flow.GenericIdentityList[flow.IdentitySkeleton], req *execution.GetTransactionErrorMessageByIndexRequest) (*execution.GetTransactionErrorMessageResponse, error) {
	ret := _m.Called(ctx, execNodes, req)

	if len(ret) == 0 {
		panic("no return value specified for ErrorMessageByIndexFromAnyEN")
	}

	var r0 *execution.GetTransactionErrorMessageResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessageByIndexRequest) (*execution.GetTransactionErrorMessageResponse, error)); ok {
		return rf(ctx, execNodes, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessageByIndexRequest) *execution.GetTransactionErrorMessageResponse); ok {
		r0 = rf(ctx, execNodes, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*execution.GetTransactionErrorMessageResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessageByIndexRequest) error); ok {
		r1 = rf(ctx, execNodes, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ErrorMessageByTransactionID provides a mock function with given fields: ctx, blockID, height, transactionID
func (_m *TxErrorMessageProvider) ErrorMessageByTransactionID(ctx context.Context, blockID flow.Identifier, height uint64, transactionID flow.Identifier) (string, error) {
	ret := _m.Called(ctx, blockID, height, transactionID)

	if len(ret) == 0 {
		panic("no return value specified for ErrorMessageByTransactionID")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, flow.Identifier, uint64, flow.Identifier) (string, error)); ok {
		return rf(ctx, blockID, height, transactionID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, flow.Identifier, uint64, flow.Identifier) string); ok {
		r0 = rf(ctx, blockID, height, transactionID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, flow.Identifier, uint64, flow.Identifier) error); ok {
		r1 = rf(ctx, blockID, height, transactionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ErrorMessageFromAnyEN provides a mock function with given fields: ctx, execNodes, req
func (_m *TxErrorMessageProvider) ErrorMessageFromAnyEN(ctx context.Context, execNodes flow.GenericIdentityList[flow.IdentitySkeleton], req *execution.GetTransactionErrorMessageRequest) (*execution.GetTransactionErrorMessageResponse, error) {
	ret := _m.Called(ctx, execNodes, req)

	if len(ret) == 0 {
		panic("no return value specified for ErrorMessageFromAnyEN")
	}

	var r0 *execution.GetTransactionErrorMessageResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessageRequest) (*execution.GetTransactionErrorMessageResponse, error)); ok {
		return rf(ctx, execNodes, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessageRequest) *execution.GetTransactionErrorMessageResponse); ok {
		r0 = rf(ctx, execNodes, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*execution.GetTransactionErrorMessageResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, flow.GenericIdentityList[flow.IdentitySkeleton], *execution.GetTransactionErrorMessageRequest) error); ok {
		r1 = rf(ctx, execNodes, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ErrorMessagesByBlockID provides a mock function with given fields: ctx, blockID, height
func (_m *TxErrorMessageProvider) ErrorMessagesByBlockID(ctx context.Context, blockID flow.Identifier, height uint64) (map[flow.Identifier]string, error) {
	ret := _m.Called(ctx, blockID, height)

	if len(ret) == 0 {
		panic("no return value specified for ErrorMessagesByBlockID")
	}

	var r0 map[flow.Identifier]string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, flow.Identifier, uint64) (map[flow.Identifier]string, error)); ok {
		return rf(ctx, blockID, height)
	}
	if rf, ok := ret.Get(0).(func(context.Context, flow.Identifier, uint64) map[flow.Identifier]string); ok {
		r0 = rf(ctx, blockID, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[flow.Identifier]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, flow.Identifier, uint64) error); ok {
		r1 = rf(ctx, blockID, height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewTxErrorMessageProvider creates a new instance of TxErrorMessageProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTxErrorMessageProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *TxErrorMessageProvider {
	mock := &TxErrorMessageProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
