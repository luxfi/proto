// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/constants"
	"github.com/luxfi/crypto/secp256k1"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/xcodectest"
	"github.com/luxfi/proto/x/fxs"
	"github.com/luxfi/proto/x/txs"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"
)

// errCantUnpackVersion is the proto/x-local sentinel matching the
// upstream codec.ErrCantUnpackVersion. We assert via error string
// because we deliberately don't import the codec package — Wave 1A.
var errCantUnpackVersion = errors.New("couldn't unpack codec version")

var (
	chainID = ids.GenerateTestID()
	keys    = secp256k1.TestKeys()
	assetID = ids.GenerateTestID()
)

func TestInvalidBlock(t *testing.T) {
	require := require.New(t)

	parser, err := NewParser(
		xcodectest.New(),
		[]fxs.Fx{
			&secp256k1fx.Fx{},
		},
	)
	require.NoError(err)

	_, err = parser.ParseBlock(nil)
	// The codec error sentinel lives in luxfi/codec — proto/x is
	// codec-free post Wave 1A — so we assert via the error's message.
	require.Error(err)
	require.Contains(err.Error(), errCantUnpackVersion.Error())
}

func TestStandardBlocks(t *testing.T) {
	// check standard block can be built and parsed
	require := require.New(t)

	parser, err := NewParser(
		xcodectest.New(),
		[]fxs.Fx{
			&secp256k1fx.Fx{},
		},
	)
	require.NoError(err)

	blkTimestamp := time.Now()
	parentID := ids.GenerateTestID()
	height := uint64(2022)
	cm := parser.Codec()
	txs, err := createTestTxs(cm)
	require.NoError(err)

	standardBlk, err := NewStandardBlock(parentID, height, blkTimestamp, txs, cm)
	require.NoError(err)

	// parse block
	parsed, err := parser.ParseBlock(standardBlk.Bytes())
	require.NoError(err)

	// compare content
	require.Equal(standardBlk.ID(), parsed.ID())
	require.Equal(standardBlk.Parent(), parsed.Parent())
	require.Equal(standardBlk.Height(), parsed.Height())
	require.Equal(standardBlk.Bytes(), parsed.Bytes())
	require.Equal(standardBlk.Timestamp(), parsed.Timestamp())

	require.IsType(&StandardBlock{}, parsed)
	parsedStandardBlk := parsed.(*StandardBlock)

	require.Equal(txs, parsedStandardBlk.Txs())
	require.Equal(parsed.Txs(), parsedStandardBlk.Txs())
}

func createTestTxs(cm txs.Codec) ([]*txs.Tx, error) {
	countTxs := 1
	testTxs := make([]*txs.Tx, 0, countTxs)
	for i := 0; i < countTxs; i++ {
		// Create the tx
		tx := &txs.Tx{Unsigned: &txs.BaseTx{BaseTx: lux.BaseTx{
			NetworkID:    constants.UnitTestID,
			BlockchainID: chainID,
			Outs: []*lux.TransferableOutput{{
				Asset: lux.Asset{ID: assetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: uint64(12345),
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{keys[0].PublicKey().Address()},
					},
				},
			}},
			Ins: []*lux.TransferableInput{{
				UTXOID: lux.UTXOID{
					TxID:        ids.ID{'t', 'x', 'I', 'D'},
					OutputIndex: 1,
				},
				Asset: lux.Asset{ID: assetID},
				In: &secp256k1fx.TransferInput{
					Amt: uint64(54321),
					Input: secp256k1fx.Input{
						SigIndices: []uint32{2},
					},
				},
			}},
			Memo: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}}}
		if err := tx.SignSECP256K1Fx(cm, [][]*secp256k1.PrivateKey{{keys[0]}}); err != nil {
			return nil, err
		}
		testTxs = append(testTxs, tx)
	}
	return testTxs, nil
}
