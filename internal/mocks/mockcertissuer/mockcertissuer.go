// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/suzerain-io/placeholder-name/pkg/registry/loginrequest (interfaces: CertIssuer)

// Package mockcertissuer is a generated GoMock package.
package mockcertissuer

import (
	pkix "crypto/x509/pkix"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockCertIssuer is a mock of CertIssuer interface
type MockCertIssuer struct {
	ctrl     *gomock.Controller
	recorder *MockCertIssuerMockRecorder
}

// MockCertIssuerMockRecorder is the mock recorder for MockCertIssuer
type MockCertIssuerMockRecorder struct {
	mock *MockCertIssuer
}

// NewMockCertIssuer creates a new mock instance
func NewMockCertIssuer(ctrl *gomock.Controller) *MockCertIssuer {
	mock := &MockCertIssuer{ctrl: ctrl}
	mock.recorder = &MockCertIssuerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCertIssuer) EXPECT() *MockCertIssuerMockRecorder {
	return m.recorder
}

// IssuePEM mocks base method
func (m *MockCertIssuer) IssuePEM(arg0 pkix.Name, arg1 []string, arg2 time.Duration) ([]byte, []byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IssuePEM", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// IssuePEM indicates an expected call of IssuePEM
func (mr *MockCertIssuerMockRecorder) IssuePEM(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IssuePEM", reflect.TypeOf((*MockCertIssuer)(nil).IssuePEM), arg0, arg1, arg2)
}