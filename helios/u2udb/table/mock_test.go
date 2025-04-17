// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Fantom-foundation/lachesis-base/u2udb (interfaces: DBProducer,DropableStore)

// Package table is a generated GoMock package.
package table

import (
	reflect "reflect"

	u2udb "github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	gomock "github.com/golang/mock/gomock"
)

// MockDBProducer is a mock of DBProducer interface.
type MockDBProducer struct {
	ctrl     *gomock.Controller
	recorder *MockDBProducerMockRecorder
}

// MockDBProducerMockRecorder is the mock recorder for MockDBProducer.
type MockDBProducerMockRecorder struct {
	mock *MockDBProducer
}

// NewMockDBProducer creates a new mock instance.
func NewMockDBProducer(ctrl *gomock.Controller) *MockDBProducer {
	mock := &MockDBProducer{ctrl: ctrl}
	mock.recorder = &MockDBProducerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBProducer) EXPECT() *MockDBProducerMockRecorder {
	return m.recorder
}

// OpenDB mocks base method.
func (m *MockDBProducer) OpenDB(arg0 string) (u2udb.Store, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenDB", arg0)
	ret0, _ := ret[0].(u2udb.Store)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenDB indicates an expected call of OpenDB.
func (mr *MockDBProducerMockRecorder) OpenDB(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenDB", reflect.TypeOf((*MockDBProducer)(nil).OpenDB), arg0)
}

// MockDropableStore is a mock of DropableStore interface.
type MockDropableStore struct {
	ctrl     *gomock.Controller
	recorder *MockDropableStoreMockRecorder
}

// MockDropableStoreMockRecorder is the mock recorder for MockDropableStore.
type MockDropableStoreMockRecorder struct {
	mock *MockDropableStore
}

// NewMockDropableStore creates a new mock instance.
func NewMockDropableStore(ctrl *gomock.Controller) *MockDropableStore {
	mock := &MockDropableStore{ctrl: ctrl}
	mock.recorder = &MockDropableStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDropableStore) EXPECT() *MockDropableStoreMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockDropableStore) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockDropableStoreMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDropableStore)(nil).Close))
}

// Compact mocks base method.
func (m *MockDropableStore) Compact(arg0, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Compact", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Compact indicates an expected call of Compact.
func (mr *MockDropableStoreMockRecorder) Compact(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Compact", reflect.TypeOf((*MockDropableStore)(nil).Compact), arg0, arg1)
}

// Delete mocks base method.
func (m *MockDropableStore) Delete(arg0 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockDropableStoreMockRecorder) Delete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockDropableStore)(nil).Delete), arg0)
}

// Drop mocks base method.
func (m *MockDropableStore) Drop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Drop")
}

// Drop indicates an expected call of Drop.
func (mr *MockDropableStoreMockRecorder) Drop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Drop", reflect.TypeOf((*MockDropableStore)(nil).Drop))
}

// Get mocks base method.
func (m *MockDropableStore) Get(arg0 []byte) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockDropableStoreMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDropableStore)(nil).Get), arg0)
}

// GetSnapshot mocks base method.
func (m *MockDropableStore) GetSnapshot() (u2udb.Snapshot, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSnapshot")
	ret0, _ := ret[0].(u2udb.Snapshot)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSnapshot indicates an expected call of GetSnapshot.
func (mr *MockDropableStoreMockRecorder) GetSnapshot() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSnapshot", reflect.TypeOf((*MockDropableStore)(nil).GetSnapshot))
}

// Has mocks base method.
func (m *MockDropableStore) Has(arg0 []byte) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Has", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Has indicates an expected call of Has.
func (mr *MockDropableStoreMockRecorder) Has(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Has", reflect.TypeOf((*MockDropableStore)(nil).Has), arg0)
}

// NewBatch mocks base method.
func (m *MockDropableStore) NewBatch() u2udb.Batch {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewBatch")
	ret0, _ := ret[0].(u2udb.Batch)
	return ret0
}

// NewBatch indicates an expected call of NewBatch.
func (mr *MockDropableStoreMockRecorder) NewBatch() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewBatch", reflect.TypeOf((*MockDropableStore)(nil).NewBatch))
}

// NewIterator mocks base method.
func (m *MockDropableStore) NewIterator(arg0, arg1 []byte) u2udb.Iterator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewIterator", arg0, arg1)
	ret0, _ := ret[0].(u2udb.Iterator)
	return ret0
}

// NewIterator indicates an expected call of NewIterator.
func (mr *MockDropableStoreMockRecorder) NewIterator(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewIterator", reflect.TypeOf((*MockDropableStore)(nil).NewIterator), arg0, arg1)
}

// Put mocks base method.
func (m *MockDropableStore) Put(arg0, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockDropableStoreMockRecorder) Put(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockDropableStore)(nil).Put), arg0, arg1)
}

// Stat mocks base method.
func (m *MockDropableStore) Stat(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stat", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stat indicates an expected call of Stat.
func (mr *MockDropableStoreMockRecorder) Stat(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stat", reflect.TypeOf((*MockDropableStore)(nil).Stat), arg0)
}
