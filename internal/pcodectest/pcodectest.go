// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package pcodectest is the canonical test wiring for the proto/p PVM
// codec set. proto/p carries no github.com/luxfi/codec import after the
// Wave 2A rip (#101); this helper package is the bridge that lets test
// suites under proto/p construct linearcodec-backed codecs without
// duplicating wire-registration logic across every test file.
//
// Production callers (luxfi/node/vms/platformvm/...) construct their
// codecs inline. This helper exists strictly so the in-tree test files
// don't need to duplicate the wiring.
package pcodectest

import (
	"math"

	"github.com/luxfi/codec"
	"github.com/luxfi/codec/linearcodec"

	"github.com/luxfi/proto/p/block"
	"github.com/luxfi/proto/p/txs"
	"github.com/luxfi/proto/p/warp"
	warpmsg "github.com/luxfi/proto/p/warp/message"
	"github.com/luxfi/proto/p/warp/payload"
)

// NewPayloadCodec returns a codec.Manager + linearcodec pair with the
// canonical warp payload types (Hash, AddressedCall) registered. Used by
// proto/p/warp/payload tests in lieu of the legacy package-level
// payload.Codec singleton.
func NewPayloadCodec() payload.Codec {
	c := linearcodec.NewDefault()
	cm := codec.NewManager(payload.MaxMessageSize)
	if err := payload.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(payload.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm
}

// NewMessageCodec returns a codec.Manager + linearcodec pair with the
// canonical warp/message types registered.
func NewMessageCodec() warpmsg.Codec {
	c := linearcodec.NewDefault()
	cm := codec.NewManager(math.MaxInt)
	if err := warpmsg.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(warpmsg.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm
}

// NewWarpCodec returns a codec.Manager + linearcodec pair with the
// canonical proto/p/warp signature + teleport types registered.
func NewWarpCodec() warp.Codec {
	c := linearcodec.NewDefault()
	cm := codec.NewManager(math.MaxInt)
	if err := warp.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(warp.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm
}

// PVMCodecs bundles the runtime and genesis PVM codecs with their
// underlying linearcodec registries. Tests that exercise both the
// regular and genesis codec paths (e.g. proto/p/block, proto/p/state)
// pull a single bundle.
type PVMCodecs struct {
	Codec            txs.Codec
	GenesisCodec     txs.Codec
	Registry         txs.LinearRegistry
	GenesisRegistry  txs.LinearRegistry
}

// NewPVMCodecs returns a PVMCodecs bundle backed by two fresh
// linearcodec registries and two codec.Manager instances, one for
// runtime txs and one for genesis txs (with the larger MaxInt32 size
// budget). Each call produces independent codecs — safe to use per-test.
// Both registries are pre-seeded with the full Apricot/Banff/Durango/
// Quasar block + tx type set, including the historical SkipRegistrations
// pre-amble.
func NewPVMCodecs() PVMCodecs {
	c := linearcodec.NewDefault()
	gc := linearcodec.NewDefault()
	cm := codec.NewDefaultManager()
	gcm := codec.NewManager(math.MaxInt32)
	if err := block.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := block.RegisterTypes(gc); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(txs.CodecVersion, c); err != nil {
		panic(err)
	}
	if err := gcm.RegisterCodec(txs.CodecVersion, gc); err != nil {
		panic(err)
	}
	return PVMCodecs{
		Codec:           cm,
		GenesisCodec:    gcm,
		Registry:        c,
		GenesisRegistry: gc,
	}
}

// NewPVMRuntimeCodec returns a single linearcodec-backed codec.Manager
// for runtime tx wire bytes. Used by tests that need a txs.Codec
// directly without going through a Parser.
func NewPVMRuntimeCodec() (txs.Codec, txs.LinearRegistry) {
	c := linearcodec.NewDefault()
	cm := codec.NewDefaultManager()
	if err := block.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(txs.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm, c
}

// NewPVMGenesisCodec returns a single linearcodec-backed codec.Manager
// for genesis tx wire bytes (MaxInt32 size budget).
func NewPVMGenesisCodec() (txs.Codec, txs.LinearRegistry) {
	c := linearcodec.NewDefault()
	cm := codec.NewManager(math.MaxInt32)
	if err := block.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(txs.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm, c
}
