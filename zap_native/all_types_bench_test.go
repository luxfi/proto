// Copyright (C) 2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_native

import (
	"testing"

	"github.com/luxfi/crypto/bls"
	"github.com/luxfi/ids"
)

// All bench functions here exercise the ZAP-native parse + build paths.
// The legacy linearcodec variants were dropped along with the codec rip.

// ─────────────────────────────────────────────────────────────────────────
// RewardValidatorTx
// ─────────────────────────────────────────────────────────────────────────

func BenchmarkParse_ZAP_RewardValidatorTx(b *testing.B) {
	tx := NewRewardValidatorTx(ids.ID{1, 2, 3, 4, 5})
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapRewardValidatorTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_RewardValidatorTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewRewardValidatorTx(ids.ID{byte(i)})
		_ = tx.Bytes()
	}
}

// ─────────────────────────────────────────────────────────────────────────
// SetL1ValidatorWeightTx
// ─────────────────────────────────────────────────────────────────────────

func BenchmarkParse_ZAP_SetL1ValidatorWeightTx(b *testing.B) {
	tx := NewSetL1ValidatorWeightTx(ids.ID{0xaa}, 42, 1000)
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapSetL1ValidatorWeightTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_SetL1ValidatorWeightTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewSetL1ValidatorWeightTx(ids.ID{byte(i)}, uint64(i), uint64(i))
		_ = tx.Bytes()
	}
}

// ─────────────────────────────────────────────────────────────────────────
// IncreaseL1ValidatorBalanceTx
// ─────────────────────────────────────────────────────────────────────────

func BenchmarkParse_ZAP_IncreaseL1ValidatorBalanceTx(b *testing.B) {
	tx := NewIncreaseL1ValidatorBalanceTx(ids.ID{0xbb}, 5000)
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapIncreaseL1ValidatorBalanceTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_IncreaseL1ValidatorBalanceTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewIncreaseL1ValidatorBalanceTx(ids.ID{byte(i)}, uint64(i))
		_ = tx.Bytes()
	}
}

// ─────────────────────────────────────────────────────────────────────────
// DisableL1ValidatorTx
// ─────────────────────────────────────────────────────────────────────────

func BenchmarkParse_ZAP_DisableL1ValidatorTx(b *testing.B) {
	tx := NewDisableL1ValidatorTx(ids.ID{0xcc})
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapDisableL1ValidatorTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_DisableL1ValidatorTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewDisableL1ValidatorTx(ids.ID{byte(i)})
		_ = tx.Bytes()
	}
}

// ─────────────────────────────────────────────────────────────────────────
// BaseTx (Batch 2)
// ─────────────────────────────────────────────────────────────────────────

var benchBaseTxMemo = []byte("LP-023 batch 2 realistic memo payload for parse/build cost measurement")

func BenchmarkParse_ZAP_BaseTx(b *testing.B) {
	tx := NewBaseTx(1337, ids.ID{0x11}, benchBaseTxMemo)
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapBaseTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_BaseTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewBaseTx(uint32(i), ids.ID{byte(i)}, benchBaseTxMemo)
		_ = tx.Bytes()
	}
}

// ─────────────────────────────────────────────────────────────────────────
// RegisterL1ValidatorTx (Batch 2)
// ─────────────────────────────────────────────────────────────────────────

var (
	benchRegBLS [bls.PublicKeyLen]byte
	benchRegPoP [bls.SignatureLen]byte
)

func init() {
	for i := range benchRegBLS {
		benchRegBLS[i] = byte(i + 1)
	}
	for i := range benchRegPoP {
		benchRegPoP[i] = byte(0xff - i)
	}
}

func BenchmarkParse_ZAP_RegisterL1ValidatorTx(b *testing.B) {
	tx := NewRegisterL1ValidatorTx(ids.ID{0xaa}, benchRegBLS, benchRegPoP, 1_900_000_000, ids.ID{0xbb})
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapRegisterL1ValidatorTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_RegisterL1ValidatorTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewRegisterL1ValidatorTx(ids.ID{byte(i)}, benchRegBLS, benchRegPoP, uint64(i), ids.ID{byte(i)})
		_ = tx.Bytes()
	}
}

// ─────────────────────────────────────────────────────────────────────────
// SlashValidatorTx (Batch 2)
// ─────────────────────────────────────────────────────────────────────────

func BenchmarkParse_ZAP_SlashValidatorTx(b *testing.B) {
	tx := NewSlashValidatorTx(ids.NodeID{0xa1}, ids.ID{0xa2}, 100_000)
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapSlashValidatorTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_SlashValidatorTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewSlashValidatorTx(ids.NodeID{byte(i)}, ids.ID{byte(i)}, uint32(i))
		_ = tx.Bytes()
	}
}

// ─────────────────────────────────────────────────────────────────────────
// TransferChainOwnershipTx (Batch 2)
// ─────────────────────────────────────────────────────────────────────────

func BenchmarkParse_ZAP_TransferChainOwnershipTx(b *testing.B) {
	tx := NewTransferChainOwnershipTx(ids.ID{0xc0}, 1, 0, ids.ShortID{0xbe, 0xef})
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapTransferChainOwnershipTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_TransferChainOwnershipTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewTransferChainOwnershipTx(ids.ID{byte(i)}, uint32(i), uint64(i), ids.ShortID{byte(i)})
		_ = tx.Bytes()
	}
}

// ─────────────────────────────────────────────────────────────────────────
// RemoveChainValidatorTx (Batch 2)
// ─────────────────────────────────────────────────────────────────────────

func BenchmarkParse_ZAP_RemoveChainValidatorTx(b *testing.B) {
	tx := NewRemoveChainValidatorTx(ids.NodeID{0x10}, ids.ID{0xfa})
	buf := tx.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := WrapRemoveChainValidatorTx(buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuild_ZAP_RemoveChainValidatorTx(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewRemoveChainValidatorTx(ids.NodeID{byte(i)}, ids.ID{byte(i)})
		_ = tx.Bytes()
	}
}
