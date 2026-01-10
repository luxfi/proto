// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txstest

import (
	"time"

	consensusctx "github.com/luxfi/consensus/context"
	"github.com/luxfi/ids"
	"github.com/luxfi/protocol/p/config"
	"github.com/luxfi/protocol/p/txs/fee"
	"github.com/luxfi/sdk/wallet/chain/p/builder"
	"github.com/luxfi/vm/components/gas"
)

func newContext(
	ctx *consensusctx.Context,
	networkID uint32,
	luxAssetID ids.ID,
	cfg *config.Config,
	internalCfg *config.Internal,
	timestamp time.Time,
) *builder.Context {
	builderContext := &builder.Context{
		NetworkID: networkID,
		ChainID:   ctx.ChainID,
		XAssetID:  luxAssetID,
	}

	// For test purposes, populate the fee configuration
	// If dynamic fees are configured, use those; otherwise use static fees
	if internalCfg != nil && internalCfg.DynamicFeeConfig.Weights != (gas.Dimensions{}) {
		// Use dynamic fee configuration
		builderContext.ComplexityWeights = internalCfg.DynamicFeeConfig.Weights
		builderContext.GasPrice = internalCfg.DynamicFeeConfig.MinPrice
	}

	// Always populate static fees as fallback or for non-dynamic transactions
	if cfg != nil {
		builderContext.StaticFeeConfig = fee.StaticConfig{
			TxFee:              cfg.TxFee,
			CreateAssetTxFee:   cfg.CreateAssetTxFee,
			CreateNetworkTxFee: cfg.CreateNetworkTxFee,
			CreateChainTxFee:   cfg.CreateChainTxFee,
		}
	}

	return builderContext
}
