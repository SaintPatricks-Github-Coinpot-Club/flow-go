// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/onflow/flow-go/storage (interfaces: Blocks,Headers,Payloads,Collections,Commits,Events,ServiceEvents,TransactionResults)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	flow "github.com/onflow/flow-go/model/flow"
	storage "github.com/onflow/flow-go/storage"
	transaction "github.com/onflow/flow-go/storage/badger/transaction"
)

// MockBlocks is a mock of Blocks interface.
type MockBlocks struct {
	ctrl     *gomock.Controller
	recorder *MockBlocksMockRecorder
}

// MockBlocksMockRecorder is the mock recorder for MockBlocks.
type MockBlocksMockRecorder struct {
	mock *MockBlocks
}

// NewMockBlocks creates a new mock instance.
func NewMockBlocks(ctrl *gomock.Controller) *MockBlocks {
	mock := &MockBlocks{ctrl: ctrl}
	mock.recorder = &MockBlocksMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBlocks) EXPECT() *MockBlocksMockRecorder {
	return m.recorder
}

// ByCollectionID mocks base method.
func (m *MockBlocks) ByCollectionID(arg0 flow.Identifier) (*flow.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByCollectionID", arg0)
	ret0, _ := ret[0].(*flow.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByCollectionID indicates an expected call of ByCollectionID.
func (mr *MockBlocksMockRecorder) ByCollectionID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByCollectionID", reflect.TypeOf((*MockBlocks)(nil).ByCollectionID), arg0)
}

// ByHeight mocks base method.
func (m *MockBlocks) ByHeight(arg0 uint64) (*flow.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByHeight", arg0)
	ret0, _ := ret[0].(*flow.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByHeight indicates an expected call of ByHeight.
func (mr *MockBlocksMockRecorder) ByHeight(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByHeight", reflect.TypeOf((*MockBlocks)(nil).ByHeight), arg0)
}

// ByID mocks base method.
func (m *MockBlocks) ByID(arg0 flow.Identifier) (*flow.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByID", arg0)
	ret0, _ := ret[0].(*flow.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByID indicates an expected call of ByID.
func (mr *MockBlocksMockRecorder) ByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByID", reflect.TypeOf((*MockBlocks)(nil).ByID), arg0)
}

// IndexBlockForCollections mocks base method.
func (m *MockBlocks) IndexBlockForCollections(arg0 flow.Identifier, arg1 []flow.Identifier) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IndexBlockForCollections", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// IndexBlockForCollections indicates an expected call of IndexBlockForCollections.
func (mr *MockBlocksMockRecorder) IndexBlockForCollections(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IndexBlockForCollections", reflect.TypeOf((*MockBlocks)(nil).IndexBlockForCollections), arg0, arg1)
}

// Store mocks base method.
func (m *MockBlocks) Store(arg0 *flow.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockBlocksMockRecorder) Store(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockBlocks)(nil).Store), arg0)
}

// StoreTx mocks base method.
func (m *MockBlocks) StoreTx(arg0 *flow.Block) func(*transaction.Tx) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreTx", arg0)
	ret0, _ := ret[0].(func(*transaction.Tx) error)
	return ret0
}

// StoreTx indicates an expected call of StoreTx.
func (mr *MockBlocksMockRecorder) StoreTx(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreTx", reflect.TypeOf((*MockBlocks)(nil).StoreTx), arg0)
}

// MockHeaders is a mock of Headers interface.
type MockHeaders struct {
	ctrl     *gomock.Controller
	recorder *MockHeadersMockRecorder
}

// MockHeadersMockRecorder is the mock recorder for MockHeaders.
type MockHeadersMockRecorder struct {
	mock *MockHeaders
}

// NewMockHeaders creates a new mock instance.
func NewMockHeaders(ctrl *gomock.Controller) *MockHeaders {
	mock := &MockHeaders{ctrl: ctrl}
	mock.recorder = &MockHeadersMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHeaders) EXPECT() *MockHeadersMockRecorder {
	return m.recorder
}

// BlockIDByHeight mocks base method.
func (m *MockHeaders) BlockIDByHeight(arg0 uint64) (flow.Identifier, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BlockIDByHeight", arg0)
	ret0, _ := ret[0].(flow.Identifier)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BlockIDByHeight indicates an expected call of BlockIDByHeight.
func (mr *MockHeadersMockRecorder) BlockIDByHeight(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockIDByHeight", reflect.TypeOf((*MockHeaders)(nil).BlockIDByHeight), arg0)
}

// ByBlockID mocks base method.
func (m *MockHeaders) ByBlockID(arg0 flow.Identifier) (*flow.Header, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockID", arg0)
	ret0, _ := ret[0].(*flow.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockID indicates an expected call of ByBlockID.
func (mr *MockHeadersMockRecorder) ByBlockID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockID", reflect.TypeOf((*MockHeaders)(nil).ByBlockID), arg0)
}

// ByHeight mocks base method.
func (m *MockHeaders) ByHeight(arg0 uint64) (*flow.Header, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByHeight", arg0)
	ret0, _ := ret[0].(*flow.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByHeight indicates an expected call of ByHeight.
func (mr *MockHeadersMockRecorder) ByHeight(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByHeight", reflect.TypeOf((*MockHeaders)(nil).ByHeight), arg0)
}

// ByParentID mocks base method.
func (m *MockHeaders) ByParentID(arg0 flow.Identifier) ([]*flow.Header, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByParentID", arg0)
	ret0, _ := ret[0].([]*flow.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByParentID indicates an expected call of ByParentID.
func (mr *MockHeadersMockRecorder) ByParentID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByParentID", reflect.TypeOf((*MockHeaders)(nil).ByParentID), arg0)
}

// Exists mocks base method.
func (m *MockHeaders) Exists(arg0 flow.Identifier) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists.
func (mr *MockHeadersMockRecorder) Exists(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockHeaders)(nil).Exists), arg0)
}

// Store mocks base method.
func (m *MockHeaders) Store(arg0 *flow.Header) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockHeadersMockRecorder) Store(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockHeaders)(nil).Store), arg0)
}

// MockPayloads is a mock of Payloads interface.
type MockPayloads struct {
	ctrl     *gomock.Controller
	recorder *MockPayloadsMockRecorder
}

// MockPayloadsMockRecorder is the mock recorder for MockPayloads.
type MockPayloadsMockRecorder struct {
	mock *MockPayloads
}

// NewMockPayloads creates a new mock instance.
func NewMockPayloads(ctrl *gomock.Controller) *MockPayloads {
	mock := &MockPayloads{ctrl: ctrl}
	mock.recorder = &MockPayloadsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPayloads) EXPECT() *MockPayloadsMockRecorder {
	return m.recorder
}

// ByBlockID mocks base method.
func (m *MockPayloads) ByBlockID(arg0 flow.Identifier) (*flow.Payload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockID", arg0)
	ret0, _ := ret[0].(*flow.Payload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockID indicates an expected call of ByBlockID.
func (mr *MockPayloadsMockRecorder) ByBlockID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockID", reflect.TypeOf((*MockPayloads)(nil).ByBlockID), arg0)
}

// Store mocks base method.
func (m *MockPayloads) Store(arg0 flow.Identifier, arg1 *flow.Payload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockPayloadsMockRecorder) Store(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockPayloads)(nil).Store), arg0, arg1)
}

// MockCollections is a mock of Collections interface.
type MockCollections struct {
	ctrl     *gomock.Controller
	recorder *MockCollectionsMockRecorder
}

// MockCollectionsMockRecorder is the mock recorder for MockCollections.
type MockCollectionsMockRecorder struct {
	mock *MockCollections
}

// NewMockCollections creates a new mock instance.
func NewMockCollections(ctrl *gomock.Controller) *MockCollections {
	mock := &MockCollections{ctrl: ctrl}
	mock.recorder = &MockCollectionsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCollections) EXPECT() *MockCollectionsMockRecorder {
	return m.recorder
}

// BatchStoreLightAndIndexByTransaction mocks base method.
func (m *MockCollections) BatchStoreLightAndIndexByTransaction(arg0 *flow.LightCollection, arg1 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchStoreLightAndIndexByTransaction", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchStoreLightAndIndexByTransaction indicates an expected call of BatchStoreLightAndIndexByTransaction.
func (mr *MockCollectionsMockRecorder) BatchStoreLightAndIndexByTransaction(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchStoreLightAndIndexByTransaction", reflect.TypeOf((*MockCollections)(nil).BatchStoreLightAndIndexByTransaction), arg0, arg1)
}

// ByID mocks base method.
func (m *MockCollections) ByID(arg0 flow.Identifier) (*flow.Collection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByID", arg0)
	ret0, _ := ret[0].(*flow.Collection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByID indicates an expected call of ByID.
func (mr *MockCollectionsMockRecorder) ByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByID", reflect.TypeOf((*MockCollections)(nil).ByID), arg0)
}

// LightByID mocks base method.
func (m *MockCollections) LightByID(arg0 flow.Identifier) (*flow.LightCollection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LightByID", arg0)
	ret0, _ := ret[0].(*flow.LightCollection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LightByID indicates an expected call of LightByID.
func (mr *MockCollectionsMockRecorder) LightByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LightByID", reflect.TypeOf((*MockCollections)(nil).LightByID), arg0)
}

// LightByTransactionID mocks base method.
func (m *MockCollections) LightByTransactionID(arg0 flow.Identifier) (*flow.LightCollection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LightByTransactionID", arg0)
	ret0, _ := ret[0].(*flow.LightCollection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LightByTransactionID indicates an expected call of LightByTransactionID.
func (mr *MockCollectionsMockRecorder) LightByTransactionID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LightByTransactionID", reflect.TypeOf((*MockCollections)(nil).LightByTransactionID), arg0)
}

// Remove mocks base method.
func (m *MockCollections) Remove(arg0 flow.Identifier) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockCollectionsMockRecorder) Remove(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockCollections)(nil).Remove), arg0)
}

// Store mocks base method.
func (m *MockCollections) Store(arg0 *flow.Collection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockCollectionsMockRecorder) Store(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockCollections)(nil).Store), arg0)
}

// StoreLightAndIndexByTransaction mocks base method.
func (m *MockCollections) StoreLightAndIndexByTransaction(arg0 *flow.LightCollection) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreLightAndIndexByTransaction", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreLightAndIndexByTransaction indicates an expected call of StoreLightAndIndexByTransaction.
func (mr *MockCollectionsMockRecorder) StoreLightAndIndexByTransaction(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreLightAndIndexByTransaction", reflect.TypeOf((*MockCollections)(nil).StoreLightAndIndexByTransaction), arg0)
}

// MockCommits is a mock of Commits interface.
type MockCommits struct {
	ctrl     *gomock.Controller
	recorder *MockCommitsMockRecorder
}

// MockCommitsMockRecorder is the mock recorder for MockCommits.
type MockCommitsMockRecorder struct {
	mock *MockCommits
}

// NewMockCommits creates a new mock instance.
func NewMockCommits(ctrl *gomock.Controller) *MockCommits {
	mock := &MockCommits{ctrl: ctrl}
	mock.recorder = &MockCommitsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCommits) EXPECT() *MockCommitsMockRecorder {
	return m.recorder
}

// BatchRemoveByBlockID mocks base method.
func (m *MockCommits) BatchRemoveByBlockID(arg0 flow.Identifier, arg1 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchRemoveByBlockID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchRemoveByBlockID indicates an expected call of BatchRemoveByBlockID.
func (mr *MockCommitsMockRecorder) BatchRemoveByBlockID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchRemoveByBlockID", reflect.TypeOf((*MockCommits)(nil).BatchRemoveByBlockID), arg0, arg1)
}

// BatchStore mocks base method.
func (m *MockCommits) BatchStore(arg0 flow.Identifier, arg1 flow.StateCommitment, arg2 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchStore", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchStore indicates an expected call of BatchStore.
func (mr *MockCommitsMockRecorder) BatchStore(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchStore", reflect.TypeOf((*MockCommits)(nil).BatchStore), arg0, arg1, arg2)
}

// ByBlockID mocks base method.
func (m *MockCommits) ByBlockID(arg0 flow.Identifier) (flow.StateCommitment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockID", arg0)
	ret0, _ := ret[0].(flow.StateCommitment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockID indicates an expected call of ByBlockID.
func (mr *MockCommitsMockRecorder) ByBlockID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockID", reflect.TypeOf((*MockCommits)(nil).ByBlockID), arg0)
}

// Store mocks base method.
func (m *MockCommits) Store(arg0 flow.Identifier, arg1 flow.StateCommitment) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockCommitsMockRecorder) Store(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockCommits)(nil).Store), arg0, arg1)
}

// MockEvents is a mock of Events interface.
type MockEvents struct {
	ctrl     *gomock.Controller
	recorder *MockEventsMockRecorder
}

// MockEventsMockRecorder is the mock recorder for MockEvents.
type MockEventsMockRecorder struct {
	mock *MockEvents
}

// NewMockEvents creates a new mock instance.
func NewMockEvents(ctrl *gomock.Controller) *MockEvents {
	mock := &MockEvents{ctrl: ctrl}
	mock.recorder = &MockEventsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEvents) EXPECT() *MockEventsMockRecorder {
	return m.recorder
}

// BatchRemoveByBlockID mocks base method.
func (m *MockEvents) BatchRemoveByBlockID(arg0 flow.Identifier, arg1 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchRemoveByBlockID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchRemoveByBlockID indicates an expected call of BatchRemoveByBlockID.
func (mr *MockEventsMockRecorder) BatchRemoveByBlockID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchRemoveByBlockID", reflect.TypeOf((*MockEvents)(nil).BatchRemoveByBlockID), arg0, arg1)
}

// BatchStore mocks base method.
func (m *MockEvents) BatchStore(arg0 flow.Identifier, arg1 []flow.EventsList, arg2 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchStore", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchStore indicates an expected call of BatchStore.
func (mr *MockEventsMockRecorder) BatchStore(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchStore", reflect.TypeOf((*MockEvents)(nil).BatchStore), arg0, arg1, arg2)
}

// ByBlockID mocks base method.
func (m *MockEvents) ByBlockID(arg0 flow.Identifier) ([]flow.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockID", arg0)
	ret0, _ := ret[0].([]flow.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockID indicates an expected call of ByBlockID.
func (mr *MockEventsMockRecorder) ByBlockID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockID", reflect.TypeOf((*MockEvents)(nil).ByBlockID), arg0)
}

// ByBlockIDEventType mocks base method.
func (m *MockEvents) ByBlockIDEventType(arg0 flow.Identifier, arg1 flow.EventType) ([]flow.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockIDEventType", arg0, arg1)
	ret0, _ := ret[0].([]flow.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockIDEventType indicates an expected call of ByBlockIDEventType.
func (mr *MockEventsMockRecorder) ByBlockIDEventType(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockIDEventType", reflect.TypeOf((*MockEvents)(nil).ByBlockIDEventType), arg0, arg1)
}

// ByBlockIDTransactionID mocks base method.
func (m *MockEvents) ByBlockIDTransactionID(arg0, arg1 flow.Identifier) ([]flow.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockIDTransactionID", arg0, arg1)
	ret0, _ := ret[0].([]flow.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockIDTransactionID indicates an expected call of ByBlockIDTransactionID.
func (mr *MockEventsMockRecorder) ByBlockIDTransactionID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockIDTransactionID", reflect.TypeOf((*MockEvents)(nil).ByBlockIDTransactionID), arg0, arg1)
}

// ByBlockIDTransactionIndex mocks base method.
func (m *MockEvents) ByBlockIDTransactionIndex(arg0 flow.Identifier, arg1 uint32) ([]flow.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockIDTransactionIndex", arg0, arg1)
	ret0, _ := ret[0].([]flow.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockIDTransactionIndex indicates an expected call of ByBlockIDTransactionIndex.
func (mr *MockEventsMockRecorder) ByBlockIDTransactionIndex(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockIDTransactionIndex", reflect.TypeOf((*MockEvents)(nil).ByBlockIDTransactionIndex), arg0, arg1)
}

// Store mocks base method.
func (m *MockEvents) Store(arg0 flow.Identifier, arg1 []flow.EventsList) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockEventsMockRecorder) Store(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockEvents)(nil).Store), arg0, arg1)
}

// MockServiceEvents is a mock of ServiceEvents interface.
type MockServiceEvents struct {
	ctrl     *gomock.Controller
	recorder *MockServiceEventsMockRecorder
}

// MockServiceEventsMockRecorder is the mock recorder for MockServiceEvents.
type MockServiceEventsMockRecorder struct {
	mock *MockServiceEvents
}

// NewMockServiceEvents creates a new mock instance.
func NewMockServiceEvents(ctrl *gomock.Controller) *MockServiceEvents {
	mock := &MockServiceEvents{ctrl: ctrl}
	mock.recorder = &MockServiceEventsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServiceEvents) EXPECT() *MockServiceEventsMockRecorder {
	return m.recorder
}

// BatchRemoveByBlockID mocks base method.
func (m *MockServiceEvents) BatchRemoveByBlockID(arg0 flow.Identifier, arg1 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchRemoveByBlockID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchRemoveByBlockID indicates an expected call of BatchRemoveByBlockID.
func (mr *MockServiceEventsMockRecorder) BatchRemoveByBlockID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchRemoveByBlockID", reflect.TypeOf((*MockServiceEvents)(nil).BatchRemoveByBlockID), arg0, arg1)
}

// BatchStore mocks base method.
func (m *MockServiceEvents) BatchStore(arg0 flow.Identifier, arg1 []flow.Event, arg2 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchStore", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchStore indicates an expected call of BatchStore.
func (mr *MockServiceEventsMockRecorder) BatchStore(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchStore", reflect.TypeOf((*MockServiceEvents)(nil).BatchStore), arg0, arg1, arg2)
}

// ByBlockID mocks base method.
func (m *MockServiceEvents) ByBlockID(arg0 flow.Identifier) ([]flow.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockID", arg0)
	ret0, _ := ret[0].([]flow.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockID indicates an expected call of ByBlockID.
func (mr *MockServiceEventsMockRecorder) ByBlockID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockID", reflect.TypeOf((*MockServiceEvents)(nil).ByBlockID), arg0)
}

// MockTransactionResults is a mock of TransactionResults interface.
type MockTransactionResults struct {
	ctrl     *gomock.Controller
	recorder *MockTransactionResultsMockRecorder
}

// MockTransactionResultsMockRecorder is the mock recorder for MockTransactionResults.
type MockTransactionResultsMockRecorder struct {
	mock *MockTransactionResults
}

// NewMockTransactionResults creates a new mock instance.
func NewMockTransactionResults(ctrl *gomock.Controller) *MockTransactionResults {
	mock := &MockTransactionResults{ctrl: ctrl}
	mock.recorder = &MockTransactionResultsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransactionResults) EXPECT() *MockTransactionResultsMockRecorder {
	return m.recorder
}

// BatchRemoveByBlockID mocks base method.
func (m *MockTransactionResults) BatchRemoveByBlockID(arg0 flow.Identifier, arg1 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchRemoveByBlockID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchRemoveByBlockID indicates an expected call of BatchRemoveByBlockID.
func (mr *MockTransactionResultsMockRecorder) BatchRemoveByBlockID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchRemoveByBlockID", reflect.TypeOf((*MockTransactionResults)(nil).BatchRemoveByBlockID), arg0, arg1)
}

// BatchStore mocks base method.
func (m *MockTransactionResults) BatchStore(arg0 flow.Identifier, arg1 []flow.TransactionResult, arg2 storage.ReaderBatchWriter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchStore", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchStore indicates an expected call of BatchStore.
func (mr *MockTransactionResultsMockRecorder) BatchStore(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchStore", reflect.TypeOf((*MockTransactionResults)(nil).BatchStore), arg0, arg1, arg2)
}

// ByBlockID mocks base method.
func (m *MockTransactionResults) ByBlockID(arg0 flow.Identifier) ([]flow.TransactionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockID", arg0)
	ret0, _ := ret[0].([]flow.TransactionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockID indicates an expected call of ByBlockID.
func (mr *MockTransactionResultsMockRecorder) ByBlockID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockID", reflect.TypeOf((*MockTransactionResults)(nil).ByBlockID), arg0)
}

// ByBlockIDTransactionID mocks base method.
func (m *MockTransactionResults) ByBlockIDTransactionID(arg0, arg1 flow.Identifier) (*flow.TransactionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockIDTransactionID", arg0, arg1)
	ret0, _ := ret[0].(*flow.TransactionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockIDTransactionID indicates an expected call of ByBlockIDTransactionID.
func (mr *MockTransactionResultsMockRecorder) ByBlockIDTransactionID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockIDTransactionID", reflect.TypeOf((*MockTransactionResults)(nil).ByBlockIDTransactionID), arg0, arg1)
}

// ByBlockIDTransactionIndex mocks base method.
func (m *MockTransactionResults) ByBlockIDTransactionIndex(arg0 flow.Identifier, arg1 uint32) (*flow.TransactionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ByBlockIDTransactionIndex", arg0, arg1)
	ret0, _ := ret[0].(*flow.TransactionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ByBlockIDTransactionIndex indicates an expected call of ByBlockIDTransactionIndex.
func (mr *MockTransactionResultsMockRecorder) ByBlockIDTransactionIndex(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ByBlockIDTransactionIndex", reflect.TypeOf((*MockTransactionResults)(nil).ByBlockIDTransactionIndex), arg0, arg1)
}
