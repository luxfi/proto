// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package xcodectest is the canonical test wiring for the proto/x XVM
// codec set. proto/x carries no github.com/luxfi/codec import after the
// Wave 1A rip (#101); this helper package is the bridge that lets test
// suites under proto/x construct ZAP-native ParserCodecs.
//
// Production callers (luxfi/sdk/wallet/chain/x/..., luxfi/node/vms/xvm/...)
// construct their codecs inline via proto/zap_codec. This helper exists
// strictly so the in-tree test files don't need to duplicate the wiring.
//
// Wire format is ZAP-native (little-endian) — proto/zap_codec is the
// single canonical construction site for the wire codec choice (LP-023).
package xcodectest

import (
	"github.com/luxfi/proto/x/txs"
	"github.com/luxfi/proto/zap_codec"
)

// New returns a ParserCodecs bundle backed by two fresh ZAP-native
// Manager instances, one for runtime txs and one for genesis txs (with
// the larger math.MaxInt32 size budget). Each call produces independent
// codecs — safe to use per-test. Both Codec slots and both Registry
// slots reference the same underlying *zap_codec.Manager per side: the
// Manager satisfies both proto/x/txs.Codec (Marshal/Unmarshal/Size) and
// proto/x/txs.Registry (RegisterType) by shape.
//
// Tx-level and fx-owned wire payload types are registered when this
// bundle is handed to txs.NewParser — see parser.go fxOwnedTypes for
// the canonical per-fx registration list.
func New() txs.ParserCodecs {
	runtime, genesis := zap_codec.NewXVMParser(txs.CodecVersion)
	return txs.ParserCodecs{
		Codec:           runtime,
		GenesisCodec:    genesis,
		Registry:        runtime,
		GenesisRegistry: genesis,
	}
}

// NewRuntimeCodec returns a single ZAP-native Manager that satisfies
// both txs.Codec and txs.Registry. Used by tests that need a Codec
// directly without going through a Parser. Fx types are NOT auto-
// registered here — callers that need fx payload roundtrip should use
// New().
func NewRuntimeCodec() (txs.Codec, txs.Registry) {
	m := zap_codec.NewPVMRuntime(txs.CodecVersion)
	return m, m
}
