// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/dfpath/dfpath.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDfpath is a mock of Dfpath interface.
type MockDfpath struct {
	ctrl     *gomock.Controller
	recorder *MockDfpathMockRecorder
}

// MockDfpathMockRecorder is the mock recorder for MockDfpath.
type MockDfpathMockRecorder struct {
	mock *MockDfpath
}

// NewMockDfpath creates a new mock instance.
func NewMockDfpath(ctrl *gomock.Controller) *MockDfpath {
	mock := &MockDfpath{ctrl: ctrl}
	mock.recorder = &MockDfpathMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDfpath) EXPECT() *MockDfpathMockRecorder {
	return m.recorder
}

// CacheDir mocks base method.
func (m *MockDfpath) CacheDir() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CacheDir")
	ret0, _ := ret[0].(string)
	return ret0
}

// CacheDir indicates an expected call of CacheDir.
func (mr *MockDfpathMockRecorder) CacheDir() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheDir", reflect.TypeOf((*MockDfpath)(nil).CacheDir))
}

// DaemonLockPath mocks base method.
func (m *MockDfpath) DaemonLockPath() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DaemonLockPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// DaemonLockPath indicates an expected call of DaemonLockPath.
func (mr *MockDfpathMockRecorder) DaemonLockPath() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DaemonLockPath", reflect.TypeOf((*MockDfpath)(nil).DaemonLockPath))
}

// DaemonSockPath mocks base method.
func (m *MockDfpath) DaemonSockPath() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DaemonSockPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// DaemonSockPath indicates an expected call of DaemonSockPath.
func (mr *MockDfpathMockRecorder) DaemonSockPath() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DaemonSockPath", reflect.TypeOf((*MockDfpath)(nil).DaemonSockPath))
}

// DataDir mocks base method.
func (m *MockDfpath) DataDir() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DataDir")
	ret0, _ := ret[0].(string)
	return ret0
}

// DataDir indicates an expected call of DataDir.
func (mr *MockDfpathMockRecorder) DataDir() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DataDir", reflect.TypeOf((*MockDfpath)(nil).DataDir))
}

// DfgetLockPath mocks base method.
func (m *MockDfpath) DfgetLockPath() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DfgetLockPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// DfgetLockPath indicates an expected call of DfgetLockPath.
func (mr *MockDfpathMockRecorder) DfgetLockPath() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DfgetLockPath", reflect.TypeOf((*MockDfpath)(nil).DfgetLockPath))
}

// LogDir mocks base method.
func (m *MockDfpath) LogDir() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LogDir")
	ret0, _ := ret[0].(string)
	return ret0
}

// LogDir indicates an expected call of LogDir.
func (mr *MockDfpathMockRecorder) LogDir() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogDir", reflect.TypeOf((*MockDfpath)(nil).LogDir))
}

// PluginDir mocks base method.
func (m *MockDfpath) PluginDir() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PluginDir")
	ret0, _ := ret[0].(string)
	return ret0
}

// PluginDir indicates an expected call of PluginDir.
func (mr *MockDfpathMockRecorder) PluginDir() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PluginDir", reflect.TypeOf((*MockDfpath)(nil).PluginDir))
}

// WorkHome mocks base method.
func (m *MockDfpath) WorkHome() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WorkHome")
	ret0, _ := ret[0].(string)
	return ret0
}

// WorkHome indicates an expected call of WorkHome.
func (mr *MockDfpathMockRecorder) WorkHome() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WorkHome", reflect.TypeOf((*MockDfpath)(nil).WorkHome))
}