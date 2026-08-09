// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stakeable

import (
	"errors"

	"github.com/luxfi/runtime"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/wire"
)

var (
	errInvalidLocktime      = errors.New("invalid locktime")
	errNestedStakeableLocks = errors.New("shouldn't nest stakeable locks")
)

type LockOut struct {
	Locktime            uint64 `serialize:"true" json:"locktime"`
	lux.TransferableOut `serialize:"true" json:"output"`
}

func (s *LockOut) InitRuntime(rt *runtime.Runtime) {
	// Initialize the context for the underlying output if it supports it
	if contextOutput, ok := s.TransferableOut.(interface{ InitRuntime(*runtime.Runtime) }); ok {
		contextOutput.InitRuntime(rt)
	}
}

func (s *LockOut) Addresses() [][]byte {
	if addressable, ok := s.TransferableOut.(lux.Addressable); ok {
		return addressable.Addresses()
	}
	return nil
}

// Bytes returns the ZAP wire envelope for this LockOut: the shared
// wire.LockedOutput schema (TypeKindReserved, ShapeKindLockedOutput),
// carrying the locktime and the inner output's own envelope.
//
// The canonical output sort compares exactly these bytes, and the chain
// re-serializes a parsed LockOut through the SAME wire.NewLockedOutput.
// Any second encoding here is a fork: the two sides then disagree on
// output ORDER and the chain rejects the tx as "outputs not sorted".
func (s *LockOut) Bytes() []byte {
	var inner []byte
	if ws, ok := s.TransferableOut.(interface{ Bytes() []byte }); ok {
		inner = ws.Bytes()
	}
	return wire.NewLockedOutput(wire.LockedOutputInput{
		Locktime:         s.Locktime,
		TransferOutBytes: inner,
	})
}

func (s *LockOut) Verify() error {
	if s.Locktime == 0 {
		return errInvalidLocktime
	}
	if _, nested := s.TransferableOut.(*LockOut); nested {
		return errNestedStakeableLocks
	}
	return s.TransferableOut.Verify()
}

type LockIn struct {
	Locktime           uint64 `serialize:"true" json:"locktime"`
	lux.TransferableIn `serialize:"true" json:"input"`
}

func (s *LockIn) Verify() error {
	if s.Locktime == 0 {
		return errInvalidLocktime
	}
	if _, nested := s.TransferableIn.(*LockIn); nested {
		return errNestedStakeableLocks
	}
	return s.TransferableIn.Verify()
}
