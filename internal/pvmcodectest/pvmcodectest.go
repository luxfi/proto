// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package pvmcodectest is the PVM-aware companion to
// proto/internal/pcodectest. It carries the proto/p package imports
// (block / txs / warp / warp/message / warp/payload) so the basic
// pcodectest helpers can stay free of those imports — that split
// lets tests inside proto/p/txs (the lowest layer in the PVM dep
// graph) reach for pcodectest without closing a test-time import
// cycle.
//
// Use pcodectest when you just need a linearcodec + codec.Manager
// pair. Use pvmcodectest when you need the PVM type set registered
// against that codec — proto/p/{block,state,txs/executor,txs/fee} etc.
// reach for the helpers here.
package pvmcodectest

import (
	"errors"
	"math"

	"github.com/luxfi/codec"
	"github.com/luxfi/codec/linearcodec"

	"github.com/luxfi/proto/p/block"
	"github.com/luxfi/proto/p/txs"
	"github.com/luxfi/proto/p/warp"
	warpmsg "github.com/luxfi/proto/p/warp/message"
	"github.com/luxfi/proto/p/warp/payload"
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

// NewPayloadCodec returns a codec.Manager + linearcodec pair with the
// canonical warp payload types (Hash, AddressedCall) registered. Used by
// proto/p/warp/payload tests in lieu of the legacy package-level
// payload.Codec singleton.
//
// NOTE: this helper imports proto/p/warp/payload — proto/p/warp/payload
// tests that need their own codec must construct it inline rather than
// reach for this helper (would close a test-time import cycle).
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
	Codec           txs.Codec
	GenesisCodec    txs.Codec
	Registry        txs.LinearRegistry
	GenesisRegistry txs.LinearRegistry
}

// NewPVMCodecs returns a PVMCodecs bundle backed by two fresh
// linearcodec registries and two codec.Manager instances, one for
// runtime txs and one for genesis txs (with the larger MaxInt32 size
// budget). Each call produces independent codecs — safe to use per-test.
// Both registries are pre-seeded with the full Apricot/Banff/Durango/
// Quasar block + tx type set, including the historical SkipRegistrations
// pre-amble.
//
// NOTE: this helper imports proto/p/block (which imports proto/p/txs).
// Tests inside proto/p/{block,txs} that need their own codec must
// construct it inline rather than reach for this helper (would close a
// test-time import cycle).
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

// NewMetadataCodec returns the validator/delegator metadata codec
// registered with the v0:"true" and v1:"true" tag layouts. This is a
// SEPARATE codec from the block/genesis codec — proto/p/state holds
// it as a distinct field on *state.
//
// Return type is the codec.Manager concrete (which satisfies
// state.MetadataCodec by shape) so this helper can stay independent of
// the proto/p/state package and avoid an import cycle on the
// state_test files that need this codec.
func NewMetadataCodec() codec.Manager {
	c0 := linearcodec.New([]string{metadataCodecVersion0Tag})
	c1 := linearcodec.New([]string{metadataCodecVersion0Tag, metadataCodecVersion1Tag})
	cm := codec.NewManager(math.MaxInt32)

	err := errors.Join(
		cm.RegisterCodec(metadataCodecVersion0, c0),
		cm.RegisterCodec(metadataCodecVersion1, c1),
	)
	if err != nil {
		panic(err)
	}
	return cm
}
