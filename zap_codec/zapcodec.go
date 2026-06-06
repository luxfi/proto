// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package zap_codec is the single canonical construction site for the
// ZAP-native wire codecs consumed by proto/p, proto/x, proto/p/warp,
// proto/p/warp/payload, proto/p/warp/message, proto/p/block, and
// proto/x/block. It exists so the SDK wallet (sdk/wallet/chain/{p,x})
// can construct its codec dependencies without importing
// github.com/luxfi/codec at the wallet layer — the codec choice is
// bound EXACTLY ONCE in this package.
//
// Wire-format relationship to linearcodec:
//
//   - This package's codec backend is luxfi/codec/zapcodec, which is the
//     little-endian sibling of linearcodec. Reflection layout (struct-tag
//     order, slice/map length prefixes, interface type-id mechanism) is
//     identical to linearcodec; the only delta is byte order:
//
//       - linearcodec writes BIG-endian for uint16/uint32/uint64 fields,
//         length prefixes, interface type-ids, codec version bytes,
//         string length prefixes (uint16 BE).
//
//       - zapcodec writes LITTLE-endian for the same set of fields.
//         x86_64 and arm64 are LE-native, so LE writes map to single MOV
//         instructions where BE writes need BSWAP.
//
//     This is a WIRE-FORMAT BREAK from the legacy linearcodec wire bytes
//     that previously fed proto/{p,x} consumers. The break is intentional
//     and aligned with LP-023 ZAP-native activation
//     (proto/zap_native/codec_select.go: ZAPActivationUnix=0 means "ZAP
//     is mandatory from genesis"). Forward-only — there is no dual-mode
//     emission and no rollback path. Network validators expecting LE
//     wire bytes will accept this codec's output; validators still on
//     BE linearcodec will not. See README.md in this package for a hex
//     dump of a canonical signed BaseTx in both modes.
//
// Architecture (Hickey decomplection):
//
//   - VALUE: the wire codec choice. Today: zapcodec LE. (Was: linearcodec
//     BE.) The value lives here, qualified by namespace, not by braided
//     prefix in every consumer.
//
//   - COMPOSITION: NewVersionedManager wraps a zapcodec.Codec with a
//     version-prefix outer layer. The outer layer carries the codec
//     version (uint16 LE, matching the inner codec's byte order) and
//     dispatches to the registered inner codec on Unmarshal. Outer +
//     inner are independently complete primitives.
//
//   - ORTHOGONAL: this package neither imports proto/{p,x} nor knows
//     about block/tx/warp types. Per-chain type registration happens at
//     the caller site (sdk/wallet/chain/{p/pcodecs.go,x/constants.go,
//     x/builder/constants.go}) which RegisterTypes onto the registries
//     returned by NewPVMRuntime / NewPVMGenesis / NewXVMParser /
//     NewWarp / NewPayload / NewMessage.
package zap_codec

import (
	"errors"
	"math"

	"github.com/luxfi/utils/wrappers"
	"github.com/luxfi/zapcodec"
)

// MaxSize is the default maximum wire-payload size for runtime txs.
// Mirrors codec.DefaultMaxSize (1 MiB) so per-tx envelope is unchanged
// vs linearcodec-backed configurations.
const MaxSize = 1024 * 1024

// VersionSize is the on-wire length of the codec-version prefix the
// outer manager prepends. Two bytes, uint16 little-endian.
const VersionSize = 2

// Sentinel errors. Mirror the shape of luxfi/codec errors so callers
// that previously asserted on codec.ErrUnknownVersion / .ErrCantUnpackVersion
// continue to work via errors.Is once they switch to errors.Is on these
// sentinels.
var (
	ErrCantPackVersion           = errors.New("zap_codec: couldn't pack codec version")
	ErrCantUnpackVersion         = errors.New("zap_codec: couldn't unpack codec version")
	ErrUnknownVersion            = errors.New("zap_codec: unknown codec version")
	ErrExtraSpace                = errors.New("zap_codec: trailing buffer space")
	ErrMaxSizeExceeded           = errors.New("zap_codec: wire payload exceeds max size")
	ErrMaxSliceLenExceeded       = errors.New("zap_codec: max slice length exceeded")
	ErrMarshalNil                = errors.New("zap_codec: can't marshal nil pointer")
	ErrUnmarshalNil              = errors.New("zap_codec: can't unmarshal into nil")
	ErrDoesNotImplementInterface = errors.New("zap_codec: does not implement interface")
)

// Manager is the version-prefix wire codec implementation. It owns the
// codec.Manager-shaped surface (Marshal/Unmarshal/Size) that proto/{p,x}'s
// local Codec interfaces structurally match, but contains zero
// reference to github.com/luxfi/codec.Manager — Marshal/Unmarshal/Size
// are implemented directly on top of a zapcodec.Codec.
//
// Manager also satisfies zapcodec.Codec's Registry surface by delegating
// RegisterType / SkipRegistrations to the inner codec. This is the same
// contract every Wave 2-era consumer expects: hand back ONE value that
// is BOTH the wire codec AND the type registry.
//
// Thread-safety: the inner zapcodec.Codec is concurrency-safe (its
// RWMutex protects the type registry). Manager itself is stateless past
// construction — no per-call mutations on the *Manager pointer — so
// callers may share a *Manager across goroutines without external sync.
type Manager struct {
	inner   zapcodec.Codec
	version uint16
	maxSize int
}

// NewVersionedManager constructs a Manager that emits/accepts wire
// bytes prefixed by a uint16 little-endian codec version. The inner
// zapcodec instance is allocated fresh; SkipRegistrations and
// RegisterType go directly to it.
//
// `version` is the codec version the manager will write on Marshal and
// require on Unmarshal. `maxSize` is the post-version-prefix wire
// budget (Marshal refuses to emit, Unmarshal refuses to read, anything
// past this size).
func NewVersionedManager(version uint16, maxSize uint64) *Manager {
	size := maxSize
	if size > math.MaxInt {
		size = math.MaxInt
	}
	return &Manager{
		inner:   zapcodec.NewDefault(),
		version: version,
		maxSize: int(size),
	}
}

// RegisterType registers val with the inner reflection codec so it can
// be unmarshalled into an interface field. Satisfies the
// proto/{p,x,block,warp}.Registry interface.
func (m *Manager) RegisterType(val interface{}) error {
	return m.inner.RegisterType(val)
}

// SkipRegistrations bumps the inner codec's next-type-id by n. Lets
// callers preserve legacy type-id slot layouts even across the linear→
// zap codec backend swap.
func (m *Manager) SkipRegistrations(n int) {
	m.inner.SkipRegistrations(n)
}

// Marshal serializes source into a fresh byte slice with the manager's
// codec version (uint16 LE) prepended. The caller-supplied version MUST
// match the manager's bound version — any other value returns
// ErrUnknownVersion.
//
// Satisfies the proto/{p,x,block,warp}.Codec interface.
func (m *Manager) Marshal(version uint16, source interface{}) ([]byte, error) {
	if version != m.version {
		return nil, ErrUnknownVersion
	}
	p := &wrappers.Packer{MaxSize: m.maxSize}
	// Match zapcodec's LE byte order for the version prefix too.
	// wrappers.PackShort is BE — bypass it and write LE directly.
	if err := writeVersionLE(p, version); err != nil {
		return nil, err
	}
	if err := m.inner.MarshalInto(source, p); err != nil {
		return nil, err
	}
	if p.Err != nil {
		return nil, p.Err
	}
	return p.Bytes[:p.Offset], nil
}

// Unmarshal deserializes bytes into dest. Returns the codec version
// read from the wire prefix. Rejects buffers smaller than VersionSize,
// buffers larger than maxSize, version mismatches, and trailing bytes.
//
// Satisfies the proto/{p,x,block,warp}.Codec interface.
func (m *Manager) Unmarshal(bytes []byte, dest interface{}) (uint16, error) {
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
	if version != m.version {
		return version, ErrUnknownVersion
	}
	if err := m.inner.UnmarshalFrom(p, dest); err != nil {
		return version, err
	}
	if p.Offset != len(bytes) {
		return version, ErrExtraSpace
	}
	return version, nil
}

// Size returns the on-wire size of value INCLUDING the manager's
// version-prefix bytes.
//
// Satisfies the proto/{p,x,block,warp}.Codec interface.
func (m *Manager) Size(version uint16, value interface{}) (int, error) {
	if version != m.version {
		return 0, ErrUnknownVersion
	}
	size, err := m.inner.Size(value)
	if err != nil {
		return 0, err
	}
	return VersionSize + size, nil
}

// NewPVMRuntime returns a Manager wired for proto/p PVM runtime
// transactions. Caller is responsible for registering block/tx types via
// proto/p/block.RegisterTypes(m) before any Marshal/Unmarshal.
//
// Budget: MaxSize (1 MiB). For genesis blobs use NewPVMGenesis.
func NewPVMRuntime(version uint16) *Manager {
	return NewVersionedManager(version, MaxSize)
}

// NewPVMGenesis returns a Manager wired for proto/p PVM genesis blobs.
// Same wire layout as NewPVMRuntime but with math.MaxInt32 size budget
// (genesis can be very large — full validator set + every chain's
// initial state).
func NewPVMGenesis(version uint16) *Manager {
	return NewVersionedManager(version, math.MaxInt32)
}

// NewXVMParser returns the four codec/registry values proto/x's
// txs.NewParser / block.NewParser require: runtime Codec, genesis
// Codec, runtime Registry, genesis Registry.
//
// Each returned Manager is independent (Marshal/Unmarshal on the
// runtime manager does NOT touch the genesis manager's type-id table).
// Caller is responsible for routing fx-owned and block-level type
// registrations onto BOTH registries — see proto/x/txs/parser.go and
// proto/x/block/parser.go for the canonical wiring.
//
// runtimeMaxSize / genesisMaxSize budget the two managers separately.
// Pass MaxSize for the runtime manager and math.MaxInt32 for the
// genesis manager to match the historical linearcodec wiring.
func NewXVMParser(version uint16) (runtime, genesis *Manager) {
	return NewVersionedManager(version, MaxSize),
		NewVersionedManager(version, math.MaxInt32)
}

// NewWarp returns a Manager wired for proto/p/warp signature + teleport
// payload types. Budget is math.MaxInt because warp signature
// aggregates may grow with validator set size.
//
// Caller is responsible for warp.RegisterTypes(m) before any
// Marshal/Unmarshal.
func NewWarp(version uint16) *Manager {
	return NewVersionedManager(version, math.MaxInt)
}

// NewPayload returns a Manager wired for proto/p/warp/payload types
// (Hash, AddressedCall). Budget is payloadMaxSize.
//
// Caller is responsible for payload.RegisterTypes(m) before any
// Marshal/Unmarshal.
func NewPayload(version uint16, payloadMaxSize uint64) *Manager {
	return NewVersionedManager(version, payloadMaxSize)
}

// NewMessage returns a Manager wired for proto/p/warp/message types.
// Budget is math.MaxInt — warp messages embed AddressedCall payloads
// whose size is bounded by the payload codec, not this one.
//
// Caller is responsible for message.RegisterTypes(m) before any
// Marshal/Unmarshal.
func NewMessage(version uint16) *Manager {
	return NewVersionedManager(version, math.MaxInt)
}

// writeVersionLE writes a uint16 codec version in little-endian. The
// wrappers.Packer's built-in PackShort is BIG-endian — we cannot reuse
// it because the zapcodec body that follows is LE.
//
// Writes two PackByte calls so no per-Marshal slice allocation is
// needed for the version prefix. Returns ErrCantPackVersion if the
// packer ran out of room (e.g. MaxSize too small).
func writeVersionLE(p *wrappers.Packer, version uint16) error {
	p.PackByte(byte(version))
	p.PackByte(byte(version >> 8))
	if p.Errored() {
		return ErrCantPackVersion
	}
	return nil
}

// readVersionLE reads a uint16 codec version in little-endian. Caller
// has already validated len(bytes) >= VersionSize. Uses two UnpackByte
// calls so the read path has no per-Unmarshal slice allocation.
func readVersionLE(p *wrappers.Packer) (uint16, error) {
	lo := p.UnpackByte()
	hi := p.UnpackByte()
	if p.Errored() {
		return 0, ErrCantUnpackVersion
	}
	return uint16(lo) | uint16(hi)<<8, nil
}
