// Copyright (C) 2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_native

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/zap"
)

// A minimal self-contained parent object carrying an OutputFullList + the
// shared OwnerAddrArray it slices into. This is the composition every real tx
// type will use once migrated off the single-address stub; testing it
// standalone proves the primitive round-trips before the type migration.
const (
	offsetOutFullParent_OutsList  = 0 // 8B list pointer
	offsetOutFullParent_AddrArray = 8 // 8B list pointer
	sizeOutFullParent             = 16
)

func buildOutputFullParent(entries []OutputFullListEntry) []byte {
	b := zap.NewBuilder(zap.HeaderSize + 64 + len(entries)*SizeTransferableOutputFull)
	outsOff, outsCount, addrsAll := WriteOutputFullList(b, entries)
	addrOff, addrCount := WriteOwnerAddrArray(b, addrsAll)

	ob := b.StartObject(sizeOutFullParent)
	ob.SetList(offsetOutFullParent_OutsList, outsOff, outsCount)
	ob.SetList(offsetOutFullParent_AddrArray, addrOff, addrCount)
	ob.FinishAsRoot()
	return b.Finish()
}

// TestOutputFull_MultiAddress_And_Stakeable_RoundTrip is the proof that the
// native output primitive faithfully carries what the single-address stub
// could not: a multi-address multisig owner AND a stakeable-locked output.
// This is the exact gap that forced multisig/staked outputs through the
// legacy reflection codec — closed here with a real round trip.
func TestOutputFull_MultiAddress_And_Stakeable_RoundTrip(t *testing.T) {
	require := require.New(t)

	asset := ids.GenerateTestID()
	a0, a1, a2 := ids.GenerateTestShortID(), ids.GenerateTestShortID(), ids.GenerateTestShortID()
	staker := ids.GenerateTestShortID()

	entries := []OutputFullListEntry{
		{
			// A 2-of-3 multisig plain TransferOutput with an owner locktime.
			AssetID:        asset,
			StakeLocktime:  0,
			Amount:         1_000_000,
			OwnerThreshold: 2,
			OwnerLocktime:  1_700_000_000,
			Addrs:          []ids.ShortID{a0, a1, a2},
		},
		{
			// A single-address stakeable.LockOut (StakeLocktime > 0).
			AssetID:        asset,
			StakeLocktime:  1_766_708_400,
			Amount:         42,
			OwnerThreshold: 1,
			OwnerLocktime:  0,
			Addrs:          []ids.ShortID{staker},
		},
	}

	buf := buildOutputFullParent(entries)

	// Re-parse from raw bytes (untrusted-input path) and read everything back.
	msg, err := zap.Parse(buf)
	require.NoError(err)
	root := msg.Root()

	outs := OutputFullListView(root, offsetOutFullParent_OutsList)
	addrArr := OwnerAddrArrayView(root, offsetOutFullParent_AddrArray)

	require.Equal(2, outs.Len())
	require.Equal(4, addrArr.Len(), "3 multisig addrs + 1 staker = 4 shared addrs")

	// Output 0: multisig plain output.
	o0 := outs.At(0)
	require.Equal(asset, o0.AssetID())
	require.False(o0.IsStakeableLocked())
	require.EqualValues(0, o0.StakeLocktime())
	require.EqualValues(1_000_000, o0.Amount())
	require.EqualValues(2, o0.OwnerThreshold())
	require.EqualValues(1_700_000_000, o0.OwnerLocktime())
	require.Equal([]ids.ShortID{a0, a1, a2}, o0.OwnerAddrs(addrArr),
		"multi-address owner must round-trip in exact order — the stub could not do this")

	// Output 1: single-address stakeable lock.
	o1 := outs.At(1)
	require.Equal(asset, o1.AssetID())
	require.True(o1.IsStakeableLocked())
	require.EqualValues(1_766_708_400, o1.StakeLocktime())
	require.EqualValues(42, o1.Amount())
	require.EqualValues(1, o1.OwnerThreshold())
	require.Equal([]ids.ShortID{staker}, o1.OwnerAddrs(addrArr))
}

// TestOutputFull_Empty_RoundTrip confirms the empty-list fast path (null
// pointers) round-trips as an empty output set with no shared addresses.
func TestOutputFull_Empty_RoundTrip(t *testing.T) {
	require := require.New(t)

	buf := buildOutputFullParent(nil)
	msg, err := zap.Parse(buf)
	require.NoError(err)
	root := msg.Root()

	outs := OutputFullListView(root, offsetOutFullParent_OutsList)
	require.True(outs.IsNull() || outs.Len() == 0)
}

// TestOutputFull_AddrSliceClamp confirms the shared-array slice is clamped:
// an out-of-range (start,count) yields nil, never a panic or phantom signer.
func TestOutputFull_AddrSliceClamp(t *testing.T) {
	require := require.New(t)

	buf := buildOutputFullParent([]OutputFullListEntry{{
		AssetID: ids.GenerateTestID(), Amount: 1, OwnerThreshold: 1,
		Addrs: []ids.ShortID{ids.GenerateTestShortID()},
	}})
	msg, err := zap.Parse(buf)
	require.NoError(err)
	addrArr := OwnerAddrArrayView(msg.Root(), offsetOutFullParent_AddrArray)

	require.Nil(addrArr.Slice(5, 10), "start past end => nil")
	require.Nil(addrArr.Slice(0, 99), "count past end => nil")
	require.Len(addrArr.Slice(0, 1), 1, "in-range slice ok")
}
