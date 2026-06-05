// Copyright (C) 2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_native

import "testing"

// BenchmarkParse_ZAP measures the cost of wrapping a ZAP-encoded buffer
// in a typed accessor (zero-copy).
func BenchmarkParse_ZAP(b *testing.B) {
	tx := NewAdvanceTimeTx(1782604800)
	buf := tx.Bytes()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t2, err := WrapAdvanceTimeTx(buf)
		if err != nil {
			b.Fatal(err)
		}
		_ = t2.Time()
	}
}

// BenchmarkBuild_ZAP measures the cost of constructing an AdvanceTimeTx
// via direct ZAP offset writes (no reflection, no codec lookup).
func BenchmarkBuild_ZAP(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := NewAdvanceTimeTx(uint64(i))
		_ = tx.Bytes()
	}
}

// BenchmarkFieldAccess_ZAP measures field read on a ZAP-wrapped accessor.
// Reads compile down to a binary.LittleEndian.Uint64 at a known offset —
// a few instructions, in the same order of magnitude as a struct-field
// load.
func BenchmarkFieldAccess_ZAP(b *testing.B) {
	tx := NewAdvanceTimeTx(1782604800)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tx.Time()
	}
}
