// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/xcodectest"
	"github.com/luxfi/proto/x/txs"
	"github.com/luxfi/runtime"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/vm/components/verify"
)

type testOperable struct {
	lux.TestTransferable `serialize:"true"`

	Outputs []verify.State `serialize:"true"`
}

func (*testOperable) InitRuntime(*runtime.Runtime) {}

func (*testOperable) InitializeRuntime(*runtime.Runtime) error { return nil }

func (o *testOperable) Outs() []verify.State {
	return o.Outputs
}

func TestOperationVerifyNil(t *testing.T) {
	op := (*txs.Operation)(nil)
	err := op.Verify()
	require.ErrorIs(t, err, txs.ErrNilOperation)
}

func TestOperationVerifyEmpty(t *testing.T) {
	op := &txs.Operation{
		Asset: lux.Asset{ID: ids.Empty},
	}
	err := op.Verify()
	require.ErrorIs(t, err, txs.ErrNilFxOperation)
}

func TestOperationVerifyUTXOIDsNotSorted(t *testing.T) {
	op := &txs.Operation{
		Asset: lux.Asset{ID: ids.Empty},
		UTXOIDs: []*lux.UTXOID{
			{
				TxID:        ids.Empty,
				OutputIndex: 1,
			},
			{
				TxID:        ids.Empty,
				OutputIndex: 0,
			},
		},
		Op: &testOperable{},
	}
	err := op.Verify()
	require.ErrorIs(t, err, txs.ErrNotSortedAndUniqueUTXOIDs)
}

func TestOperationVerify(t *testing.T) {
	assetID := ids.GenerateTestID()
	op := &txs.Operation{
		Asset: lux.Asset{ID: assetID},
		UTXOIDs: []*lux.UTXOID{
			{
				TxID:        assetID,
				OutputIndex: 1,
			},
		},
		Op: &testOperable{},
	}
	require.NoError(t, op.Verify())
}

func TestOperationSorting(t *testing.T) {
	require := require.New(t)

	m, c := xcodectest.NewRuntimeCodec()
	require.NoError(c.RegisterType(&testOperable{}))

	ops := []*txs.Operation{
		{
			Asset: lux.Asset{ID: ids.Empty},
			UTXOIDs: []*lux.UTXOID{
				{
					TxID:        ids.Empty,
					OutputIndex: 1,
				},
			},
			Op: &testOperable{},
		},
		{
			Asset: lux.Asset{ID: ids.Empty},
			UTXOIDs: []*lux.UTXOID{
				{
					TxID:        ids.Empty,
					OutputIndex: 0,
				},
			},
			Op: &testOperable{},
		},
	}
	require.False(txs.IsSortedAndUniqueOperations(ops, m))
	txs.SortOperations(ops, m)
	require.True(txs.IsSortedAndUniqueOperations(ops, m))
	ops = append(ops, &txs.Operation{
		Asset: lux.Asset{ID: ids.Empty},
		UTXOIDs: []*lux.UTXOID{
			{
				TxID:        ids.Empty,
				OutputIndex: 1,
			},
		},
		Op: &testOperable{},
	})
	require.False(txs.IsSortedAndUniqueOperations(ops, m))
}

func TestOperationTxNotState(t *testing.T) {
	intf := interface{}(&txs.OperationTx{})
	_, ok := intf.(verify.State)
	require.False(t, ok)
}
