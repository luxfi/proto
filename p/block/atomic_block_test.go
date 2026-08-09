// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/block"
	"github.com/luxfi/proto/p/txs"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/vm/components/verify"
)

func TestNewApricotAtomicBlock(t *testing.T) {
	require := require.New(t)
	codecs := pvmcodectest.NewPVMCodecs()
	c := codecs.GenesisCodec

	parentID := ids.GenerateTestID()
	height := uint64(1337)
	tx := &txs.Tx{
		Unsigned: &txs.ImportTx{
			BaseTx: txs.BaseTx{
				BaseTx: lux.BaseTx{
					Ins:  []*lux.TransferableInput{},
					Outs: []*lux.TransferableOutput{},
				},
			},
			ImportedInputs: []*lux.TransferableInput{},
		},
		Creds: []verify.Verifiable{},
	}
	require.NoError(tx.Initialize())

	blk, err := block.NewApricotAtomicBlock(c,
		parentID,
		height,
		tx,
	)
	require.NoError(err)

	// Make sure the block and tx are initialized
	require.NotEmpty(blk.Bytes())
	require.NotEmpty(blk.Tx.Bytes())
	require.NotEqual(ids.Empty, blk.Tx.ID())
	require.Equal(tx.Bytes(), blk.Tx.Bytes())
	require.Equal(parentID, blk.Parent())
	require.Equal(height, blk.Height())
}
