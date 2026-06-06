// Copyright (C) 2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package zap_native

import (
	"errors"
	"os"
)

// ZAPActivationUnix is the unix timestamp at which P-chain wire format
// switches from legacy linearcodec (CodecVersionV0/V1) to native ZAP.
// The switch governs the WRITE path: post-activation blocks MUST be
// ZAP-encoded. Validator coordination is documented in LP-023.
//
// Value: 1766708400 = 2025-12-25T16:20:00 PST. Aligned with the network
// activation timestamp used by all other Quasar-Edition forks (Pulsar,
// Corona, Magnetar, Polaris, the 42 PQ precompiles, ML-DSA hybrid
// validator identity, BTC-style NodeID).
//
// Pre-activation reads are handled transparently by the zap_codec
// Unmarshal path's BE-fallback decoder — luxd binaries can mount and
// continue serving DBs written by pre-LP-023 versions (v1.28.x and
// earlier) without operator intervention. Writes from post-activation
// timestamps emit LE; reads accept either endianness.
const ZAPActivationUnix uint64 = 1766708400

// LegacyEnabled is true when the operator has set LUXD_ENABLE_LEGACY_CODEC=1.
//
// Default: true. Validators upgrading from v1.28.x or earlier hold P-chain
// state written with the pre-LP-023 BE codec version prefix. zap_codec
// reads handle BE transparently via fallback in Unmarshal, but the broader
// legacy linearcodec path remains reachable so historical tx bytes that
// did not migrate cleanly through the BE-fallback shim can still be
// decoded.
//
// Operators can explicitly disable via LUXD_ENABLE_LEGACY_CODEC=0 once
// their DB has been fully resynced from a post-activation snapshot.
var LegacyEnabled = os.Getenv("LUXD_ENABLE_LEGACY_CODEC") != "0"

// ErrLegacyCodecDisabled is returned by legacy-path parsers when the operator
// has not enabled legacy codec support via the LUXD_ENABLE_LEGACY_CODEC env var.
// Native ZAP is the default; legacy is opt-in.
var ErrLegacyCodecDisabled = errors.New("zap_native: legacy codec not enabled (set LUXD_ENABLE_LEGACY_CODEC=1 to read pre-activation blocks)")

// IsZAPBytes reports whether the byte buffer is a ZAP-encoded message
// (recognised by the 4-byte "ZAP\x00" magic). Cheap O(1) check.
//
// Callers use this to discriminate ZAP-encoded txs/blocks from legacy
// linearcodec-encoded ones during the cross-activation read window (which
// only happens when LUXD_ENABLE_LEGACY_CODEC=1).
func IsZAPBytes(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	return b[0] == 'Z' && b[1] == 'A' && b[2] == 'P' && b[3] == 0
}

// ShouldUseZAPForWrite reports whether new outgoing txs/blocks should be
// encoded as ZAP.
//
// Default behavior — native ZAP always. ZAPActivationUnix is now 0
// (always-on), so the timestamp gate is a no-op except when the operator
// explicitly opts back into legacy via LUXD_ENABLE_LEGACY_CODEC=1, in
// which case the original forward-date semantics are preserved for the
// LegacyEnabled path only.
func ShouldUseZAPForWrite(blockTimestamp uint64) bool {
	if !LegacyEnabled {
		return true // native ZAP default
	}
	return blockTimestamp >= ZAPActivationUnix
}
