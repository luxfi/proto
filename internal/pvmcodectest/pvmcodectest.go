// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package pvmcodectest is the PVM-aware companion to
// proto/internal/pcodectest. It carries the proto/p package imports
// (block / txs / warp / warp/message / warp/payload) so the basic
// pcodectest helpers can stay free of those imports — that split
// lets tests inside proto/p/txs (the lowest layer in the PVM dep
// graph) reach for pcodectest without closing a test-time import
// cycle.
//
// Use pcodectest when you just need a Codec + Manager pair. Use
// pvmcodectest when you need the PVM type set registered against
// that codec — proto/p/{block,state,txs/executor,txs/fee} etc. reach
// for the helpers here.
//
// Wire format is ZAP-native (little-endian) — proto/zap_codec is the
// single canonical construction site for the wire codec choice (LP-023).
package pvmcodectest

import (
	"errors"
	"math"

	"github.com/luxfi/proto/p/block"
	"github.com/luxfi/proto/p/txs"
	"github.com/luxfi/proto/p/warp"
	warpmsg "github.com/luxfi/proto/p/warp/message"
	"github.com/luxfi/proto/p/warp/payload"
	"github.com/luxfi/proto/zap_codec"
)

// Metadata codec version tags. These mirror state.CodecVersion0Tag and
// state.CodecVersion1Tag — duplicated here to keep pvmcodectest free of
// a state-package import (and the resulting test-time import cycle).
const (
	metadataCodecVersion0Tag        = "v0"
	metadataCodecVersion0    uint16 = 0

	metadataCodecVersion1Tag        = "v1"
	metadataCodecVersion1    uint16 = 1
)

// NewPayloadCodec returns a ZAP-native Manager with the canonical warp
// payload types (Hash, AddressedCall) registered. Used by
// proto/p/warp/payload tests in lieu of the legacy package-level
// payload.Codec singleton.
//
// NOTE: this helper imports proto/p/warp/payload — proto/p/warp/payload
// tests that need their own codec must construct it inline rather than
// reach for this helper (would close a test-time import cycle).
func NewPayloadCodec() payload.Codec {
	mm := zap_codec.NewPayload(payload.CodecVersion, payload.MaxMessageSize)
	if err := payload.RegisterTypes(mm); err != nil {
		panic(err)
	}
	return mm
}

// NewMessageCodec returns a ZAP-native Manager with the canonical
// warp/message types registered.
func NewMessageCodec() warpmsg.Codec {
	mm := zap_codec.NewMessage(warpmsg.CodecVersion)
	if err := warpmsg.RegisterTypes(mm); err != nil {
		panic(err)
	}
	return mm
}

// NewWarpCodec returns a ZAP-native Manager with the canonical
// proto/p/warp signature + teleport types registered.
func NewWarpCodec() warp.Codec {
	mm := zap_codec.NewWarp(warp.CodecVersion)
	if err := warp.RegisterTypes(mm); err != nil {
		panic(err)
	}
	return mm
}

// PVMCodecs bundles the runtime and genesis PVM codecs with their
// underlying registries. Tests that exercise both the regular and
// genesis codec paths (e.g. proto/p/block, proto/p/state) pull a
// single bundle. Codec and Registry slots reference the same
// underlying *zap_codec.Manager per side — the Manager satisfies
// both interfaces by shape.
type PVMCodecs struct {
	Codec           txs.Codec
	GenesisCodec    txs.Codec
	Registry        txs.LinearRegistry
	GenesisRegistry txs.LinearRegistry
}

// NewPVMCodecs returns a PVMCodecs bundle backed by two fresh
// ZAP-native Manager instances, one for runtime txs and one for
// genesis txs (with the larger MaxInt32 size budget). Each call
// produces independent codecs — safe to use per-test. Both managers
// are pre-seeded with the full Apricot/Banff/Durango/Quasar block + tx
// type set, including the historical SkipRegistrations pre-amble.
//
// NOTE: this helper imports proto/p/block (which imports proto/p/txs).
// Tests inside proto/p/{block,txs} that need their own codec must
// construct it inline rather than reach for this helper (would close a
// test-time import cycle).
func NewPVMCodecs() PVMCodecs {
	runtime := zap_codec.NewPVMRuntime(txs.CodecVersion)
	genesis := zap_codec.NewPVMGenesis(txs.CodecVersion)
	if err := block.RegisterTypes(runtime); err != nil {
		panic(err)
	}
	if err := block.RegisterTypes(genesis); err != nil {
		panic(err)
	}
	return PVMCodecs{
		Codec:           runtime,
		GenesisCodec:    genesis,
		Registry:        runtime,
		GenesisRegistry: genesis,
	}
}

// NewPVMRuntimeCodec returns a single ZAP-native Manager seeded with
// the full PVM tx type set. Used by tests that need a txs.Codec
// directly without going through a Parser. The same Manager is
// returned as both Codec and LinearRegistry — it satisfies both
// interfaces by shape.
func NewPVMRuntimeCodec() (txs.Codec, txs.LinearRegistry) {
	mm := zap_codec.NewPVMRuntime(txs.CodecVersion)
	if err := block.RegisterTypes(mm); err != nil {
		panic(err)
	}
	return mm, mm
}

// NewPVMGenesisCodec returns a single ZAP-native Manager for genesis
// tx wire bytes (MaxInt32 size budget). Same caveat as
// NewPVMRuntimeCodec — the Manager is returned as both Codec and
// LinearRegistry.
func NewPVMGenesisCodec() (txs.Codec, txs.LinearRegistry) {
	mm := zap_codec.NewPVMGenesis(txs.CodecVersion)
	if err := block.RegisterTypes(mm); err != nil {
		panic(err)
	}
	return mm, mm
}

// NewMetadataCodec returns the validator/delegator metadata codec
// registered with the v0:"true" and v1:"true" tag layouts. This is a
// SEPARATE codec from the block/genesis codec — proto/p/state holds
// it as a distinct field on *state.
//
// Returns a multi-version Manager with two inner codec versions
// registered: 0 (v0-only tag) and 1 (v0 + v1 tags). The Manager
// satisfies state.MetadataCodec by shape so this helper can stay
// independent of the proto/p/state package and avoid an import cycle
// on the state_test files that need this codec.
func NewMetadataCodec() zap_codec.MultiManager {
	c0 := zap_codec.NewLinearCodecWithTags(metadataCodecVersion0Tag)
	c1 := zap_codec.NewLinearCodecWithTags(metadataCodecVersion0Tag, metadataCodecVersion1Tag)
	mm := zap_codec.NewManager(math.MaxInt32)

	err := errors.Join(
		mm.RegisterCodec(metadataCodecVersion0, c0),
		mm.RegisterCodec(metadataCodecVersion1, c1),
	)
	if err != nil {
		panic(err)
	}
	return mm
}
