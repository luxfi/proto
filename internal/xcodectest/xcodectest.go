// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package xcodectest is the canonical test wiring for the proto/x XVM
// codec set. proto/x carries no github.com/luxfi/codec import after the
// Wave 1A rip (#101); this helper package is the bridge that lets test
// suites under proto/x construct linearcodec-backed ParserCodecs.
//
// Production callers (luxfi/sdk/wallet/chain/x/..., luxfi/node/vms/xvm/...)
// construct their codecs inline. This helper exists strictly so the
// in-tree test files don't need to duplicate the wiring.
package xcodectest

import (
	"math"

	"github.com/luxfi/codec"
	"github.com/luxfi/codec/linearcodec"
	"github.com/luxfi/proto/x/txs"
)

// New returns a ParserCodecs bundle backed by two fresh linearcodec
// registries and two codec.Manager instances, one for runtime txs and
// one for genesis txs (with the larger MaxInt32 size budget). Each call
// produces independent codecs — safe to use per-test.
//
// Tx-level and fx-owned wire payload types are registered when this
// bundle is handed to txs.NewParser — see parser.go fxOwnedTypes for
// the canonical per-fx registration list.
func New() txs.ParserCodecs {
	c := linearcodec.NewDefault()
	gc := linearcodec.NewDefault()
	cm := codec.NewDefaultManager()
	gcm := codec.NewManager(math.MaxInt32)
	if err := cm.RegisterCodec(txs.CodecVersion, c); err != nil {
		panic(err)
	}
	if err := gcm.RegisterCodec(txs.CodecVersion, gc); err != nil {
		panic(err)
	}
	return txs.ParserCodecs{
		Codec:           cm,
		GenesisCodec:    gcm,
		Registry:        c,
		GenesisRegistry: gc,
	}
}

// NewRuntimeCodec returns a single linearcodec-backed codec.Manager
// for runtime tx wire bytes. Used by tests that need a Codec directly
// without going through a Parser. Fx types are NOT auto-registered
// here — callers that need fx payload roundtrip should use New().
func NewRuntimeCodec() (txs.Codec, txs.Registry) {
	c := linearcodec.NewDefault()
	cm := codec.NewDefaultManager()
	if err := cm.RegisterCodec(txs.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm, c
}
