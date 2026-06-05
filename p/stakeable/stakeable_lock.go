// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package stakeable

import (
	"encoding/binary"
	"errors"

	"github.com/luxfi/runtime"
	lux "github.com/luxfi/utxo"
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

// Bytes returns a stable byte representation of this LockOut for use in
// canonical sort comparisons (see lux.IsSortedTransferableOutputs).
//
// Layout: stakeableLockMarker || big-endian Locktime || inner-output
// Bytes() (empty if the inner output does not expose Bytes).
//
// The leading 0xFF marker ensures LockOut sorts AFTER any unlocked fxs
// primitive (whose first byte is a dense TypeKind << 0xFF), preserving
// the legacy "unlocked first, then locked" ordering that downstream tx
// SyntacticVerify code expects.
func (s *LockOut) Bytes() []byte {
	const stakeableLockMarker byte = 0xFF
	var inner []byte
	if ws, ok := s.TransferableOut.(interface{ Bytes() []byte }); ok {
		inner = ws.Bytes()
	}
	b := make([]byte, 1+8+len(inner))
	b[0] = stakeableLockMarker
	binary.BigEndian.PutUint64(b[1:9], s.Locktime)
	copy(b[9:], inner)
	return b
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
