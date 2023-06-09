// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package service is a generated GoMock package.
package service

import (
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// AddNode mocks base method.
func (m *MockService) AddNode() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddNode")
	ret0, _ := ret[0].(error)
	return ret0
}

// AddNode indicates an expected call of AddNode.
func (mr *MockServiceMockRecorder) AddNode() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddNode", reflect.TypeOf((*MockService)(nil).AddNode))
}

// Load mocks base method.
func (m *MockService) Load(ctx context.Context, name string) (io.ReadCloser, uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Load", ctx, name)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Load indicates an expected call of Load.
func (mr *MockServiceMockRecorder) Load(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Load", reflect.TypeOf((*MockService)(nil).Load), ctx, name)
}

// Stats mocks base method.
func (m *MockService) Stats() []float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stats")
	ret0, _ := ret[0].([]float64)
	return ret0
}

// Stats indicates an expected call of Stats.
func (mr *MockServiceMockRecorder) Stats() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stats", reflect.TypeOf((*MockService)(nil).Stats))
}

// Store mocks base method.
func (m *MockService) Store(ctx context.Context, name string, size uint64, obj io.Reader) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", ctx, name, size, obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockServiceMockRecorder) Store(ctx, name, size, obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockService)(nil).Store), ctx, name, size, obj)
}
