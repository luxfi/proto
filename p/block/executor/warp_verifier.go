// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"context"

	"github.com/luxfi/proto/p/block"
	txexecutor "github.com/luxfi/proto/p/txs/executor"
	"github.com/luxfi/proto/p/warp"
	validators "github.com/luxfi/validators"
)

// VerifyWarpMessages verifies all warp messages in the block. The
// warpCodec is the proto/p/warp codec used to parse cross-chain warp
// messages embedded in the L1-validator family of txs; callers thread
// it in from their PVM bundle. If any of the warp messages are
// invalid, an error is returned.
func VerifyWarpMessages(
	ctx context.Context,
	networkID uint32,
	warpCodec warp.Codec,
	validatorState validators.State,
	pChainHeight uint64,
	b block.Block,
) error {
	for _, tx := range b.Txs() {
		err := txexecutor.VerifyWarpMessages(
			ctx,
			networkID,
			warpCodec,
			validatorState,
			pChainHeight,
			tx.Unsigned,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
