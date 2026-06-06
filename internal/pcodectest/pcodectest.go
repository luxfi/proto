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
//
// This file holds the codec-agnostic helpers — generic linearcodec /
// codec.Manager construction, sentinel error re-exports, and shared
// type aliases. PVM-specific helpers (NewPVMCodecs, NewPVMRuntimeCodec,
// NewMetadataCodec, NewWarpCodec, NewMessageCodec, NewPayloadCodec) live
// in pvm.go and import proto/p/{block,txs,warp,...}. The split keeps
// this file importable from tests in those same proto/p packages
// without closing a test-time import cycle.
package pcodectest

import (
	"math"

	"github.com/luxfi/codec"
	"github.com/luxfi/codec/linearcodec"
	"github.com/luxfi/codec/wrappers"
)

// ErrCantUnpackVersion is re-exported from luxfi/codec so test files
// under proto/p can assert on the codec's "missing version byte"
// sentinel without importing luxfi/codec themselves.
var ErrCantUnpackVersion = codec.ErrCantUnpackVersion

// VersionSize re-exports codec.VersionSize — the on-wire length of the
// codec-version prefix (2 bytes). Tests that compute expected
// Marshal-output sizes subtract this constant to isolate the payload
// component; they reach for the re-export here to avoid importing
// luxfi/codec directly.
const VersionSize = codec.VersionSize

// ErrInsufficientLength is re-exported from luxfi/codec/wrappers for
// the same reason — proto/p test files can `require.ErrorIs(err,
// pcodectest.ErrInsufficientLength)` without picking up the wrappers
// import directly.
var ErrInsufficientLength = wrappers.ErrInsufficientLength

// LinearCodec re-exports linearcodec.Codec (the union of codec.Registry
// + codec.Codec) so test files can hold a concrete linear codec without
// importing luxfi/codec/linearcodec directly. Each call to
// NewLinearCodec returns its own fresh codec — no global state.
type LinearCodec = linearcodec.Codec

// CodecManager re-exports codec.Manager so test files can hold a
// concrete codec.Manager without importing luxfi/codec directly. Used
// in test files that need to register a custom codec versions table.
type CodecManager = codec.Manager

// NewLinearCodec returns a fresh linearcodec-backed Codec instance.
// Used by:
//   - fx-style tests that need a `codec.Registry` parameter for
//     `fx.Initialize(...)` but never exercise the wire codec.
//   - fuzz tests that exercise MarshalInto / UnmarshalFrom directly
//     against a Packer.
//   - in-package _test.go files that need to build their own codec
//     against package-internal RegisterTypes (the helper here can't
//     import the package because that would close a test-time import
//     cycle through proto/p/block / proto/p/txs).
func NewLinearCodec() LinearCodec {
	return linearcodec.NewDefault()
}

// NewLinearCodecWithTags returns a fresh linearcodec.Codec with the
// supplied struct-tag names. Mirrors linearcodec.New([]string{tag...}).
// Used by the metadata codec wiring where v0:"true" / v1:"true" tags
// select per-version field sets.
func NewLinearCodecWithTags(tags ...string) LinearCodec {
	return linearcodec.New(tags)
}

// NewCodecManager returns a fresh codec.Manager with the supplied max
// wire-payload size. Mirrors codec.NewManager(maxSize). Tests use this
// to register their own codec versions against a Manager.
func NewCodecManager(maxSize uint64) CodecManager {
	return codec.NewManager(maxSize)
}

// NewDefaultCodecManager returns a fresh codec.Manager with the
// default max wire-payload size. Mirrors codec.NewDefaultManager().
func NewDefaultCodecManager() CodecManager {
	return codec.NewDefaultManager()
}

// MaxInt32Manager returns a fresh codec.Manager sized for genesis-style
// blobs (math.MaxInt32 budget). Tests building genesis codecs reach for
// this rather than re-deriving the budget at every call site.
func MaxInt32Manager() CodecManager {
	return codec.NewManager(math.MaxInt32)
}

// MaxIntManager returns a fresh codec.Manager sized for warp-style
// payloads (math.MaxInt budget — effectively unbounded).
func MaxIntManager() CodecManager {
	return codec.NewManager(uint64(math.MaxInt))
}

// Errs re-exports wrappers.Errs (a multi-error accumulator the legacy
// platformvm tests use to fold multiple ErrorIs targets into a single
// require.NoError chain). proto/p test files use the type alias to
// keep the codec/wrappers import out of their import block.
type Errs = wrappers.Errs

// Packer re-exports wrappers.Packer so fuzz tests can build/read
// MarshalInto / UnmarshalFrom byte streams without importing the
// codec/wrappers subpackage directly. The fuzz tests need the raw
// Packer because they're exercising codec.MarshalInto codepaths that
// don't go through Manager.Marshal — pcodectest is the bridge.
type Packer = wrappers.Packer
