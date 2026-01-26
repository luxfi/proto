// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"github.com/luxfi/runtime"
	"github.com/luxfi/ids"
	"github.com/luxfi/math/set"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"
)

// ContextInitializable defines the interface for initializing context
type ContextInitializable interface {
	InitRuntime(rt *runtime.Runtime)
}

// UnsignedTx is an unsigned transaction
type UnsignedTx interface {
	// TODO: Remove this initialization pattern from both the platformvm and the
	// avm.
	ContextInitializable
	secp256k1fx.UnsignedTx
	SetBytes(unsignedBytes []byte)

	// InputIDs returns the set of inputs this transaction consumes
	InputIDs() set.Set[ids.ID]

	Outputs() []*lux.TransferableOutput

	// Attempts to verify this transaction without any provided state.
	SyntacticVerify(rt *runtime.Runtime) error

	// Visit calls [visitor] with this transaction's concrete type
	Visit(visitor Visitor) error
}
