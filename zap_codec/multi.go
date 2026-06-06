// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_codec

import (
	"math"

	"github.com/luxfi/utils/wrappers"
	"github.com/luxfi/zapcodec"
)

// Codec is the local wire-codec interface — Marshal/Unmarshal/Size on
// a *wrappers.Packer. Structurally identical to the legacy
// github.com/luxfi/codec.Codec interface; every zapcodec.Codec instance
// satisfies it by embedding the same three methods. Consumers of
// proto/zap_codec hold this local type rather than importing the
// upstream codec package.
type Codec interface {
	MarshalInto(interface{}, *wrappers.Packer) error
	UnmarshalFrom(*wrappers.Packer, interface{}) error
	Size(value interface{}) (int, error)
}

// Registry is the local type-registration interface. Structurally
// identical to github.com/luxfi/codec.Registry. Every zapcodec.Codec
// instance satisfies it via the embedded zapcodec.RegisterType.
type Registry interface {
	RegisterType(interface{}) error
}

// LinearRegistry extends Registry with the slot-skipping operation used
// by historical PVM/XVM type layouts. zapcodec.Codec satisfies this
// (its SkipRegistrations bumps the next-type-id counter).
type LinearRegistry interface {
	Registry
	SkipRegistrations(int)
}

// GeneralCodec is the Codec + Registry union — every concrete codec
// instance the platform uses (currently zapcodec) satisfies it.
type GeneralCodec interface {
	Codec
	Registry
}

// LinearCodec is the GeneralCodec + SkipRegistrations surface every VM
// consumer holds. The name "Linear" refers to LINEAR TYPE-ID
// ASSIGNMENT (sequential ids assigned in registration order), NOT to
// the historical linearcodec big-endian wire format. The underlying
// implementation is luxfi/codec/zapcodec — little-endian on the wire.
type LinearCodec = zapcodec.Codec

// ZAPCodec is an explicit alias for the same zapcodec.Codec surface
// callers reach for when they want to emphasise the ZAP-native wire
// layout (e.g. platformvm txs V2). Same backing type as LinearCodec;
// the alias exists for readability at call sites.
type ZAPCodec = zapcodec.Codec

// Packer re-exports wrappers.Packer for callers that need to drive
// MarshalInto / UnmarshalFrom byte streams directly (fuzz tests,
// size-accounting code).
type Packer = wrappers.Packer

// Errs re-exports wrappers.Errs — a multi-error accumulator used by
// per-VM codec.go files that fan in many RegisterCodec / RegisterType
// calls under a single error tap.
type Errs = wrappers.Errs

// ErrInsufficientLength is the zapcodec underflow sentinel. Callers
// asserting on the codec's "read past buffer end" condition do so via
// this re-export so they don't need to import the zapcodec subpackage.
//
// IMPORTANT: this points to zapcodec.ErrInsufficientLength, not
// wrappers.ErrInsufficientLength. The zapcodec internal packer raises
// its own sentinel which is propagated unchanged through
// Manager.Unmarshal — assertions against wrappers.ErrInsufficientLength
// would not match the chain.
var ErrInsufficientLength = zapcodec.ErrInsufficientLength

// Byte-length constants for the wire-size accounting code in node/vms.
// Identical values to wrappers.{Byte,Short,Int,Long,Bool}Len — re-
// exported here so consumers don't reach across into luxfi/codec.
const (
	ByteLen  = wrappers.ByteLen
	ShortLen = wrappers.ShortLen
	IntLen   = wrappers.IntLen
	LongLen  = wrappers.LongLen
	BoolLen  = wrappers.BoolLen
)

// DefaultMaxSize is the default wire-payload budget when the caller
// doesn't size their Manager explicitly. Matches the legacy
// codec.DefaultMaxSize (1 MiB).
const DefaultMaxSize = 1024 * 1024

// NewLinearCodec returns a fresh zapcodec-backed Codec instance with
// the default ("serialize") struct tag. The name "Linear" refers to
// LINEAR TYPE-ID ASSIGNMENT (sequential ids), NOT to the historical
// linearcodec big-endian wire layout — this codec emits LE wire bytes.
func NewLinearCodec() LinearCodec {
	return zapcodec.NewDefault()
}

// NewLinearCodecWithTags returns a fresh zapcodec-backed Codec instance
// that honours the supplied struct-tag names. Used by the metadata
// codec wiring where v0:"true" / v1:"true" tags select per-version
// field sets.
func NewLinearCodecWithTags(tags ...string) LinearCodec {
	return zapcodec.New(tags)
}

// NewZAPCodec is an alias for NewLinearCodec — both return a
// zapcodec-backed Codec. Call sites that want to emphasise the ZAP
// wire layout use this name; otherwise they're interchangeable.
func NewZAPCodec() ZAPCodec {
	return zapcodec.NewDefault()
}

// NewDefaultManager returns a multi-version Manager with the default
// 1 MiB wire-payload budget. Equivalent to NewManager(DefaultMaxSize).
func NewDefaultManager() MultiManager {
	return NewManager(DefaultMaxSize)
}

// NewManager returns a multi-version Manager with the supplied wire-
// payload budget. Versions are registered after construction via
// MultiManager.RegisterCodec.
func NewManager(maxSize uint64) MultiManager {
	size := maxSize
	if size > math.MaxInt {
		size = math.MaxInt
	}
	return &multiManager{
		maxSize: int(size),
		codecs:  make(map[uint16]Codec),
	}
}

// NewMaxInt32Manager returns a multi-version Manager sized for
// genesis-style blobs (math.MaxInt32 budget). VM genesis codec wiring
// reaches for this rather than re-deriving the budget at every call
// site.
func NewMaxInt32Manager() MultiManager {
	return NewManager(math.MaxInt32)
}

// NewMaxIntManager returns a multi-version Manager sized for warp /
// proposervm-block style payloads (math.MaxInt budget — effectively
// unbounded; the p2p layer caps real-world wire sizes well below this).
func NewMaxIntManager() MultiManager {
	return NewManager(math.MaxInt)
}

// MultiManager is a multi-codec-version dispatcher. It owns the
// version-prefix outer layer and routes Marshal/Unmarshal to the
// registered inner Codec for the supplied version. Wire layout:
//
//	[uint16 LE codec version] [inner codec body]
//
// The version-prefix byte order is LE to match the inner zapcodec body
// byte order. This deliberately diverges from the legacy
// luxfi/codec.Manager (which prepends BE) — see proto/zap_codec/README.md
// for the wire-format break note.
//
// Thread-safety: RegisterCodec is NOT goroutine-safe — call it once at
// init time before any Marshal/Unmarshal. After that, Marshal /
// Unmarshal / Size are concurrency-safe to the extent the inner Codec
// implementations are (zapcodec.Codec is, via its RWMutex-protected
// type registry).
type MultiManager interface {
	RegisterCodec(version uint16, codec Codec) error
	Marshal(version uint16, source interface{}) ([]byte, error)
	Unmarshal(bytes []byte, dest interface{}) (uint16, error)
	Size(version uint16, value interface{}) (int, error)
}

type multiManager struct {
	maxSize int
	codecs  map[uint16]Codec
}

// RegisterCodec registers an inner Codec at the supplied version.
// Returns ErrUnknownVersion if the version is already registered (the
// upstream codec.Manager returns codec.ErrDuplicateType for this case;
// we surface ErrUnknownVersion to match the rest of the package's
// sentinel choices, but both indicate the same invariant violation).
func (m *multiManager) RegisterCodec(version uint16, codec Codec) error {
	if _, exists := m.codecs[version]; exists {
		return ErrUnknownVersion
	}
	m.codecs[version] = codec
	return nil
}

// Marshal serializes source under the codec registered at the supplied
// version. The wire bytes start with the version (uint16 LE), followed
// by the inner codec body.
func (m *multiManager) Marshal(version uint16, source interface{}) ([]byte, error) {
	c, exists := m.codecs[version]
	if !exists {
		return nil, ErrUnknownVersion
	}
	p := &wrappers.Packer{MaxSize: m.maxSize}
	if err := writeVersionLE(p, version); err != nil {
		return nil, err
	}
	if err := c.MarshalInto(source, p); err != nil {
		return nil, err
	}
	if p.Err != nil {
		return nil, p.Err
	}
	return p.Bytes[:p.Offset], nil
}

// Unmarshal deserializes bytes into dest using the inner codec
// registered for the version embedded in the wire prefix. Returns the
// version read from the wire, even on error (so callers can log the
// version that failed). Rejects buffers larger than maxSize, version
// not registered, and trailing bytes past the end of the inner body.
func (m *multiManager) Unmarshal(bytes []byte, dest interface{}) (uint16, error) {
	if len(bytes) < VersionSize {
		return 0, ErrCantUnpackVersion
	}
	if len(bytes) > m.maxSize {
		return 0, ErrMaxSizeExceeded
	}
	p := &wrappers.Packer{Bytes: bytes, MaxSize: m.maxSize}
	version, err := readVersionLE(p)
	if err != nil {
		return 0, err
	}
	c, exists := m.codecs[version]
	if !exists {
		// LP-023 cutover compat: pre-Dec-25-2025 writers used BE for
		// the version prefix. If LE decode yields an unregistered
		// version, retry as BE before erroring. Registered version IDs
		// are small (0..N for a few codec generations) — BE-decoded
		// small numbers don't collide with LE-decoded small numbers
		// except at palindromic byte pairs (where the two decodings
		// agree).
		beVersion := peekVersionBE(bytes)
		if cBE, beExists := m.codecs[beVersion]; beExists {
			p.Offset = VersionSize
			if err := cBE.UnmarshalFrom(p, dest); err != nil {
				return beVersion, err
			}
			if p.Offset != len(bytes) {
				return beVersion, ErrExtraSpace
			}
			return beVersion, nil
		}
		return version, ErrUnknownVersion
	}
	if err := c.UnmarshalFrom(p, dest); err != nil {
		return version, err
	}
	if p.Offset != len(bytes) {
		return version, ErrExtraSpace
	}
	return version, nil
}

// Size returns the on-wire size of value under the codec registered at
// the supplied version, INCLUDING the manager's version-prefix bytes.
func (m *multiManager) Size(version uint16, value interface{}) (int, error) {
	c, exists := m.codecs[version]
	if !exists {
		return 0, ErrUnknownVersion
	}
	size, err := c.Size(value)
	if err != nil {
		return 0, err
	}
	return VersionSize + size, nil
}
