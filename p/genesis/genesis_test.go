// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/address"
	"github.com/luxfi/constants"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/genesis"
	"github.com/luxfi/proto/p/txs"
	"github.com/luxfi/utxo/secp256k1fx"
)

func newTestCodec() genesis.Codec {
	return pvmcodectest.NewPVMCodecs().GenesisCodec
}

func createTestGenesis(t *testing.T) *genesis.Genesis {
	require := require.New(t)
	c := newTestCodec()

	nodeID := ids.BuildTestNodeID([]byte{1})
	addr, err := address.FormatBech32(constants.UnitTestHRP, nodeID.Bytes())
	require.NoError(err)

	validator := genesis.PermissionlessValidator{
		Validator: genesis.Validator{
			StartTime: 0,
			EndTime:   20,
			NodeID:    nodeID,
		},
		RewardOwner: &genesis.Owner{
			Threshold: 1,
			Addresses: []string{addr},
		},
		Staked: []genesis.Allocation{{
			Amount:  987654321,
			Address: addr,
		}},
	}

	g, err := genesis.New(
		c,
		ids.ID{'d', 'u', 'm', 'm', 'y', ' ', 'I', 'D'},
		constants.UnitTestID,
		[]genesis.Allocation{
			{
				Address: addr,
				Amount:  123456789,
			},
		},
		[]genesis.PermissionlessValidator{validator},
		nil,
		5,
		0,
		"Test Genesis",
	)
	require.NoError(err)

	return g
}

func TestNewInvalidUTXOBalance(t *testing.T) {
	require := require.New(t)
	c := newTestCodec()
	nodeID := ids.BuildTestNodeID([]byte{1, 2, 3})
	addr, err := address.FormatBech32(constants.UnitTestHRP, nodeID.Bytes())
	require.NoError(err)

	utxo := genesis.Allocation{
		Address: addr,
		Amount:  0,
	}
	weight := uint64(987654321)
	validator := genesis.PermissionlessValidator{
		Validator: genesis.Validator{
			EndTime: 15,
			Weight:  weight,
			NodeID:  nodeID,
		},
		RewardOwner: &genesis.Owner{
			Threshold: 1,
			Addresses: []string{addr},
		},
		Staked: []genesis.Allocation{{
			Amount:  weight,
			Address: addr,
		}},
	}

	g, err := genesis.New(
		c,
		ids.Empty,
		constants.UnitTestID,
		[]genesis.Allocation{utxo},
		[]genesis.PermissionlessValidator{validator},
		nil,
		5,
		0,
		"",
	)
	require.Error(err)
	require.Nil(g)
}

func TestNewInvalidStakeWeight(t *testing.T) {
	require := require.New(t)
	c := newTestCodec()
	nodeID := ids.BuildTestNodeID([]byte{1, 2, 3})
	addr, err := address.FormatBech32(constants.UnitTestHRP, nodeID.Bytes())
	require.NoError(err)

	utxo := genesis.Allocation{
		Address: addr,
		Amount:  123456789,
	}

	validator := genesis.PermissionlessValidator{
		Validator: genesis.Validator{
			StartTime: 0,
			EndTime:   15,
			NodeID:    nodeID,
		},
		RewardOwner: &genesis.Owner{
			Threshold: 1,
			Addresses: []string{addr},
		},
		Staked: []genesis.Allocation{{
			Amount:  0,
			Address: addr,
		}},
	}

	g, err := genesis.New(
		c,
		ids.Empty,
		0,
		[]genesis.Allocation{utxo},
		[]genesis.PermissionlessValidator{validator},
		nil,
		5,
		0,
		"",
	)
	require.Error(err)
	require.Nil(g)
}

func TestNewInvalidEndtime(t *testing.T) {
	require := require.New(t)
	c := newTestCodec()
	nodeID := ids.BuildTestNodeID([]byte{1, 2, 3})
	addr, err := address.FormatBech32(constants.UnitTestHRP, nodeID.Bytes())
	require.NoError(err)

	utxo := genesis.Allocation{
		Address: addr,
		Amount:  123456789,
	}

	weight := uint64(987654321)
	validator := genesis.PermissionlessValidator{
		Validator: genesis.Validator{
			StartTime: 0,
			EndTime:   5,
			NodeID:    nodeID,
		},
		RewardOwner: &genesis.Owner{
			Threshold: 1,
			Addresses: []string{addr},
		},
		Staked: []genesis.Allocation{{
			Amount:  weight,
			Address: addr,
		}},
	}

	g, err := genesis.New(
		c,
		ids.Empty,
		constants.UnitTestID,
		[]genesis.Allocation{utxo},
		[]genesis.PermissionlessValidator{validator},
		nil,
		5,
		0,
		"",
	)
	require.Error(err)
	require.Nil(g)
}

func TestGenesisBytes(t *testing.T) {
	require := require.New(t)
	c := newTestCodec()
	g := createTestGenesis(t)
	bytes, err := g.Bytes(c)
	require.NoError(err)
	require.NotEmpty(bytes)
}

func TestGenesis(t *testing.T) {
	require := require.New(t)
	g := createTestGenesis(t)

	luxAssetID := ids.ID{'d', 'u', 'm', 'm', 'y', ' ', 'I', 'D'}
	nodeID := ids.BuildTestNodeID([]byte{1})
	require.Equal("Test Genesis", g.Message)

	// Validate allocations
	require.Len(g.UTXOs, 1)
	utxo := g.UTXOs[0]
	require.Equal(luxAssetID, utxo.Asset.ID)
	output, ok := utxo.Out.(*secp256k1fx.TransferOutput)
	require.True(ok)
	require.Equal(uint64(123456789), output.Amt)
	require.Len(output.OutputOwners.Addrs, 1)

	// Validate validator
	require.Len(g.Validators, 1)
	validator := g.Validators[0]
	txValidator, ok := validator.Unsigned.(*txs.AddValidatorTx)
	require.True(ok)
	require.Equal(nodeID, txValidator.Validator.NodeID)
	require.Equal(uint64(20), txValidator.Validator.End)
	require.Len(txValidator.StakeOuts, 1)
	stakeOut := txValidator.StakeOuts[0]
	stakeOutput, ok := stakeOut.Out.(*secp256k1fx.TransferOutput)
	require.True(ok)
	require.Equal(uint64(987654321), stakeOutput.Amt)

	require.Empty(g.Chains)
	require.Equal(uint64(5), g.Timestamp)
	require.Equal(uint64(0), g.InitialSupply)
}

func TestNewReturnsSortedValidators(t *testing.T) {
	require := require.New(t)
	c := newTestCodec()
	nodeID := ids.BuildTestNodeID([]byte{1})
	addr, err := address.FormatBech32(constants.UnitTestHRP, nodeID.Bytes())
	require.NoError(err)

	allocation := genesis.Allocation{
		Address: addr,
		Amount:  123456789,
	}

	weight := uint64(987654321)
	validator1 := genesis.PermissionlessValidator{
		Validator: genesis.Validator{
			StartTime: 0,
			EndTime:   20,
			NodeID:    nodeID,
		},
		RewardOwner: &genesis.Owner{
			Threshold: 1,
			Addresses: []string{addr},
		},
		Staked: []genesis.Allocation{{
			Amount:  weight,
			Address: addr,
		}},
	}

	validator2 := genesis.PermissionlessValidator{
		Validator: genesis.Validator{
			StartTime: 3,
			EndTime:   15,
			NodeID:    nodeID,
		},
		RewardOwner: &genesis.Owner{
			Threshold: 1,
			Addresses: []string{addr},
		},
		Staked: []genesis.Allocation{{
			Amount:  weight,
			Address: addr,
		}},
	}

	validator3 := genesis.PermissionlessValidator{
		Validator: genesis.Validator{
			StartTime: 1,
			EndTime:   10,
			NodeID:    nodeID,
		},
		RewardOwner: &genesis.Owner{
			Threshold: 1,
			Addresses: []string{addr},
		},
		Staked: []genesis.Allocation{{
			Amount:  weight,
			Address: addr,
		}},
	}

	luxAssetID := ids.ID{'d', 'u', 'm', 'm', 'y', ' ', 'I', 'D'}
	g, err := genesis.New(
		c,
		luxAssetID,
		constants.UnitTestID,
		[]genesis.Allocation{allocation},
		[]genesis.PermissionlessValidator{
			validator1,
			validator2,
			validator3,
		},
		nil,
		5,
		0,
		"",
	)
	require.NoError(err)
	genesisBytes, err := g.Bytes(c)
	require.NoError(err)
	require.NotEmpty(genesisBytes)
	require.Len(g.Validators, 3)
}

func TestAllocationCompare(t *testing.T) {
	var (
		smallerAddr = ids.ShortID{}
		largerAddr  = ids.ShortID{1}
	)
	smallerAddrStr, err := address.FormatBech32("lux", smallerAddr[:])
	require.NoError(t, err)
	largerAddrStr, err := address.FormatBech32("lux", largerAddr[:])
	require.NoError(t, err)

	type test struct {
		name     string
		alloc1   genesis.Allocation
		alloc2   genesis.Allocation
		expected int
	}
	tests := []test{
		{
			name:     "both empty",
			alloc1:   genesis.Allocation{},
			alloc2:   genesis.Allocation{},
			expected: 0,
		},
		{
			name:   "locktime smaller",
			alloc1: genesis.Allocation{},
			alloc2: genesis.Allocation{
				Locktime: 1,
			},
			expected: -1,
		},
		{
			name:   "amount smaller",
			alloc1: genesis.Allocation{},
			alloc2: genesis.Allocation{
				Amount: 1,
			},
			expected: -1,
		},
		{
			name: "address smaller",
			alloc1: genesis.Allocation{
				Address: smallerAddrStr,
			},
			alloc2: genesis.Allocation{
				Address: largerAddrStr,
			},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			require.Equal(tt.expected, tt.alloc1.Compare(tt.alloc2))
			require.Equal(-tt.expected, tt.alloc2.Compare(tt.alloc1))
		})
	}
}
