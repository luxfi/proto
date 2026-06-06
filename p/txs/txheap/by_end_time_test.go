// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txheap_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/txs"
	"github.com/luxfi/proto/p/txs/txheap"
	"github.com/luxfi/utxo/secp256k1fx"
)

func TestByEndTime(t *testing.T) {
	require := require.New(t)
	c, _ := pvmcodectest.NewPVMRuntimeCodec()

	txHeap := txheap.NewByEndTime()

	baseTime := time.Now()

	utx0 := &txs.AddValidatorTx{
		Validator: txs.Validator{
			NodeID: ids.BuildTestNodeID([]byte{0}),
			Start:  uint64(baseTime.Unix()),
			End:    uint64(baseTime.Unix()) + 1,
		},
		RewardsOwner: &secp256k1fx.OutputOwners{},
	}
	tx0 := &txs.Tx{Unsigned: utx0}
	require.NoError(tx0.Initialize(c))

	utx1 := &txs.AddValidatorTx{
		Validator: txs.Validator{
			NodeID: ids.BuildTestNodeID([]byte{1}),
			Start:  uint64(baseTime.Unix()),
			End:    uint64(baseTime.Unix()) + 2,
		},
		RewardsOwner: &secp256k1fx.OutputOwners{},
	}
	tx1 := &txs.Tx{Unsigned: utx1}
	require.NoError(tx1.Initialize(c))

	utx2 := &txs.AddValidatorTx{
		Validator: txs.Validator{
			NodeID: ids.BuildTestNodeID([]byte{1}),
			Start:  uint64(baseTime.Unix()),
			End:    uint64(baseTime.Unix()) + 3,
		},
		RewardsOwner: &secp256k1fx.OutputOwners{},
	}
	tx2 := &txs.Tx{Unsigned: utx2}
	require.NoError(tx2.Initialize(c))

	txHeap.Add(tx2)
	require.Equal(utx2.EndTime(), txHeap.Timestamp())

	txHeap.Add(tx1)
	require.Equal(utx1.EndTime(), txHeap.Timestamp())

	txHeap.Add(tx0)
	require.Equal(utx0.EndTime(), txHeap.Timestamp())
	require.Equal(tx0, txHeap.Peek())
}
