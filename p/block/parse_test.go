// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/crypto/secp256k1"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/block"
	"github.com/luxfi/proto/p/txs"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"
)

var preFundedKeys = secp256k1.TestKeys()

// pvmCodecs lazy-constructs both runtime and genesis codecs once per
// test process. Tests iterate over both to exercise the historical
// double-roundtrip the legacy Codec / GenesisCodec singletons guarded.
func pvmCodecs() pvmcodectest.PVMCodecs {
	return pvmcodectest.NewPVMCodecs()
}

func TestStandardBlocks(t *testing.T) {
	// check Apricot standard block can be built and parsed
	require := require.New(t)
	codecs := pvmCodecs()
	blkTimestamp := time.Now()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	decisionTxs, err := testDecisionTxs(codecs.Codec)
	require.NoError(err)

	for _, cdc := range []block.Codec{codecs.Codec, codecs.GenesisCodec} {
		// build block
		apricotStandardBlk, err := block.NewApricotStandardBlock(cdc, parentID, height, decisionTxs)
		require.NoError(err)

		// parse block
		parsed, err := block.Parse(cdc, apricotStandardBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(apricotStandardBlk.ID(), parsed.ID())
		require.Equal(apricotStandardBlk.Bytes(), parsed.Bytes())
		require.Equal(apricotStandardBlk.Parent(), parsed.Parent())
		require.Equal(apricotStandardBlk.Height(), parsed.Height())

		require.IsType(&block.ApricotStandardBlock{}, parsed)
		require.Equal(decisionTxs, parsed.Txs())

		// check that banff standard block can be built and parsed
		banffStandardBlk, err := block.NewBanffStandardBlock(cdc, blkTimestamp, parentID, height, decisionTxs)
		require.NoError(err)

		// parse block
		parsed, err = block.Parse(cdc, banffStandardBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(banffStandardBlk.ID(), parsed.ID())
		require.Equal(banffStandardBlk.Bytes(), parsed.Bytes())
		require.Equal(banffStandardBlk.Parent(), parsed.Parent())
		require.Equal(banffStandardBlk.Height(), parsed.Height())
		require.IsType(&block.BanffStandardBlock{}, parsed)
		parsedBanffStandardBlk := parsed.(*block.BanffStandardBlock)
		require.Equal(decisionTxs, parsedBanffStandardBlk.Txs())

		// timestamp check for banff blocks only
		require.Equal(banffStandardBlk.Timestamp(), parsedBanffStandardBlk.Timestamp())

		// backward compatibility check
		require.Equal(parsed.Txs(), parsedBanffStandardBlk.Txs())
	}
}

func TestProposalBlocks(t *testing.T) {
	// check Apricot proposal block can be built and parsed
	require := require.New(t)
	codecs := pvmCodecs()
	blkTimestamp := time.Now()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	proposalTx, err := testProposalTx(codecs.Codec)
	require.NoError(err)
	decisionTxs, err := testDecisionTxs(codecs.Codec)
	require.NoError(err)

	for _, cdc := range []block.Codec{codecs.Codec, codecs.GenesisCodec} {
		// build block
		apricotProposalBlk, err := block.NewApricotProposalBlock(cdc,
			parentID,
			height,
			proposalTx,
		)
		require.NoError(err)

		// parse block
		parsed, err := block.Parse(cdc, apricotProposalBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(apricotProposalBlk.ID(), parsed.ID())
		require.Equal(apricotProposalBlk.Bytes(), parsed.Bytes())
		require.Equal(apricotProposalBlk.Parent(), parsed.Parent())
		require.Equal(apricotProposalBlk.Height(), parsed.Height())

		require.IsType(&block.ApricotProposalBlock{}, parsed)
		parsedApricotProposalBlk := parsed.(*block.ApricotProposalBlock)
		require.Equal([]*txs.Tx{proposalTx}, parsedApricotProposalBlk.Txs())

		// check that banff proposal block can be built and parsed
		banffProposalBlk, err := block.NewBanffProposalBlock(cdc,
			blkTimestamp,
			parentID,
			height,
			proposalTx,
			[]*txs.Tx{},
		)
		require.NoError(err)

		// parse block
		parsed, err = block.Parse(cdc, banffProposalBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(banffProposalBlk.ID(), parsed.ID())
		require.Equal(banffProposalBlk.Bytes(), parsed.Bytes())
		require.Equal(banffProposalBlk.Parent(), parsed.Parent())
		require.Equal(banffProposalBlk.Height(), parsed.Height())
		require.IsType(&block.BanffProposalBlock{}, parsed)
		parsedBanffProposalBlk := parsed.(*block.BanffProposalBlock)
		require.Equal([]*txs.Tx{proposalTx}, parsedBanffProposalBlk.Txs())

		// timestamp check for banff blocks only
		require.Equal(banffProposalBlk.Timestamp(), parsedBanffProposalBlk.Timestamp())

		// backward compatibility check
		require.Equal(parsedApricotProposalBlk.Txs(), parsedBanffProposalBlk.Txs())

		// check that banff proposal block with decisionTxs can be built and parsed
		banffProposalBlkWithDecisionTxs, err := block.NewBanffProposalBlock(cdc,
			blkTimestamp,
			parentID,
			height,
			proposalTx,
			decisionTxs,
		)
		require.NoError(err)

		// parse block
		parsed, err = block.Parse(cdc, banffProposalBlkWithDecisionTxs.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(banffProposalBlkWithDecisionTxs.ID(), parsed.ID())
		require.Equal(banffProposalBlkWithDecisionTxs.Bytes(), parsed.Bytes())
		require.Equal(banffProposalBlkWithDecisionTxs.Parent(), parsed.Parent())
		require.Equal(banffProposalBlkWithDecisionTxs.Height(), parsed.Height())
		require.IsType(&block.BanffProposalBlock{}, parsed)
		parsedBanffProposalBlkWithDecisionTxs := parsed.(*block.BanffProposalBlock)

		l := len(decisionTxs)
		expectedTxs := make([]*txs.Tx, l+1)
		copy(expectedTxs, decisionTxs)
		expectedTxs[l] = proposalTx
		require.Equal(expectedTxs, parsedBanffProposalBlkWithDecisionTxs.Txs())

		require.Equal(banffProposalBlkWithDecisionTxs.Timestamp(), parsedBanffProposalBlkWithDecisionTxs.Timestamp())
	}
}

func TestCommitBlock(t *testing.T) {
	// check Apricot commit block can be built and parsed
	require := require.New(t)
	codecs := pvmCodecs()
	blkTimestamp := time.Now()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)

	for _, cdc := range []block.Codec{codecs.Codec, codecs.GenesisCodec} {
		// build block
		apricotCommitBlk, err := block.NewApricotCommitBlock(cdc, parentID, height)
		require.NoError(err)

		// parse block
		parsed, err := block.Parse(cdc, apricotCommitBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(apricotCommitBlk.ID(), parsed.ID())
		require.Equal(apricotCommitBlk.Bytes(), parsed.Bytes())
		require.Equal(apricotCommitBlk.Parent(), parsed.Parent())
		require.Equal(apricotCommitBlk.Height(), parsed.Height())

		// check that banff commit block can be built and parsed
		banffCommitBlk, err := block.NewBanffCommitBlock(cdc, blkTimestamp, parentID, height)
		require.NoError(err)

		// parse block
		parsed, err = block.Parse(cdc, banffCommitBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(banffCommitBlk.ID(), parsed.ID())
		require.Equal(banffCommitBlk.Bytes(), parsed.Bytes())
		require.Equal(banffCommitBlk.Parent(), parsed.Parent())
		require.Equal(banffCommitBlk.Height(), parsed.Height())

		// timestamp check for banff blocks only
		require.IsType(&block.BanffCommitBlock{}, parsed)
		parsedBanffCommitBlk := parsed.(*block.BanffCommitBlock)
		require.Equal(banffCommitBlk.Timestamp(), parsedBanffCommitBlk.Timestamp())
	}
}

func TestAbortBlock(t *testing.T) {
	// check Apricot abort block can be built and parsed
	require := require.New(t)
	codecs := pvmCodecs()
	blkTimestamp := time.Now()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)

	for _, cdc := range []block.Codec{codecs.Codec, codecs.GenesisCodec} {
		// build block
		apricotAbortBlk, err := block.NewApricotAbortBlock(cdc, parentID, height)
		require.NoError(err)

		// parse block
		parsed, err := block.Parse(cdc, apricotAbortBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(apricotAbortBlk.ID(), parsed.ID())
		require.Equal(apricotAbortBlk.Bytes(), parsed.Bytes())
		require.Equal(apricotAbortBlk.Parent(), parsed.Parent())
		require.Equal(apricotAbortBlk.Height(), parsed.Height())

		// check that banff abort block can be built and parsed
		banffAbortBlk, err := block.NewBanffAbortBlock(cdc, blkTimestamp, parentID, height)
		require.NoError(err)

		// parse block
		parsed, err = block.Parse(cdc, banffAbortBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(banffAbortBlk.ID(), parsed.ID())
		require.Equal(banffAbortBlk.Bytes(), parsed.Bytes())
		require.Equal(banffAbortBlk.Parent(), parsed.Parent())
		require.Equal(banffAbortBlk.Height(), parsed.Height())

		// timestamp check for banff blocks only
		require.IsType(&block.BanffAbortBlock{}, parsed)
		parsedBanffAbortBlk := parsed.(*block.BanffAbortBlock)
		require.Equal(banffAbortBlk.Timestamp(), parsedBanffAbortBlk.Timestamp())
	}
}

func TestAtomicBlock(t *testing.T) {
	// check atomic block can be built and parsed
	require := require.New(t)
	codecs := pvmCodecs()
	parentID := ids.ID{'p', 'a', 'r', 'e', 'n', 't', 'I', 'D'}
	height := uint64(2022)
	atomicTx, err := testAtomicTx(codecs.Codec)
	require.NoError(err)

	for _, cdc := range []block.Codec{codecs.Codec, codecs.GenesisCodec} {
		// build block
		atomicBlk, err := block.NewApricotAtomicBlock(cdc,
			parentID,
			height,
			atomicTx,
		)
		require.NoError(err)

		// parse block
		parsed, err := block.Parse(cdc, atomicBlk.Bytes())
		require.NoError(err)

		// compare content
		require.Equal(atomicBlk.ID(), parsed.ID())
		require.Equal(atomicBlk.Bytes(), parsed.Bytes())
		require.Equal(atomicBlk.Parent(), parsed.Parent())
		require.Equal(atomicBlk.Height(), parsed.Height())

		require.IsType(&block.ApricotAtomicBlock{}, parsed)
		parsedAtomicBlk := parsed.(*block.ApricotAtomicBlock)
		require.Equal([]*txs.Tx{atomicTx}, parsedAtomicBlk.Txs())
	}
}

func testAtomicTx(c txs.Codec) (*txs.Tx, error) {
	utx := &txs.ImportTx{
		BaseTx: txs.BaseTx{BaseTx: lux.BaseTx{
			NetworkID:    10,
			BlockchainID: ids.ID{'c', 'h', 'a', 'i', 'n', 'I', 'D'},
			Outs: []*lux.TransferableOutput{{
				Asset: lux.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
				Out: &secp256k1fx.TransferOutput{
					Amt: uint64(1234),
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{preFundedKeys[0].Address()},
					},
				},
			}},
			Ins: []*lux.TransferableInput{{
				UTXOID: lux.UTXOID{
					TxID:        ids.ID{'t', 'x', 'I', 'D'},
					OutputIndex: 2,
				},
				Asset: lux.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
				In: &secp256k1fx.TransferInput{
					Amt:   uint64(5678),
					Input: secp256k1fx.Input{SigIndices: []uint32{0}},
				},
			}},
			Memo: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}},
		SourceChain: ids.ID{'c', 'h', 'a', 'i', 'n'},
		ImportedInputs: []*lux.TransferableInput{{
			UTXOID: lux.UTXOID{
				TxID:        ids.Empty.Prefix(1),
				OutputIndex: 1,
			},
			Asset: lux.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
			In: &secp256k1fx.TransferInput{
				Amt:   50000,
				Input: secp256k1fx.Input{SigIndices: []uint32{0}},
			},
		}},
	}
	signers := [][]*secp256k1.PrivateKey{{preFundedKeys[0]}}
	return txs.NewSigned(utx, c, signers)
}

func testDecisionTxs(c txs.Codec) ([]*txs.Tx, error) {
	countTxs := 2
	decisionTxs := make([]*txs.Tx, 0, countTxs)
	for i := 0; i < countTxs; i++ {
		// Create the tx
		utx := &txs.CreateChainTx{
			BaseTx: txs.BaseTx{BaseTx: lux.BaseTx{
				NetworkID:    10,
				BlockchainID: ids.ID{'c', 'h', 'a', 'i', 'n', 'I', 'D'},
				Outs: []*lux.TransferableOutput{{
					Asset: lux.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
					Out: &secp256k1fx.TransferOutput{
						Amt: uint64(1234),
						OutputOwners: secp256k1fx.OutputOwners{
							Threshold: 1,
							Addrs:     []ids.ShortID{preFundedKeys[0].Address()},
						},
					},
				}},
				Ins: []*lux.TransferableInput{{
					UTXOID: lux.UTXOID{
						TxID:        ids.ID{'t', 'x', 'I', 'D'},
						OutputIndex: 2,
					},
					Asset: lux.Asset{ID: ids.ID{'a', 's', 's', 'e', 'r', 't'}},
					In: &secp256k1fx.TransferInput{
						Amt:   uint64(5678),
						Input: secp256k1fx.Input{SigIndices: []uint32{0}},
					},
				}},
				Memo: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			}},
			ValidateNetworkID: ids.ID{'s', 'u', 'b', 'n', 'e', 't', 'I', 'D'},
			BlockchainName:    "a chain",
			VMID:              ids.GenerateTestID(),
			FxIDs:             []ids.ID{ids.GenerateTestID()},
			GenesisData:       []byte{'g', 'e', 'n', 'D', 'a', 't', 'a'},
			ChainAuth:         &secp256k1fx.Input{SigIndices: []uint32{1}},
		}

		signers := [][]*secp256k1.PrivateKey{{preFundedKeys[0]}}
		tx, err := txs.NewSigned(utx, c, signers)
		if err != nil {
			return nil, err
		}
		decisionTxs = append(decisionTxs, tx)
	}
	return decisionTxs, nil
}

func testProposalTx(c txs.Codec) (*txs.Tx, error) {
	utx := &txs.RewardValidatorTx{
		TxID: ids.ID{'r', 'e', 'w', 'a', 'r', 'd', 'I', 'D'},
	}

	signers := [][]*secp256k1.PrivateKey{{preFundedKeys[0]}}
	return txs.NewSigned(utx, c, signers)
}
