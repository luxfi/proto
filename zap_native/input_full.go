// Copyright (C) 2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_native

import (
	"github.com/luxfi/ids"
	"github.com/luxfi/zap"
)

// TransferableInputFull is the COMPLETE P-chain input primitive — the
// stakeable-aware counterpart of TransferableInput (input_list.go). It
// faithfully models a lux.TransferableInput whose FxInput is either:
//
//   - secp256k1fx.TransferInput{Amt, Input{SigIndices}} — StakeLocktime == 0,
//     OR
//   - stakeable.LockIn{Locktime, TransferInput} — StakeLocktime ==
//     LockIn.Locktime (> 0), inner amount/sig-indices in the same entry.
//
// SigIndices live in the parent tx's shared SigIndicesArray (reused verbatim
// from input_list.go via WriteSigIndicesArray / SigIndicesArrayView), sliced
// by (SigIndicesStart, SigIndicesCount). Entries stay fixed-stride.
//
// Wire layout per entry (stride 96 bytes; uint64 reads alignment-tolerant):
//
//	TxID              32B    @ 0     (ids.ID; UTXO.TxID)
//	OutputIndex       uint32 @ 32
//	AssetID           32B    @ 36    (ids.ID)
//	StakeLocktime     uint64 @ 68    (0 = plain input; >0 = stakeable.LockIn)
//	Amount            uint64 @ 76
//	SigIndicesStart   uint32 @ 84    (index into shared SigIndices array)
//	SigIndicesCount   uint32 @ 88    (slice length)
//	Reserved          4B     @ 92    (reserved-zero; 8-aligned stride 96)
const (
	OffsetTransferableInputFull_TxID            = 0
	OffsetTransferableInputFull_OutputIndex     = 32 // uint32
	OffsetTransferableInputFull_AssetID         = 36 // 32B
	OffsetTransferableInputFull_StakeLocktime   = 68 // uint64
	OffsetTransferableInputFull_Amount          = 76 // uint64
	OffsetTransferableInputFull_SigIndicesStart = 84 // uint32
	OffsetTransferableInputFull_SigIndicesCount = 88 // uint32
	SizeTransferableInputFull                   = 96
)

// TransferableInputFull is the zero-copy view over one entry in an
// InputFullList. READ-ONLY: backed by the underlying ZAP buffer.
type TransferableInputFull struct {
	obj zap.Object
}

func (t TransferableInputFull) TxID() ids.ID {
	var out ids.ID
	for i := 0; i < 32; i++ {
		out[i] = t.obj.Uint8(OffsetTransferableInputFull_TxID + i)
	}
	return out
}

func (t TransferableInputFull) OutputIndex() uint32 {
	return t.obj.Uint32(OffsetTransferableInputFull_OutputIndex)
}

func (t TransferableInputFull) AssetID() ids.ID {
	var out ids.ID
	for i := 0; i < 32; i++ {
		out[i] = t.obj.Uint8(OffsetTransferableInputFull_AssetID + i)
	}
	return out
}

// StakeLocktime returns the stakeable-lock locktime. Zero means a plain
// secp256k1fx.TransferInput; non-zero means a stakeable.LockIn with this
// locktime.
func (t TransferableInputFull) StakeLocktime() uint64 {
	return t.obj.Uint64(OffsetTransferableInputFull_StakeLocktime)
}

// IsStakeableLocked reports whether the input is a stakeable.LockIn.
func (t TransferableInputFull) IsStakeableLocked() bool {
	return t.StakeLocktime() != 0
}

func (t TransferableInputFull) Amount() uint64 {
	return t.obj.Uint64(OffsetTransferableInputFull_Amount)
}

func (t TransferableInputFull) SigIndicesStart() uint32 {
	return t.obj.Uint32(OffsetTransferableInputFull_SigIndicesStart)
}

func (t TransferableInputFull) SigIndicesCount() uint32 {
	return t.obj.Uint32(OffsetTransferableInputFull_SigIndicesCount)
}

// SigIndices resolves this input's signature indices by slicing the parent's
// shared SigIndicesArray (the same array TransferableInput uses). Pass the
// array obtained via SigIndicesArrayView on the same parent object.
func (t TransferableInputFull) SigIndices(shared SigIndicesArray) []uint32 {
	return shared.Slice(t.SigIndicesStart(), t.SigIndicesCount())
}

// InputFullList is the zero-copy view over a list of TransferableInputFull
// entries.
type InputFullList struct {
	list zap.List
}

func (l InputFullList) Len() int     { return l.list.Len() }
func (l InputFullList) IsNull() bool  { return l.list.IsNull() }

// At returns the i'th TransferableInputFull, or the zero value when out of
// range.
func (l InputFullList) At(i int) TransferableInputFull {
	if i < 0 || i >= l.list.Len() {
		return TransferableInputFull{}
	}
	return TransferableInputFull{obj: l.list.Object(i, SizeTransferableInputFull)}
}

// InputFullListView reads an InputFullList from a parent object's field
// offset, applying the per-stride clamp.
func InputFullListView(parent zap.Object, fieldOffset int) InputFullList {
	return InputFullList{list: parent.ListStride(fieldOffset, SizeTransferableInputFull)}
}

// InputFullListEntry is the constructor input for an InputFullList.
type InputFullListEntry struct {
	TxID          ids.ID
	OutputIndex   uint32
	AssetID       ids.ID
	StakeLocktime uint64 // 0 = plain TransferInput; >0 = stakeable.LockIn
	Amount        uint64
	SigIndices    []uint32
}

// WriteInputFullList writes a stakeable-aware input list and returns the
// concatenated shared SigIndices array (emit it via WriteSigIndicesArray,
// exactly as WriteInputList's contract).
func WriteInputFullList(b *zap.Builder, entries []InputFullListEntry) (
	inputListOffset, inputListCount int,
	sigIndicesAll []uint32,
) {
	if len(entries) == 0 {
		return 0, 0, nil
	}

	totalSigs := 0
	for _, e := range entries {
		totalSigs += len(e.SigIndices)
	}
	sigIndicesAll = make([]uint32, 0, totalSigs)

	lb := b.StartList(SizeTransferableInputFull)
	cursor := uint32(0)
	for _, e := range entries {
		var entry [SizeTransferableInputFull]byte
		copy(entry[OffsetTransferableInputFull_TxID:], e.TxID[:])
		putU32(entry[OffsetTransferableInputFull_OutputIndex:], e.OutputIndex)
		copy(entry[OffsetTransferableInputFull_AssetID:], e.AssetID[:])
		putU64(entry[OffsetTransferableInputFull_StakeLocktime:], e.StakeLocktime)
		putU64(entry[OffsetTransferableInputFull_Amount:], e.Amount)
		putU32(entry[OffsetTransferableInputFull_SigIndicesStart:], cursor)
		putU32(entry[OffsetTransferableInputFull_SigIndicesCount:], uint32(len(e.SigIndices)))
		lb.AddBytes(entry[:])

		sigIndicesAll = append(sigIndicesAll, e.SigIndices...)
		cursor += uint32(len(e.SigIndices))
	}
	off, _ := lb.Finish()
	return off, len(entries), sigIndicesAll
}
