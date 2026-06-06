// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_codec

import (
	"testing"
)

// walletTx mirrors the typical wallet-emitted transaction shape: a u32
// network id, two 32-byte hashes, a u64 nonce + amount, and a
// short memo blob. This is roughly the wire footprint of a PVM
// AddPermissionlessValidator or an XVM BaseTx — the workload sdk/wallet
// is built to marshal/unmarshal on every tx.
//
// Comparative benchmarks against the historical linearcodec baseline
// live in the bench module at bench/modes/zap_vs_codec/. This file
// carries the ZAP-native bench only so proto/zap_codec is testable
// without re-importing the archived luxfi/codec module.
type walletTx struct {
	NetworkID    uint32   `serialize:"true"`
	BlockchainID [32]byte `serialize:"true"`
	NodeID       [32]byte `serialize:"true"`
	Nonce        uint64   `serialize:"true"`
	Amount       uint64   `serialize:"true"`
	Memo         []byte   `serialize:"true"`
}

func makeWalletTx() *walletTx {
	tx := &walletTx{
		NetworkID: 0x12345678,
		Nonce:     0xdeadbeefcafebabe,
		Amount:    1_000_000_000_000_000,
		Memo:      make([]byte, 32),
	}
	for i := range tx.BlockchainID {
		tx.BlockchainID[i] = byte(i)
	}
	for i := range tx.NodeID {
		tx.NodeID[i] = byte(i ^ 0xff)
	}
	for i := range tx.Memo {
		tx.Memo[i] = byte(i + 1)
	}
	return tx
}

// BenchmarkMarshal_ZAPNative measures the post-Wave-2G wallet marshal
// cost — zapcodec little-endian, proto/zap_codec.Manager-wrapped.
func BenchmarkMarshal_ZAPNative(b *testing.B) {
	cm := NewVersionedManager(0, MaxSize)
	tx := makeWalletTx()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := cm.Marshal(0, tx)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf
	}
}

// BenchmarkUnmarshal_ZAPNative measures the post-Wave-2G wallet
// unmarshal cost. Buffer is pre-built so the bench measures parse
// only.
func BenchmarkUnmarshal_ZAPNative(b *testing.B) {
	cm := NewVersionedManager(0, MaxSize)
	tx := makeWalletTx()
	buf, err := cm.Marshal(0, tx)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(buf)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out := &walletTx{}
		if _, err := cm.Unmarshal(buf, out); err != nil {
			b.Fatal(err)
		}
	}
}
