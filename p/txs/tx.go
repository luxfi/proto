// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"errors"
	"fmt"

	"github.com/luxfi/crypto/hash"
	"github.com/luxfi/crypto/secp256k1"
	"github.com/luxfi/ids"
	"github.com/luxfi/math/set"
	"github.com/luxfi/p2p/gossip"
	"github.com/luxfi/runtime"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"
	"github.com/luxfi/vm/components/verify"
)

var (
	_ gossip.Gossipable = (*Tx)(nil)

	ErrNilSignedTx = errors.New("nil signed tx is not valid")

	errSignedTxNotInitialized = errors.New("signed tx was never initialized and is not valid")
)

// Tx is a signed transaction: an unsigned tx plus its credentials. The signed
// bytes are the unsigned zap buffer followed by the credential buffer —
// nothing wraps the pair, and the unsigned buffer's own size field is where a
// reader splits them.
type Tx struct {
	// The body of this transaction
	Unsigned UnsignedTx `serialize:"true" json:"unsignedTx"`

	// The credentials of this transaction
	Creds []verify.Verifiable `serialize:"true" json:"credentials"`

	TxID  ids.ID `json:"id"`
	bytes []byte
}

// NewSigned builds a signed tx from an unsigned body and signs it.
func NewSigned(unsigned UnsignedTx, signers [][]*secp256k1.PrivateKey) (*Tx, error) {
	res := &Tx{Unsigned: unsigned}
	return res, res.Sign(signers)
}

// Initialize binds the signed bytes and TxID from the unsigned body and the
// credentials already attached — for txs assembled without signing here (a
// multisig collection, a genesis tx).
func (tx *Tx) Initialize() error {
	unsignedBytes, err := Marshal(tx.Unsigned)
	if err != nil {
		return fmt.Errorf("couldn't marshal unsigned tx: %w", err)
	}
	tx.Unsigned.SetBytes(unsignedBytes)
	return tx.Bind(unsignedBytes)
}

// SetBytes binds the exact signed bytes and derives TxID = hash(signedBytes).
func (tx *Tx) SetBytes(signedBytes []byte) {
	tx.bytes = signedBytes
	tx.TxID = hash.ComputeHash256Array(signedBytes)
}

// Bind sets the signed bytes to the given unsigned buffer followed by the
// credentials currently attached. It is the one place the pair is joined.
func (tx *Tx) Bind(unsignedBytes []byte) error {
	signedBytes := unsignedBytes
	if len(tx.Creds) > 0 {
		credsBuf, err := writeCredsBuf(tx.Creds)
		if err != nil {
			return fmt.Errorf("couldn't encode credentials: %w", err)
		}
		signedBytes = concat(unsignedBytes, credsBuf)
	}
	tx.SetBytes(signedBytes)
	return nil
}

// Parse reads a signed tx from its wire bytes: the leading self-delimiting zap
// buffer is the unsigned body, any remainder is the credential buffer.
func Parse(signedBytes []byte) (*Tx, error) {
	n, err := zapLen(signedBytes)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse tx: %w", err)
	}
	unsigned, err := Unmarshal(signedBytes[:n])
	if err != nil {
		return nil, fmt.Errorf("couldn't parse unsigned tx: %w", err)
	}
	tx := &Tx{Unsigned: unsigned}
	if len(signedBytes) > n {
		creds, err := parseCredsBuf(signedBytes[n:])
		if err != nil {
			return nil, fmt.Errorf("couldn't parse credentials: %w", err)
		}
		tx.Creds = creds
	}
	tx.SetBytes(signedBytes)
	return tx, nil
}

func (tx *Tx) Bytes() []byte {
	return tx.bytes
}

func (tx *Tx) Size() int {
	return len(tx.bytes)
}

func (tx *Tx) ID() ids.ID {
	return tx.TxID
}

func (tx *Tx) GossipID() ids.ID {
	return tx.TxID
}

// UTXOs returns the UTXOs transaction is producing.
func (tx *Tx) UTXOs() []*lux.UTXO {
	outs := tx.Unsigned.Outputs()
	utxos := make([]*lux.UTXO, len(outs))
	for i, out := range outs {
		utxos[i] = &lux.UTXO{
			UTXOID: lux.UTXOID{
				TxID:        tx.TxID,
				OutputIndex: uint32(i),
			},
			Asset: lux.Asset{ID: out.AssetID()},
			Out:   out.Out,
		}
	}
	return utxos
}

// InputIDs returns the set of inputs this transaction consumes
func (tx *Tx) InputIDs() set.Set[ids.ID] {
	return tx.Unsigned.InputIDs()
}

func (tx *Tx) SyntacticVerify(rt *runtime.Runtime) error {
	switch {
	case tx == nil:
		return ErrNilSignedTx
	case tx.TxID == ids.Empty:
		return errSignedTxNotInitialized
	default:
		return tx.Unsigned.SyntacticVerify(rt)
	}
}

// Sign attaches a credential per signer set over the unsigned bytes, then
// binds signed bytes = unsigned ‖ creds.
func (tx *Tx) Sign(signers [][]*secp256k1.PrivateKey) error {
	unsignedBytes, err := Marshal(tx.Unsigned)
	if err != nil {
		return fmt.Errorf("couldn't marshal unsigned tx: %w", err)
	}
	tx.Unsigned.SetBytes(unsignedBytes)
	h := hash.ComputeHash256(unsignedBytes)

	tx.Creds = nil
	for _, keys := range signers {
		cred := &secp256k1fx.Credential{
			Sigs: make([][secp256k1.SignatureLen]byte, len(keys)),
		}
		for j, key := range keys {
			sig, err := key.SignHash(h)
			if err != nil {
				return fmt.Errorf("problem signing tx: %w", err)
			}
			copy(cred.Sigs[j][:], sig)
		}
		tx.Creds = append(tx.Creds, cred)
	}
	return tx.Bind(unsignedBytes)
}
