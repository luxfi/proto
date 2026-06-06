// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"cmp"
	"slices"
)

func sortByCompare[T interface{ Compare(T) int }](s []T) {
	slices.SortFunc(s, func(a, b T) int {
		return a.Compare(b)
	})
}

// SortByCompare is the exported entry point for sortByCompare. Tests in
// external test packages (proto/p/txs_test, etc.) reach for this when
// they need to assemble the canonical input/output ordering that
// matches what the production txs.Tx serialiser produces.
func SortByCompare[T interface{ Compare(T) int }](s []T) {
	sortByCompare(s)
}

func isSortedAndUniqueByCompare[T interface{ Compare(T) int }](s []T) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1].Compare(s[i]) >= 0 {
			return false
		}
	}
	return true
}

func isSortedAndUniqueOrdered[T cmp.Ordered](s []T) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] >= s[i] {
			return false
		}
	}
	return true
}
