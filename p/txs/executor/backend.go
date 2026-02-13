// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"github.com/luxfi/atomic"
	"github.com/luxfi/ids"
	log "github.com/luxfi/log"
	"github.com/luxfi/protocol/p/config"
	"github.com/luxfi/protocol/p/fx"
	"github.com/luxfi/protocol/p/reward"
	"github.com/luxfi/protocol/p/utxo"
	"github.com/luxfi/runtime"
	"github.com/luxfi/timer/mockable"
	"github.com/luxfi/validators/uptime"
)

type Backend struct {
	Config       *config.Internal
	Rt           *runtime.Runtime
	Clk          *mockable.Clock
	Fx           fx.Fx
	FlowChecker  utxo.Verifier
	Uptimes      uptime.Calculator
	Rewards      reward.Calculator
	Bootstrapped *atomic.Atomic[bool]
	Log          log.Logger
}

// SharedMemory provides cross-chain atomic operations
type SharedMemory interface {
	Get(peerChainID ids.ID, keys [][]byte) ([][]byte, error)
	Apply(requests map[ids.ID]interface{}, batch ...interface{}) error
}
