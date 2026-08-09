// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"bytes"
	"context"
	"errors"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/p/fx"
	"github.com/luxfi/proto/p/security"
	"github.com/luxfi/proto/p/signer"
	"github.com/luxfi/proto/p/warp/message"
	"github.com/luxfi/runtime"
	"github.com/luxfi/utils"
	"github.com/luxfi/utxo/secp256k1fx"
	"github.com/luxfi/vm/components/verify"
	"github.com/luxfi/vm/types"
)

// MaxChainAddressLength bounds the manager address a network may name.
const MaxChainAddressLength = 4096

var (
	_ UnsignedTx                        = (*CreateNetworkTx)(nil)
	_ utils.Sortable[*NetworkValidator] = (*NetworkValidator)(nil)

	ErrZeroWeight                   = errors.New("validator weight must be non-zero")
	ErrAddressTooLong               = errors.New("address is too long")
	ErrValidatorsNotSortedAndUnique = errors.New("validators must be sorted and unique")
	ErrOwnSetMustIncludeValidator   = errors.New("sovereign (non-restaking) network must include at least one genesis validator")
	ErrNoOwnSetButHasValidators     = errors.New("network with no own set must not carry validators")
	ErrContractManagerNeedsAddress  = errors.New("contract-governed own set requires a manager address")
)

// NetworkValidator is a genesis validator of a network's own validator set.
type NetworkValidator struct {
	NodeID                types.JSONByteSlice       `json:"nodeID"`
	Weight                uint64                    `json:"weight"`
	Balance               uint64                    `json:"balance"`
	Signer                signer.ProofOfPossession  `json:"signer"`
	RemainingBalanceOwner message.PChainOwner       `json:"remainingBalanceOwner"`
	DeactivationOwner     message.PChainOwner       `json:"deactivationOwner"`
}

func (v *NetworkValidator) Compare(o *NetworkValidator) int { return bytes.Compare(v.NodeID, o.NodeID) }

func (v *NetworkValidator) Verify() error {
	if v.Weight == 0 {
		return ErrZeroWeight
	}
	nodeID, err := ids.ToNodeID(v.NodeID)
	if err != nil {
		return err
	}
	if nodeID == ids.EmptyNodeID {
		return errEmptyNodeID
	}
	return verify.All(
		&v.Signer,
		&secp256k1fx.OutputOwners{Threshold: v.RemainingBalanceOwner.Threshold, Addrs: v.RemainingBalanceOwner.Addresses},
		&secp256k1fx.OutputOwners{Threshold: v.DeactivationOwner.Threshold, Addrs: v.DeactivationOwner.Addresses},
	)
}

// CreateNetworkTx is the sole network constructor — the ∅→Network birth. It
// creates a network at any level of the hierarchy in one tx:
//
//   - Parent is the parent network. Primary ⇒ an L1; an L1 ⇒ an L2; recurse.
//     The tx is byte-identical at every level — only Parent differs.
//   - Owner authorises future admin against the network record.
//   - Security is the two-axis model: restake the parent and/or run an own set
//     (Admission + Manager).
//   - ManagerChainID (+ ManagerAddress) names the chain hosting a
//     Contract-governed own set's staking contract; ids.Empty ⇒ P-Chain
//     governed, where the owner is the authority.
//
// Chains are not created here — CreateChainTx is the sole chain constructor.
type CreateNetworkTx struct {
	// Metadata, inputs and outputs
	BaseTx
	// Parent network; ids.Empty for a network anchored at the primary network
	Parent ids.ID `json:"parent"`
	// Who is authorized to manage this network
	Owner fx.Owner `json:"owner"`
	// How this network is secured
	Security security.Mode `json:"security"`
	// Genesis validators of the network's own set
	Validators []*NetworkValidator `json:"validators"`
	// Chain hosting the staking contract of a Contract-governed own set
	ManagerChainID ids.ID `json:"managerChainID"`
	// Address of that staking contract
	ManagerAddress types.JSONByteSlice `json:"managerAddress"`
}

// InitRuntime sets the FxID fields in the inputs and outputs of this
// [CreateNetworkTx].
func (tx *CreateNetworkTx) InitRuntime(rt *runtime.Runtime) {
	tx.BaseTx.InitRuntime(rt)
}

// SyntacticVerify verifies that this transaction is well-formed
func (tx *CreateNetworkTx) SyntacticVerify(rt *runtime.Runtime) error {
	switch {
	case tx == nil:
		return ErrNilTx
	case tx.SyntacticallyVerified: // already passed syntactic verification
		return nil
	}

	// enum ranges + the cross-axis invariant (restake parent || own set).
	if err := tx.Security.Valid(); err != nil {
		return err
	}
	switch {
	case !tx.Security.Sovereign() && len(tx.Validators) != 0:
		return ErrNoOwnSetButHasValidators
	case !tx.Security.RestakeParent && len(tx.Validators) == 0:
		return ErrOwnSetMustIncludeValidator
	case tx.Security.Sovereign() && tx.Security.Manager == security.Contract && len(tx.ManagerAddress) == 0:
		return ErrContractManagerNeedsAddress
	case !utils.IsSortedAndUnique(tx.Validators):
		return ErrValidatorsNotSortedAndUnique
	case len(tx.ManagerAddress) > MaxChainAddressLength:
		return ErrAddressTooLong
	}

	if err := tx.BaseTx.SyntacticVerify(rt); err != nil {
		return err
	}
	if err := tx.Owner.Verify(); err != nil {
		return err
	}
	for _, vdr := range tx.Validators {
		if err := vdr.Verify(); err != nil {
			return err
		}
	}

	tx.SyntacticallyVerified = true
	return nil
}

func (tx *CreateNetworkTx) Visit(visitor Visitor) error {
	return visitor.CreateNetworkTx(tx)
}

// InitializeWithContext initializes the transaction with consensus context
func (tx *CreateNetworkTx) InitializeWithContext(ctx context.Context) error {
	return nil
}
