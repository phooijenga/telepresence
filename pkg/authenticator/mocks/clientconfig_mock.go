// Code generated by MockGen. DO NOT EDIT.
// SourceAddr: k8s.io/client-go/tools/clientcmd (interfaces: ClientConfig)

// Package mock_authenticator is a generated GoMock package.
package mock_authenticator

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	rest "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	api "k8s.io/client-go/tools/clientcmd/api"
)

// MockClientConfig is a mock of ClientConfig interface.
type MockClientConfig struct {
	ctrl     *gomock.Controller
	recorder *MockClientConfigMockRecorder
}

// MockClientConfigMockRecorder is the mock recorder for MockClientConfig.
type MockClientConfigMockRecorder struct {
	mock *MockClientConfig
}

// NewMockClientConfig creates a new mock instance.
func NewMockClientConfig(ctrl *gomock.Controller) *MockClientConfig {
	mock := &MockClientConfig{ctrl: ctrl}
	mock.recorder = &MockClientConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientConfig) EXPECT() *MockClientConfigMockRecorder {
	return m.recorder
}

// ClientConfig mocks base method.
func (m *MockClientConfig) ClientConfig() (*rest.Config, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClientConfig")
	ret0, _ := ret[0].(*rest.Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ClientConfig indicates an expected call of ClientConfig.
func (mr *MockClientConfigMockRecorder) ClientConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClientConfig", reflect.TypeOf((*MockClientConfig)(nil).ClientConfig))
}

// ConfigAccess mocks base method.
func (m *MockClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfigAccess")
	ret0, _ := ret[0].(clientcmd.ConfigAccess)
	return ret0
}

// ConfigAccess indicates an expected call of ConfigAccess.
func (mr *MockClientConfigMockRecorder) ConfigAccess() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfigAccess", reflect.TypeOf((*MockClientConfig)(nil).ConfigAccess))
}

// Namespace mocks base method.
func (m *MockClientConfig) Namespace() (string, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Namespace")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Namespace indicates an expected call of Namespace.
func (mr *MockClientConfigMockRecorder) Namespace() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Namespace", reflect.TypeOf((*MockClientConfig)(nil).Namespace))
}

// RawConfig mocks base method.
func (m *MockClientConfig) RawConfig() (api.Config, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RawConfig")
	ret0, _ := ret[0].(api.Config)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RawConfig indicates an expected call of RawConfig.
func (mr *MockClientConfigMockRecorder) RawConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RawConfig", reflect.TypeOf((*MockClientConfig)(nil).RawConfig))
}
