// Copyright (C) 2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_native

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/zap"
)

const (
	offsetInFullParent_InsList = 0 // 8B list pointer
	offsetInFullParent_SigArr  = 8 // 8B list pointer (shared SigIndices)
	sizeInFullParent           = 16
)

func buildInputFullParent(entries []InputFullListEntry) []byte {
	b := zap.NewBuilder(zap.HeaderSize + 64 + len(entries)*SizeTransferableInputFull)
	insOff, insCount, sigIdx := WriteInputFullList(b, entries)
	sigOff, sigCount := WriteSigIndicesArray(b, sigIdx)

	ob := b.StartObject(sizeInFullParent)
	ob.SetList(offsetInFullParent_InsList, insOff, insCount)
	ob.SetList(offsetInFullParent_SigArr, sigOff, sigCount)
	ob.FinishAsRoot()
	return b.Finish()
}

// TestInputFull_Stakeable_And_MultiSig_RoundTrip proves the stakeable-aware
// input primitive: a plain multi-sig-indexed input AND a stakeable.LockIn
// both survive Parse from raw bytes, with SigIndices sliced from the shared
// array.
func TestInputFull_Stakeable_And_MultiSig_RoundTrip(t *testing.T) {
	require := require.New(t)

	txid := ids.GenerateTestID()
	asset := ids.GenerateTestID()

	entries := []InputFullListEntry{
		{
			TxID: txid, OutputIndex: 3, AssetID: asset,
			StakeLocktime: 0, Amount: 1_000_000,
			SigIndices: []uint32{0, 2, 5}, // 3-sig plain input
		},
		{
			TxID: txid, OutputIndex: 7, AssetID: asset,
			StakeLocktime: 1_766_708_400, Amount: 42, // stakeable.LockIn
			SigIndices: []uint32{1},
		},
	}

	buf := buildInputFullParent(entries)
	msg, err := zap.Parse(buf)
	require.NoError(err)
	root := msg.Root()

	ins := InputFullListView(root, offsetInFullParent_InsList)
	sigArr := SigIndicesArrayView(root, offsetInFullParent_SigArr)

	require.Equal(2, ins.Len())
	require.Equal(4, sigArr.Len(), "3 + 1 sig indices in the shared array")

	i0 := ins.At(0)
	require.Equal(txid, i0.TxID())
	require.EqualValues(3, i0.OutputIndex())
	require.Equal(asset, i0.AssetID())
	require.False(i0.IsStakeableLocked())
	require.EqualValues(1_000_000, i0.Amount())
	require.Equal([]uint32{0, 2, 5}, i0.SigIndices(sigArr))

	i1 := ins.At(1)
	require.EqualValues(7, i1.OutputIndex())
	require.True(i1.IsStakeableLocked())
	require.EqualValues(1_766_708_400, i1.StakeLocktime())
	require.EqualValues(42, i1.Amount())
	require.Equal([]uint32{1}, i1.SigIndices(sigArr))
}
