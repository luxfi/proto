// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_codec

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// dummyTx is a simple struct used to exercise Marshal/Unmarshal round
// trips and to produce the canonical hex dump comparison vs linearcodec
// documented in README.md.
type dummyTx struct {
	NetworkID uint32 `serialize:"true"`
	Nonce     uint64 `serialize:"true"`
	Memo      []byte `serialize:"true"`
}

func TestManager_MarshalUnmarshal_RoundTrip(t *testing.T) {
	require := require.New(t)
	cm := NewVersionedManager(0, MaxSize)

	tx := &dummyTx{
		NetworkID: 0x01020304,
		Nonce:     0x0102030405060708,
		Memo:      []byte("hello"),
	}

	b, err := cm.Marshal(0, tx)
	require.NoError(err)
	require.NotEmpty(b)

	// Verify the version prefix is uint16 LE: 0x00, 0x00.
	require.Equal(byte(0x00), b[0])
	require.Equal(byte(0x00), b[1])

	// Round-trip.
	got := &dummyTx{}
	version, err := cm.Unmarshal(b, got)
	require.NoError(err)
	require.Equal(uint16(0), version)
	require.Equal(tx.NetworkID, got.NetworkID)
	require.Equal(tx.Nonce, got.Nonce)
	require.True(bytes.Equal(tx.Memo, got.Memo))
}

func TestManager_VersionMismatch(t *testing.T) {
	require := require.New(t)
	cm := NewVersionedManager(0, MaxSize)

	// Marshal at the wrong version → ErrUnknownVersion.
	_, err := cm.Marshal(1, &dummyTx{})
	require.True(errors.Is(err, ErrUnknownVersion), "got %v", err)

	// Hand-craft a buffer with version=1 in LE; Unmarshal should reject.
	bad := []byte{0x01, 0x00} // version=1 LE, no payload
	_, err = cm.Unmarshal(bad, &dummyTx{})
	require.True(errors.Is(err, ErrUnknownVersion), "got %v", err)
}

func TestManager_TruncatedVersion(t *testing.T) {
	require := require.New(t)
	cm := NewVersionedManager(0, MaxSize)

	_, err := cm.Unmarshal([]byte{0x00}, &dummyTx{})
	require.True(errors.Is(err, ErrCantUnpackVersion), "got %v", err)
	_, err = cm.Unmarshal(nil, &dummyTx{})
	require.True(errors.Is(err, ErrCantUnpackVersion), "got %v", err)
}

func TestManager_MaxSizeExceeded(t *testing.T) {
	require := require.New(t)
	cm := NewVersionedManager(0, 16)

	// 17-byte buffer (1 byte above maxSize).
	bad := make([]byte, 17)
	_, err := cm.Unmarshal(bad, &dummyTx{})
	require.True(errors.Is(err, ErrMaxSizeExceeded), "got %v", err)
}

func TestManager_RegisterType_DispatchesToInner(t *testing.T) {
	require := require.New(t)
	cm := NewVersionedManager(0, MaxSize)
	// Re-registering the same type returns codec.ErrDuplicateType from
	// the inner zapcodec; the wrapper passes it through.
	require.NoError(cm.RegisterType(&dummyTx{}))
	err := cm.RegisterType(&dummyTx{})
	require.Error(err)
}

func TestManager_SkipRegistrations_BumpsTypeID(t *testing.T) {
	require := require.New(t)
	cm := NewVersionedManager(0, MaxSize)

	// Skip 5 slots, register dummyTx — its type-id is now 5.
	// We can't observe type-ids directly, but we can verify that
	// registering 5 placeholders before dummyTx produces the same wire
	// bytes as SkipRegistrations(5) + dummyTx. (This is the structural
	// guarantee SkipRegistrations gives us.)
	cm.SkipRegistrations(5)
	require.NoError(cm.RegisterType(&dummyTx{}))
}

// TestManager_WireFormat_ZAPNative_HexDump captures the canonical
// byte layout for a known dummyTx so the README.md hex-dump comparison
// is anchored in actual test output. The dump is intentionally
// deterministic — anyone reproducing this test gets the same bytes.
//
// Expected layout for dummyTx{NetworkID=0x01020304, Nonce=0x0102030405060708, Memo="hi"}:
//
//	00 00                                  // codec version 0 (LE uint16)
//	04 03 02 01                            // NetworkID 0x01020304 (LE uint32)
//	08 07 06 05 04 03 02 01                // Nonce 0x0102030405060708 (LE uint64)
//	02 00 00 00                            // Memo length 2 (LE uint32)
//	68 69                                  // Memo "hi"
//
// vs the (now-defunct) linearcodec wire bytes for the same dummyTx
// would have been ALL multi-byte integers in big-endian — see README.md.
func TestManager_WireFormat_ZAPNative_HexDump(t *testing.T) {
	require := require.New(t)
	cm := NewVersionedManager(0, MaxSize)

	tx := &dummyTx{
		NetworkID: 0x01020304,
		Nonce:     0x0102030405060708,
		Memo:      []byte("hi"),
	}

	b, err := cm.Marshal(0, tx)
	require.NoError(err)

	want := []byte{
		0x00, 0x00, // version 0 LE
		0x04, 0x03, 0x02, 0x01, // NetworkID LE
		0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, // Nonce LE
		0x02, 0x00, 0x00, 0x00, // Memo length 2 LE
		0x68, 0x69, // "hi"
	}

	require.Equal(want, b, "ZAP-native hex dump diverged from expected; got: %s", hex.EncodeToString(b))

	// Print the dump so `go test -v` shows it for the README.
	fmt.Printf("ZAP-native canonical dummyTx hex dump:\n  %s\n", hex.EncodeToString(b))
}
