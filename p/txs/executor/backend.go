// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"github.com/luxfi/atomic"
	"github.com/luxfi/ids"
	log "github.com/luxfi/log"
	"github.com/luxfi/proto/p/config"
	"github.com/luxfi/proto/p/fx"
	"github.com/luxfi/proto/p/reward"
	"github.com/luxfi/proto/p/txs"
	"github.com/luxfi/proto/p/utxo"
	"github.com/luxfi/proto/p/warp"
	warpmsg "github.com/luxfi/proto/p/warp/message"
	"github.com/luxfi/proto/p/warp/payload"
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

	// Wire codecs threaded in from the PVM bundle. Each Backend instance
	// MUST hold non-nil codecs after Wave 2A of the codec rip (#101);
	// the executor uses them for every Marshal/Unmarshal that historically
	// went through a package-level codec.Manager singleton.
	//
	// TxCodec        — proto/p runtime tx codec (UTXO + owner marshalling)
	// WarpCodec      — proto/p/warp codec (UnsignedMessage / Message)
	// WarpMsgCodec   — proto/p/warp/message codec (RegisterL1Validator
	//                  / L1ValidatorRegistration / L1ValidatorWeight /
	//                  ChainToL1Conversion)
	// PayloadCodec   — proto/p/warp/payload codec (Hash + AddressedCall)
	TxCodec      txs.Codec
	WarpCodec    warp.Codec
	WarpMsgCodec warpmsg.Codec
	PayloadCodec payload.Codec
}

// SharedMemory provides cross-chain atomic operations
type SharedMemory interface {
	Get(peerChainID ids.ID, keys [][]byte) ([][]byte, error)
	Apply(requests map[ids.ID]interface{}, batch ...interface{}) error
}
