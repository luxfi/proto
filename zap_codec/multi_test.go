// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_codec

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// dummyV0 / dummyV1 are two distinct serialized shapes registered at
// different codec versions on the same MultiManager. The round-trip
// test verifies the manager dispatches to the right inner codec based
// on the version embedded in the wire prefix.
type dummyV0 struct {
	A uint32 `serialize:"true"`
}

type dummyV1 struct {
	A uint32 `serialize:"true"`
	B uint64 `serialize:"true"`
}

func TestMultiManager_RoundTrip_PerVersion(t *testing.T) {
	require := require.New(t)
	mm := NewDefaultManager()

	c0 := NewLinearCodec()
	require.NoError(c0.RegisterType(&dummyV0{}))
	require.NoError(mm.RegisterCodec(0, c0))

	c1 := NewLinearCodec()
	require.NoError(c1.RegisterType(&dummyV1{}))
	require.NoError(mm.RegisterCodec(1, c1))

	v0 := &dummyV0{A: 0xdeadbeef}
	b0, err := mm.Marshal(0, v0)
	require.NoError(err)
	require.Equal(byte(0x00), b0[0]) // LE version 0
	require.Equal(byte(0x00), b0[1])

	v1 := &dummyV1{A: 0xcafe, B: 0x0102030405060708}
	b1, err := mm.Marshal(1, v1)
	require.NoError(err)
	require.Equal(byte(0x01), b1[0]) // LE version 1
	require.Equal(byte(0x00), b1[1])

	got0 := &dummyV0{}
	version, err := mm.Unmarshal(b0, got0)
	require.NoError(err)
	require.Equal(uint16(0), version)
	require.Equal(v0.A, got0.A)

	got1 := &dummyV1{}
	version, err = mm.Unmarshal(b1, got1)
	require.NoError(err)
	require.Equal(uint16(1), version)
	require.Equal(v1.A, got1.A)
	require.Equal(v1.B, got1.B)
}

func TestMultiManager_UnknownVersion_OnMarshal(t *testing.T) {
	require := require.New(t)
	mm := NewDefaultManager()

	_, err := mm.Marshal(0, &dummyV0{})
	require.True(errors.Is(err, ErrUnknownVersion), "got %v", err)
}

func TestMultiManager_UnknownVersion_OnUnmarshal(t *testing.T) {
	require := require.New(t)
	mm := NewDefaultManager()

	c0 := NewLinearCodec()
	require.NoError(c0.RegisterType(&dummyV0{}))
	require.NoError(mm.RegisterCodec(0, c0))

	// Bytes claim version 1 (which isn't registered).
	bad := []byte{0x01, 0x00}
	_, err := mm.Unmarshal(bad, &dummyV0{})
	require.True(errors.Is(err, ErrUnknownVersion), "got %v", err)
}

func TestMultiManager_DuplicateRegisterCodec(t *testing.T) {
	require := require.New(t)
	mm := NewDefaultManager()

	c := NewLinearCodec()
	require.NoError(mm.RegisterCodec(0, c))
	require.Error(mm.RegisterCodec(0, c))
}

func TestMultiManager_TruncatedVersion(t *testing.T) {
	require := require.New(t)
	mm := NewDefaultManager()

	_, err := mm.Unmarshal([]byte{0x00}, &dummyV0{})
	require.True(errors.Is(err, ErrCantUnpackVersion), "got %v", err)
	_, err = mm.Unmarshal(nil, &dummyV0{})
	require.True(errors.Is(err, ErrCantUnpackVersion), "got %v", err)
}

func TestMultiManager_TrailingBytes(t *testing.T) {
	require := require.New(t)
	mm := NewDefaultManager()

	c0 := NewLinearCodec()
	require.NoError(c0.RegisterType(&dummyV0{}))
	require.NoError(mm.RegisterCodec(0, c0))

	v0 := &dummyV0{A: 1}
	b, err := mm.Marshal(0, v0)
	require.NoError(err)

	// Append a trailing byte; Unmarshal should refuse.
	bad := append(b, 0xff)
	_, err = mm.Unmarshal(bad, &dummyV0{})
	require.True(errors.Is(err, ErrExtraSpace), "got %v", err)
}

func TestMultiManager_Size_IncludesVersionPrefix(t *testing.T) {
	require := require.New(t)
	mm := NewDefaultManager()

	c0 := NewLinearCodec()
	require.NoError(c0.RegisterType(&dummyV0{}))
	require.NoError(mm.RegisterCodec(0, c0))

	size, err := mm.Size(0, &dummyV0{A: 1})
	require.NoError(err)
	// VersionSize=2 + uint32=4 = 6 bytes total.
	require.Equal(VersionSize+4, size)
}

func TestNewMaxInt32Manager_AcceptsLargeBlobs(t *testing.T) {
	require := require.New(t)
	mm := NewMaxInt32Manager()

	c0 := NewLinearCodec()
	require.NoError(c0.RegisterType(&dummyV0{}))
	require.NoError(mm.RegisterCodec(0, c0))

	b, err := mm.Marshal(0, &dummyV0{A: 1})
	require.NoError(err)
	require.NotEmpty(b)
}

func TestNewMaxIntManager_AcceptsLargeBlobs(t *testing.T) {
	require := require.New(t)
	mm := NewMaxIntManager()

	c0 := NewLinearCodec()
	require.NoError(c0.RegisterType(&dummyV0{}))
	require.NoError(mm.RegisterCodec(0, c0))

	b, err := mm.Marshal(0, &dummyV0{A: 1})
	require.NoError(err)
	require.NotEmpty(b)
}
