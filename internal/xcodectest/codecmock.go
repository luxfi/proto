// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package xcodectest

import (
	"reflect"

	"github.com/luxfi/proto/x/txs"
	gomock "go.uber.org/mock/gomock"
)

// CodecMock is a gomock-driven mock of the txs.Codec interface.
//
// It mirrors the legacy luxfi/codec/codecmock.Manager API surface so
// existing tests can be ported by swapping the import: codecmock.NewManager
// → xcodectest.NewCodecMock. The package deliberately does not depend on
// github.com/luxfi/codec — proto/x carries no such import after Wave 1A.
type CodecMock struct {
	ctrl     *gomock.Controller
	recorder *CodecMockRecorder
}

var _ txs.Codec = (*CodecMock)(nil)

// CodecMockRecorder is the mock recorder.
type CodecMockRecorder struct {
	mock *CodecMock
}

// NewCodecMock creates a fresh CodecMock.
func NewCodecMock(ctrl *gomock.Controller) *CodecMock {
	m := &CodecMock{ctrl: ctrl}
	m.recorder = &CodecMockRecorder{mock: m}
	return m
}

// EXPECT returns the recorder so tests can program expectations.
func (m *CodecMock) EXPECT() *CodecMockRecorder {
	return m.recorder
}

// Marshal records the call and returns the configured result.
func (m *CodecMock) Marshal(version uint16, source interface{}) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Marshal", version, source)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Marshal expectation.
func (mr *CodecMockRecorder) Marshal(version, source any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Marshal", reflect.TypeOf((*CodecMock)(nil).Marshal), version, source)
}

// Unmarshal records the call and returns the configured result.
func (m *CodecMock) Unmarshal(source []byte, destination interface{}) (uint16, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unmarshal", source, destination)
	ret0, _ := ret[0].(uint16)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Unmarshal expectation.
func (mr *CodecMockRecorder) Unmarshal(source, destination any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unmarshal", reflect.TypeOf((*CodecMock)(nil).Unmarshal), source, destination)
}

// Size records the call and returns the configured result.
func (m *CodecMock) Size(version uint16, value interface{}) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size", version, value)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Size expectation.
func (mr *CodecMockRecorder) Size(version, value any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*CodecMock)(nil).Size), version, value)
}
