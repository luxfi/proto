// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"errors"

	"github.com/luxfi/proto/p/signer"
	"github.com/luxfi/proto/p/stakeable"
	"github.com/luxfi/utxo/secp256k1fx"
)

const (
	CodecVersion = 0
	Version      = CodecVersion
)

// Codec is the proto/p/txs-local wire codec interface. It is
// structurally identical to the legacy codec.Manager surface that lived
// in `github.com/luxfi/codec`, but defined locally so proto/p carries no
// import of that package. Any concrete codec (linearcodec via
// codec.Manager) satisfies this interface by shape, and callers from
// outside proto/p inject the implementation at parser construction time
// (see proto/internal/pcodectest).
//
// Wave 2A of the codec rip (#101). The shape of this interface MUST
// stay byte-compatible with codec.Manager so existing wire bytes
// continue to roundtrip during the multi-wave migration.
type Codec interface {
	Marshal(version uint16, source interface{}) ([]byte, error)
	Unmarshal(bytes []byte, dest interface{}) (uint16, error)
	Size(version uint16, value interface{}) (int, error)
}

// Registry is the proto/p/txs-local type-registration surface. Same
// rationale as Codec — structurally mirrors codec.Registry. Note that
// LinearCodec is the surface that supports SkipRegistrations needed by
// the historical PVM type layout — callers thread that concrete type in
// where order-sensitive registrations are required.
type Registry interface {
	RegisterType(interface{}) error
}

// LinearRegistry extends Registry with the slot-skipping operation used
// by the Apricot / Banff / Durango / Quasar registration sequences. The
// linearcodec implementation satisfies this; callers wire it through
// when building the wire codec.
type LinearRegistry interface {
	Registry
	SkipRegistrations(int)
}

// RegisterApricotTypes registers the type information for transactions that
// were valid during the Apricot series of upgrades.
func RegisterApricotTypes(targetCodec LinearRegistry) error {
	// The secp256k1fx is registered here because this is the same place it is
	// registered in the XVM. This ensures that the typeIDs match up for utxos
	// in shared memory.
	errs := []error{
		targetCodec.RegisterType(&secp256k1fx.TransferInput{}),
	}
	targetCodec.SkipRegistrations(1)
	errs = append(errs, targetCodec.RegisterType(&secp256k1fx.TransferOutput{}))
	targetCodec.SkipRegistrations(1)
	errs = append(errs,
		targetCodec.RegisterType(&secp256k1fx.Credential{}),
		targetCodec.RegisterType(&secp256k1fx.Input{}),
		targetCodec.RegisterType(&secp256k1fx.OutputOwners{}),

		targetCodec.RegisterType(&AddValidatorTx{}),
		targetCodec.RegisterType(&AddChainValidatorTx{}),
		targetCodec.RegisterType(&AddDelegatorTx{}),
		targetCodec.RegisterType(&CreateNetworkTx{}),
		targetCodec.RegisterType(&CreateChainTx{}),
		targetCodec.RegisterType(&ImportTx{}),
		targetCodec.RegisterType(&ExportTx{}),
		targetCodec.RegisterType(&AdvanceTimeTx{}),
		targetCodec.RegisterType(&RewardValidatorTx{}),

		targetCodec.RegisterType(&stakeable.LockIn{}),
		targetCodec.RegisterType(&stakeable.LockOut{}),
	)
	return errors.Join(errs...)
}

// RegisterBanffTypes registers the type information for transactions that were
// valid during the Banff series of upgrades.
func RegisterBanffTypes(targetCodec Registry) error {
	return errors.Join(
		targetCodec.RegisterType(&RemoveChainValidatorTx{}),
		targetCodec.RegisterType(&TransformChainTx{}),
		targetCodec.RegisterType(&AddPermissionlessValidatorTx{}),
		targetCodec.RegisterType(&AddPermissionlessDelegatorTx{}),

		targetCodec.RegisterType(&signer.Empty{}),
		targetCodec.RegisterType(&signer.ProofOfPossession{}),
	)
}

// RegisterDurangoTypes registers the type information for transactions that
// were valid during the Durango series of upgrades.
func RegisterDurangoTypes(targetCodec Registry) error {
	return errors.Join(
		targetCodec.RegisterType(&TransferChainOwnershipTx{}),
		targetCodec.RegisterType(&BaseTx{}),
	)
}

// RegisterQuasarTypes registers the type information for transactions that
// were valid during the Quasar Edition series of upgrades.
func RegisterQuasarTypes(targetCodec Registry) error {
	return errors.Join(
		targetCodec.RegisterType(&ConvertNetworkToL1Tx{}),
		targetCodec.RegisterType(&RegisterL1ValidatorTx{}),
		targetCodec.RegisterType(&SetL1ValidatorWeightTx{}),
		targetCodec.RegisterType(&IncreaseL1ValidatorBalanceTx{}),
		targetCodec.RegisterType(&DisableL1ValidatorTx{}),
	)
}

// RegisterTypes is the canonical full-history registrar — skips the
// per-block slots reserved at the head of the PVM tx codec, then calls
// Apricot, Banff, Durango, and Quasar in order. The intermediate
// SkipRegistrations(4) preserves the historical layout between the
// pre-Durango and post-Durango type IDs.
//
// Callers that need a codec that can parse any historic tx wire bytes
// invoke this. Both the runtime Codec and the GenesisCodec in
// pcodectest are seeded through this function so the runtime and
// genesis layouts stay in lock-step.
func RegisterTypes(targetCodec LinearRegistry) error {
	// Order in which type are registered affects the byte representation
	// generated by marshalling ops. To maintain codec type ordering, we
	// skip positions for the blocks.
	targetCodec.SkipRegistrations(5)

	if err := errors.Join(
		RegisterApricotTypes(targetCodec),
		RegisterBanffTypes(targetCodec),
	); err != nil {
		return err
	}

	targetCodec.SkipRegistrations(4)

	return errors.Join(
		RegisterDurangoTypes(targetCodec),
		RegisterQuasarTypes(targetCodec),
	)
}
