// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"errors"

	"github.com/luxfi/proto/p/txs"
)

const CodecVersion = txs.CodecVersion

// Codec is the proto/p/block-local wire codec interface. It is
// structurally identical to the legacy codec.Manager surface that lived
// in `github.com/luxfi/codec`, but defined locally so proto/p carries
// no import of that package. Block consumers (state, executor, parser,
// builder) thread their concrete codec instance through this interface.
//
// Wave 2A of the codec rip (#101). Stays byte-compatible with
// codec.Manager so existing wire bytes continue to roundtrip during
// the multi-wave migration.
type Codec interface {
	Marshal(version uint16, source interface{}) ([]byte, error)
	Unmarshal(bytes []byte, dest interface{}) (uint16, error)
	Size(version uint16, value interface{}) (int, error)
}

// Registry mirrors the legacy codec.Registry / linearcodec.Codec
// surface needed by the block-type registrar.
type Registry interface {
	RegisterType(interface{}) error
}

// LinearRegistry extends Registry with SkipRegistrations so the
// txs.RegisterTypes pre-amble (which reserves head slots for blocks)
// can fan in.
type LinearRegistry interface {
	Registry
	SkipRegistrations(int)
}

// RegisterApricotTypes registers the type information for blocks that were
// valid during the Apricot series of upgrades.
func RegisterApricotTypes(targetCodec LinearRegistry) error {
	return errors.Join(
		targetCodec.RegisterType(&ApricotProposalBlock{}),
		targetCodec.RegisterType(&ApricotAbortBlock{}),
		targetCodec.RegisterType(&ApricotCommitBlock{}),
		targetCodec.RegisterType(&ApricotStandardBlock{}),
		targetCodec.RegisterType(&ApricotAtomicBlock{}),
		txs.RegisterApricotTypes(targetCodec),
	)
}

// RegisterBanffTypes registers the type information for blocks that were valid
// during the Banff series of upgrades.
func RegisterBanffTypes(targetCodec LinearRegistry) error {
	return errors.Join(
		txs.RegisterBanffTypes(targetCodec),
		targetCodec.RegisterType(&BanffProposalBlock{}),
		targetCodec.RegisterType(&BanffAbortBlock{}),
		targetCodec.RegisterType(&BanffCommitBlock{}),
		targetCodec.RegisterType(&BanffStandardBlock{}),
	)
}

// RegisterDurangoTypes registers the type information for blocks that were
// valid during the Durango series of upgrades.
func RegisterDurangoTypes(targetCodec LinearRegistry) error {
	return txs.RegisterDurangoTypes(targetCodec)
}

// RegisterQuasarTypes registers the type information for blocks that were valid
// during the Quasar Edition series of upgrades.
func RegisterQuasarTypes(targetCodec LinearRegistry) error {
	return txs.RegisterQuasarTypes(targetCodec)
}

// RegisterTypes is the canonical full-history block-and-tx registrar.
// It seeds Apricot, Banff, Durango, and Quasar block + tx types in the
// order required by the historical PVM layout. Callers that need a
// codec capable of decoding any historic PVM block invoke this. The
// caller is responsible for additionally registering any state-only
// types (e.g. state.stateBlk for legacy block storage) on the genesis
// codec via RegisterGenesisStateBlockType.
func RegisterTypes(targetCodec LinearRegistry) error {
	return errors.Join(
		RegisterApricotTypes(targetCodec),
		RegisterBanffTypes(targetCodec),
		RegisterDurangoTypes(targetCodec),
		RegisterQuasarTypes(targetCodec),
	)
}
