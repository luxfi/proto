// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"cmp"
	"fmt"

	"github.com/luxfi/address"
	"github.com/luxfi/ids"
	"github.com/luxfi/ordering"
	"github.com/luxfi/proto/x/txs"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"
)

// Genesis represents the genesis state of the XVM
type Genesis struct {
	Txs []*GenesisAsset `serialize:"true"`
}

// GenesisAsset represents an asset in the genesis block
type GenesisAsset struct {
	Alias             string `serialize:"true"`
	txs.CreateAssetTx `serialize:"true"`
}

// Compare implements ordering.Sortable for GenesisAsset
func (g *GenesisAsset) Compare(other *GenesisAsset) int {
	return cmp.Compare(g.Alias, other.Alias)
}

// AssetInitialState describes the initial state of an asset
type AssetInitialState struct {
	FixedCap    []GenesisHolder
	VariableCap []GenesisOwners
}

// GenesisAssetDefinition describes a genesis asset and its initial state
type GenesisAssetDefinition struct {
	Name         string
	Symbol       string
	Denomination byte
	InitialState AssetInitialState
	Memo         []byte
}

// GenesisHolder describes how much asset is owned by an address
type GenesisHolder struct {
	Amount  uint64
	Address string
}

// GenesisOwners describes who can perform an action
type GenesisOwners struct {
	Threshold uint32
	Minters   []string
}

// NewGenesis creates a new Genesis from genesis data.
//
// The genesisCodec is the wire codec used to compute deterministic
// per-fx output bytes for the InitialState.Sort canonicalization step.
// It must already be wired through a Parser (see
// luxfi/proto/x/txs.NewParser) so that fx-owned output types
// (secp256k1fx.TransferOutput, secp256k1fx.MintOutput) are
// register-resolvable before Marshal is called.
//
// Wave 1A of the codec rip (#101): the codec is now an injected
// dependency rather than constructed inline — proto/x/genesis no
// longer carries any github.com/luxfi/codec import.
func NewGenesis(
	networkID uint32,
	genesisData map[string]GenesisAssetDefinition,
	genesisCodec txs.Codec,
) (*Genesis, error) {
	if genesisCodec == nil {
		return nil, fmt.Errorf("genesis: genesisCodec must be non-nil")
	}
	g := &Genesis{}
	for assetAlias, assetDefinition := range genesisData {
		asset := GenesisAsset{
			Alias: assetAlias,
			CreateAssetTx: txs.CreateAssetTx{
				BaseTx: txs.BaseTx{BaseTx: lux.BaseTx{
					NetworkID:    networkID,
					BlockchainID: ids.Empty,
					Memo:         assetDefinition.Memo,
				}},
				Name:         assetDefinition.Name,
				Symbol:       assetDefinition.Symbol,
				Denomination: assetDefinition.Denomination,
			},
		}

		initialState := &txs.InitialState{
			FxIndex: 0, // secp256k1fx
		}
		for _, holder := range assetDefinition.InitialState.FixedCap {
			_, addrbuff, err := address.ParseBech32(holder.Address)
			if err != nil {
				return nil, fmt.Errorf("problem parsing holder address: %w", err)
			}
			addr, err := ids.ToShortID(addrbuff)
			if err != nil {
				return nil, fmt.Errorf("problem parsing holder address: %w", err)
			}
			initialState.Outs = append(initialState.Outs, &secp256k1fx.TransferOutput{
				Amt: holder.Amount,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			})
		}
		for _, owners := range assetDefinition.InitialState.VariableCap {
			out := &secp256k1fx.MintOutput{
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: owners.Threshold,
				},
			}
			for _, addrStr := range owners.Minters {
				_, addrBytes, err := address.ParseBech32(addrStr)
				if err != nil {
					return nil, fmt.Errorf("problem parsing minters address: %w", err)
				}
				addr, err := ids.ToShortID(addrBytes)
				if err != nil {
					return nil, fmt.Errorf("problem parsing minters address: %w", err)
				}
				out.Addrs = append(out.Addrs, addr)
			}
			out.Sort()

			initialState.Outs = append(initialState.Outs, out)
		}

		if len(initialState.Outs) > 0 {
			initialState.Sort(genesisCodec)
			asset.States = append(asset.States, initialState)
		}

		ordering.Sort(asset.States)
		g.Txs = append(g.Txs, &asset)
	}
	ordering.Sort(g.Txs)

	return g, nil
}

// Bytes serializes the Genesis to bytes using the supplied genesis
// codec. The codec must accept Genesis and all its transitively
// referenced fx-owned types (see NewGenesis godoc).
//
// Wave 1A: the genesisCodec is now an injected dependency, supplied
// by the caller's Parser.GenesisCodec().
func (g *Genesis) Bytes(genesisCodec txs.Codec) ([]byte, error) {
	if genesisCodec == nil {
		return nil, fmt.Errorf("genesis: genesisCodec must be non-nil")
	}
	return genesisCodec.Marshal(txs.CodecVersion, g)
}
