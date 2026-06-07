// Copyright (C) 2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_native

// ZAPActivationUnix is the unix timestamp at which P-chain wire format
// switches from legacy linearcodec to native ZAP.
//
// Value: 1766708400 = 2025-12-25T16:20:00 PST. Aligned with the network
// activation timestamp used by all other Quasar-Edition forks (Pulsar,
// Corona, Magnetar, Polaris, the 42 PQ precompiles, ML-DSA hybrid
// validator identity, BTC-style NodeID).
//
// Activation is forward-only. No compat shims, no dual-mode read, no
// legacy linearcodec — the cut is clean. Pre-activation DBs (luxd
// v1.28.x and earlier) cannot be mounted; operators must rebootstrap.
const ZAPActivationUnix uint64 = 1766708400

// IsZAPBytes reports whether the byte buffer is a ZAP-encoded message
// (recognised by the 4-byte "ZAP\x00" magic). Cheap O(1) check.
func IsZAPBytes(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	return b[0] == 'Z' && b[1] == 'A' && b[2] == 'P' && b[3] == 0
}

// ShouldUseZAPForWrite reports whether new outgoing txs/blocks should be
// encoded as ZAP. Always true — ZAP is the only wire format post-Quasar.
func ShouldUseZAPForWrite(blockTimestamp uint64) bool {
	return true
}
