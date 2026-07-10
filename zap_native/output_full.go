// Copyright (C) 2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_native

import (
	"github.com/luxfi/ids"
	"github.com/luxfi/zap"
)

// TransferableOutputFull is the COMPLETE P-chain output primitive — the
// replacement for the single-address TransferableOutput stub in
// output_list.go. It faithfully models a lux.TransferableOutput whose
// FxOutput is either:
//
//   - secp256k1fx.TransferOutput{Amt, OutputOwners{Locktime, Threshold,
//     Addrs}} — StakeLocktime == 0, OR
//   - stakeable.LockOut{Locktime, TransferOutput} — StakeLocktime ==
//     LockOut.Locktime (> 0), with the inner TransferOutput's Amount/Owner
//     carried in the same entry.
//
// The multi-address owner is carried via a slice [OwnerAddrStart,
// OwnerAddrStart+OwnerAddrCount) into a shared OwnerAddrArray on the parent
// tx — exactly the sibling-array architecture InputList uses for SigIndices.
// This keeps each entry FIXED-STRIDE (branch-free zap.List.Object indexing)
// while supporting arbitrary-length address lists, and eliminates the
// single-address stub's "multi-address owners go through legacy codec" hole.
//
// Wire layout per entry (stride 72 bytes; uint64 reads alignment-tolerant):
//
//	AssetID           32B    @ 0     (ids.ID)
//	StakeLocktime     uint64 @ 32    (0 = plain output; >0 = stakeable.LockOut)
//	Amount            uint64 @ 40
//	OwnerThreshold    uint32 @ 48
//	OwnerLocktime     uint64 @ 52    (OutputOwners.Locktime)
//	OwnerAddrStart    uint32 @ 60    (index into shared OwnerAddrArray)
//	OwnerAddrCount    uint32 @ 64    (slice length)
//	Reserved          4B     @ 68    (reserved-zero; 8-aligned stride 72)
const (
	OffsetTransferableOutputFull_AssetID        = 0
	OffsetTransferableOutputFull_StakeLocktime  = 32 // uint64
	OffsetTransferableOutputFull_Amount         = 40 // uint64
	OffsetTransferableOutputFull_OwnerThreshold = 48 // uint32
	OffsetTransferableOutputFull_OwnerLocktime  = 52 // uint64
	OffsetTransferableOutputFull_OwnerAddrStart = 60 // uint32 (index into shared array)
	OffsetTransferableOutputFull_OwnerAddrCount = 64 // uint32 (slice length)
	SizeTransferableOutputFull                  = 72
)

// TransferableOutputFull is the zero-copy view over one entry in an
// OutputFullList.
//
// READ-ONLY: backed by the underlying ZAP buffer. Mutation corrupts the
// parsed tx and breaks TxID = hash(buffer).
type TransferableOutputFull struct {
	obj zap.Object
}

// AssetID returns the asset identifier (32B).
func (t TransferableOutputFull) AssetID() ids.ID {
	var out ids.ID
	for i := 0; i < 32; i++ {
		out[i] = t.obj.Uint8(OffsetTransferableOutputFull_AssetID + i)
	}
	return out
}

// StakeLocktime returns the stakeable-lock locktime. Zero means the output is
// a plain secp256k1fx.TransferOutput; non-zero means it is wrapped in a
// stakeable.LockOut with this locktime.
func (t TransferableOutputFull) StakeLocktime() uint64 {
	return t.obj.Uint64(OffsetTransferableOutputFull_StakeLocktime)
}

// IsStakeableLocked reports whether the output is a stakeable.LockOut.
func (t TransferableOutputFull) IsStakeableLocked() bool {
	return t.StakeLocktime() != 0
}

// Amount returns the (inner) TransferOutput amount.
func (t TransferableOutputFull) Amount() uint64 {
	return t.obj.Uint64(OffsetTransferableOutputFull_Amount)
}

// OwnerThreshold returns the OutputOwners threshold (signatures required).
func (t TransferableOutputFull) OwnerThreshold() uint32 {
	return t.obj.Uint32(OffsetTransferableOutputFull_OwnerThreshold)
}

// OwnerLocktime returns the OutputOwners locktime.
func (t TransferableOutputFull) OwnerLocktime() uint64 {
	return t.obj.Uint64(OffsetTransferableOutputFull_OwnerLocktime)
}

// OwnerAddrStart returns the start index into the parent tx's shared
// OwnerAddrArray for this output's owner-address slice.
func (t TransferableOutputFull) OwnerAddrStart() uint32 {
	return t.obj.Uint32(OffsetTransferableOutputFull_OwnerAddrStart)
}

// OwnerAddrCount returns the number of owner addresses for this output.
func (t TransferableOutputFull) OwnerAddrCount() uint32 {
	return t.obj.Uint32(OffsetTransferableOutputFull_OwnerAddrCount)
}

// OwnerAddrs resolves this output's owner addresses by slicing the parent's
// shared OwnerAddrArray. Pass the array obtained via OwnerAddrArrayView on
// the same parent object. The slice is bounds-clamped inside Slice.
func (t TransferableOutputFull) OwnerAddrs(shared OwnerAddrArray) []ids.ShortID {
	return shared.Slice(t.OwnerAddrStart(), t.OwnerAddrCount())
}

// OutputFullList is the zero-copy view over a list of TransferableOutputFull
// entries.
type OutputFullList struct {
	list zap.List
}

// Len returns the entry count.
func (l OutputFullList) Len() int { return l.list.Len() }

// IsNull returns true if no list pointer was set.
func (l OutputFullList) IsNull() bool { return l.list.IsNull() }

// At returns the i'th TransferableOutputFull. Returns the zero value when out
// of range (defensive: a clamped ListStride view may have Len()=0 after a
// poisoned wire length was rejected).
func (l OutputFullList) At(i int) TransferableOutputFull {
	if i < 0 || i >= l.list.Len() {
		return TransferableOutputFull{}
	}
	return TransferableOutputFull{obj: l.list.Object(i, SizeTransferableOutputFull)}
}

// OutputFullListView reads an OutputFullList from a parent object's field
// offset, applying the per-stride clamp (zap v0.7.2+) so a poisoned length
// field is rejected here.
func OutputFullListView(parent zap.Object, fieldOffset int) OutputFullList {
	return OutputFullList{list: parent.ListStride(fieldOffset, SizeTransferableOutputFull)}
}

// OutputFullListEntry is the constructor input for an OutputFullList. Addrs is
// the per-output owner-address slice; WriteOutputFullList concatenates them
// all and returns the shared array for the parent to emit via
// WriteOwnerAddrArray.
type OutputFullListEntry struct {
	AssetID       ids.ID
	StakeLocktime uint64 // 0 = plain TransferOutput; >0 = stakeable.LockOut
	Amount        uint64
	OwnerThreshold uint32
	OwnerLocktime  uint64
	Addrs          []ids.ShortID
}

// WriteOutputFullList writes an output list and emits its fixed-stride
// entries. Each entry's OwnerAddrStart points into the address array returned
// by the second result; callers MUST write that array via WriteOwnerAddrArray
// and store its (offset, length) in the parent tx's OwnerAddrArray field.
//
// Mirrors WriteInputList's shared-array contract exactly.
func WriteOutputFullList(b *zap.Builder, entries []OutputFullListEntry) (
	outputListOffset, outputListCount int,
	ownerAddrsAll []ids.ShortID,
) {
	if len(entries) == 0 {
		return 0, 0, nil
	}

	totalAddrs := 0
	for _, e := range entries {
		totalAddrs += len(e.Addrs)
	}
	ownerAddrsAll = make([]ids.ShortID, 0, totalAddrs)

	lb := b.StartList(SizeTransferableOutputFull)
	cursor := uint32(0)
	for _, e := range entries {
		var entry [SizeTransferableOutputFull]byte
		copy(entry[OffsetTransferableOutputFull_AssetID:], e.AssetID[:])
		putU64(entry[OffsetTransferableOutputFull_StakeLocktime:], e.StakeLocktime)
		putU64(entry[OffsetTransferableOutputFull_Amount:], e.Amount)
		putU32(entry[OffsetTransferableOutputFull_OwnerThreshold:], e.OwnerThreshold)
		putU64(entry[OffsetTransferableOutputFull_OwnerLocktime:], e.OwnerLocktime)
		putU32(entry[OffsetTransferableOutputFull_OwnerAddrStart:], cursor)
		putU32(entry[OffsetTransferableOutputFull_OwnerAddrCount:], uint32(len(e.Addrs)))
		lb.AddBytes(entry[:])

		ownerAddrsAll = append(ownerAddrsAll, e.Addrs...)
		cursor += uint32(len(e.Addrs))
	}
	off, _ := lb.Finish()
	return off, len(entries), ownerAddrsAll
}

// WriteOwnerAddrArray writes the concatenated owner-address array (20-byte
// stride) as a ZAP list and returns (offset, entryCount) for
// ObjectBuilder.SetList. Pair with WriteOutputFullList.
func WriteOwnerAddrArray(b *zap.Builder, addrs []ids.ShortID) (offset, entryCount int) {
	if len(addrs) == 0 {
		return 0, 0
	}
	lb := b.StartList(ids.ShortIDLen)
	for _, a := range addrs {
		lb.AddBytes(a[:])
	}
	off, _ := lb.Finish()
	return off, len(addrs)
}

// OwnerAddrArray is the zero-copy view over the parent tx's shared owner-
// address array. Each output's owner slice is Slice(start, count).
type OwnerAddrArray struct {
	list zap.List
}

// Len returns the total number of owner addresses across all outputs.
func (a OwnerAddrArray) Len() int { return a.list.Len() }

// IsNull returns true if no array pointer was set.
func (a OwnerAddrArray) IsNull() bool { return a.list.IsNull() }

// At returns the i'th address, or the zero ShortID when out of range.
func (a OwnerAddrArray) At(i int) ids.ShortID {
	var out ids.ShortID
	if i < 0 || i >= a.list.Len() {
		return out
	}
	obj := a.list.Object(i, ids.ShortIDLen)
	for j := 0; j < ids.ShortIDLen; j++ {
		out[j] = obj.Uint8(j)
	}
	return out
}

// Slice returns the per-output owner-address window [start, start+count) as a
// fresh []ids.ShortID. start and count come from attacker-controlled
// TransferableOutputFull fields, so both are clamped against Len(); any value
// outside the array yields nil without panicking. Callers MUST go through
// Slice (not At in a raw loop) so the clamp is enforced in one place —
// mirrors SigIndicesArray.Slice (RED-HIGH-3).
func (a OwnerAddrArray) Slice(start, count uint32) []ids.ShortID {
	total := uint32(a.list.Len())
	if start > total {
		return nil
	}
	if count > total-start {
		return nil
	}
	if count == 0 {
		return nil
	}
	out := make([]ids.ShortID, count)
	for i := uint32(0); i < count; i++ {
		out[i] = a.At(int(start + i))
	}
	return out
}

// OwnerAddrArrayView reads the shared owner-address array from a parent
// object's field offset, applying the per-stride clamp (stride =
// ids.ShortIDLen).
func OwnerAddrArrayView(parent zap.Object, fieldOffset int) OwnerAddrArray {
	return OwnerAddrArray{list: parent.ListStride(fieldOffset, ids.ShortIDLen)}
}
