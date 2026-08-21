// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"errors"

	"github.com/luxfi/constants"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/p/security"
	"github.com/luxfi/runtime"
	"github.com/luxfi/vm/components/verify"
	"github.com/luxfi/vm/types"
)

var (
	_ UnsignedTx = (*ConvertNetworkToL1Tx)(nil)

	ErrConvertPermissionlessChain          = errors.New("cannot convert a permissionless chain")
	ErrConvertMustIncludeValidators        = errors.New("conversion must include at least one validator")
	ErrConvertValidatorsNotSortedAndUnique = errors.New("conversion validators must be sorted and unique")
)

// ConvertNetworkToL1Validator is a genesis validator of the promoted network's
// own set. Same value as a network's birth validator — one definition.
type ConvertNetworkToL1Validator = NetworkValidator

// ConvertNetworkToL1Tx promotes an existing network: inherited security to
// sovereign, re-anchored on a new parent.
type ConvertNetworkToL1Tx struct {
	// Metadata, inputs and outputs
	BaseTx `serialize:"true"`
	// ID of the network being promoted
	Chain ids.ID `serialize:"true" json:"chainID"`
	// New parent network
	Parent ids.ID `json:"parent"`
	// Blockchain where the network manager lives
	ManagerChainID ids.ID `serialize:"true" json:"managerChainID"`
	// Address of the network manager
	Address types.JSONByteSlice `serialize:"true" json:"address"`
	// Initial pay-as-you-go validators for the network
	Validators []*ConvertNetworkToL1Validator `serialize:"true" json:"validators"`
	// Authorizes this conversion
	ChainAuth verify.Verifiable `serialize:"true" json:"chainAuthorization"`
	// How the promoted network is secured
	Security security.Mode `json:"security"`
}

func (tx *ConvertNetworkToL1Tx) SyntacticVerify(rt *runtime.Runtime) error {
	switch {
	case tx == nil:
		return ErrNilTx
	case tx.SyntacticallyVerified:
		// already passed syntactic verification
		return nil
	case tx.Chain == constants.PrimaryNetworkID:
		return ErrConvertPermissionlessChain
	case len(tx.Address) > MaxChainAddressLength:
		return ErrAddressTooLong
	case len(tx.Validators) == 0:
		return ErrConvertMustIncludeValidators
	case !isSortedAndUniqueByCompare(tx.Validators):
		return ErrConvertValidatorsNotSortedAndUnique
	}

	if err := tx.BaseTx.SyntacticVerify(rt); err != nil {
		return err
	}
	for _, vdr := range tx.Validators {
		if err := vdr.Verify(); err != nil {
			return err
		}
	}
	if err := tx.ChainAuth.Verify(); err != nil {
		return err
	}

	tx.SyntacticallyVerified = true
	return nil
}

func (tx *ConvertNetworkToL1Tx) Visit(visitor Visitor) error {
	return visitor.ConvertNetworkToL1Tx(tx)
}
