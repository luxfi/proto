// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package pcodectest is the canonical test wiring for the proto/p PVM
// codec set. proto/p carries no github.com/luxfi/codec import after the
// Wave 2A rip (#101); this helper package is the bridge that lets test
// suites under proto/p construct ZAP-native codecs without duplicating
// wire-registration logic across every test file.
//
// Production callers (luxfi/node/vms/platformvm/...) construct their
// codecs inline via proto/zap_codec. This helper exists strictly so
// the in-tree test files don't need to duplicate the wiring.
//
// This file holds the codec-agnostic helpers — generic Codec / Manager
// construction, sentinel error re-exports, and shared type aliases.
// PVM-specific helpers (NewPVMCodecs, NewPVMRuntimeCodec,
// NewMetadataCodec, NewWarpCodec, NewMessageCodec, NewPayloadCodec) live
// in pvmcodectest and import proto/p/{block,txs,warp,...}. The split
// keeps this file importable from tests in those same proto/p packages
// without closing a test-time import cycle.
//
// Wire format is ZAP-native (little-endian) — proto/zap_codec is the
// single canonical construction site for the wire codec choice (LP-023).
package pcodectest

import (
	"math"

	"github.com/luxfi/proto/zap_codec"
)

// ErrCantUnpackVersion is re-exported from proto/zap_codec so test
// files under proto/p can assert on the codec's "missing version byte"
// sentinel without importing zap_codec themselves.
var ErrCantUnpackVersion = zap_codec.ErrCantUnpackVersion

// VersionSize re-exports zap_codec.VersionSize — the on-wire length of
// the codec-version prefix (2 bytes). Tests that compute expected
// Marshal-output sizes subtract this constant to isolate the payload
// component.
const VersionSize = zap_codec.VersionSize

// ErrInsufficientLength is re-exported from proto/zap_codec for the
// same reason — proto/p test files can `require.ErrorIs(err,
// pcodectest.ErrInsufficientLength)` without picking up an upstream
// codec import directly.
var ErrInsufficientLength = zap_codec.ErrInsufficientLength

// LinearCodec re-exports zap_codec.LinearCodec (the union of Registry
// + Codec + SkipRegistrations) so test files can hold a concrete codec
// without importing proto/zap_codec directly. Each call to
// NewLinearCodec returns its own fresh codec — no global state.
//
// Despite the historical "Linear" name, this codec is backed by
// luxfi/codec/zapcodec — little-endian wire bytes. The name refers to
// LINEAR TYPE-ID ASSIGNMENT (sequential ids assigned in registration
// order), not to the legacy linearcodec big-endian wire format.
type LinearCodec = zap_codec.LinearCodec

// CodecManager re-exports zap_codec.MultiManager so test files can
// hold a concrete multi-version manager without importing
// proto/zap_codec directly. Used in test files that need to register
// a custom codec-versions table.
type CodecManager = zap_codec.MultiManager

// NewLinearCodec returns a fresh ZAP-native Codec instance. Used by:
//   - fx-style tests that need a `Registry` parameter for
//     `fx.Initialize(...)` but never exercise the wire codec.
//   - fuzz tests that exercise MarshalInto / UnmarshalFrom directly
//     against a Packer.
//   - in-package _test.go files that need to build their own codec
//     against package-internal RegisterTypes (the helper here can't
//     import the package because that would close a test-time import
//     cycle through proto/p/block / proto/p/txs).
func NewLinearCodec() LinearCodec {
	return zap_codec.NewLinearCodec()
}

// NewLinearCodecWithTags returns a fresh ZAP-native Codec instance
// that honours the supplied struct-tag names. Used by the metadata
// codec wiring where v0:"true" / v1:"true" tags select per-version
// field sets.
func NewLinearCodecWithTags(tags ...string) LinearCodec {
	return zap_codec.NewLinearCodecWithTags(tags...)
}

// NewCodecManager returns a fresh multi-version Manager with the
// supplied max wire-payload size. Tests use this to register their
// own codec versions against a Manager.
func NewCodecManager(maxSize uint64) CodecManager {
	return zap_codec.NewManager(maxSize)
}

// NewDefaultCodecManager returns a fresh multi-version Manager with
// the default max wire-payload size (1 MiB).
func NewDefaultCodecManager() CodecManager {
	return zap_codec.NewDefaultManager()
}

// MaxInt32Manager returns a fresh multi-version Manager sized for
// genesis-style blobs (math.MaxInt32 budget). Tests building genesis
// codecs reach for this rather than re-deriving the budget at every
// call site.
func MaxInt32Manager() CodecManager {
	return zap_codec.NewMaxInt32Manager()
}

// MaxIntManager returns a fresh multi-version Manager sized for warp-
// style payloads (math.MaxInt budget — effectively unbounded).
func MaxIntManager() CodecManager {
	return zap_codec.NewManager(uint64(math.MaxInt))
}

// Errs re-exports zap_codec.Errs (a multi-error accumulator the legacy
// platformvm tests use to fold multiple ErrorIs targets into a single
// require.NoError chain). proto/p test files use the type alias to
// keep the codec/wrappers import out of their import block.
type Errs = zap_codec.Errs

// Packer re-exports zap_codec.Packer so fuzz tests can build/read
// MarshalInto / UnmarshalFrom byte streams without importing the
// codec/wrappers subpackage directly. The fuzz tests need the raw
// Packer because they're exercising MarshalInto codepaths that don't
// go through Manager.Marshal — pcodectest is the bridge.
type Packer = zap_codec.Packer
