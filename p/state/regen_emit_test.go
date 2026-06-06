// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

//go:build emit_fixtures

package state

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/luxfi/proto/internal/pvmcodectest"
)

// TestEmitValidatorMetadataFixtures emits the ZAP-native wire bytes for
// the validator-metadata fixtures used by TestParseValidatorMetadata.
// Run via `go test -tags emit_fixtures -run TestEmitValidatorMetadataFixtures -v`
// to capture the regenerated wire bytes, then paste them into
// metadata_validator_test.go.
func TestEmitValidatorMetadataFixtures(t *testing.T) {
	c := pvmcodectest.NewMetadataCodec()

	type emit struct {
		name        string
		value       interface{}
		codecVer    uint16
	}

	cases := []emit{
		{
			name: "pre-delegatee reward",
			value: &preDelegateeRewardMetadata{
				UpDuration:      time.Duration(6000000),
				LastUpdated:     900000,
				PotentialReward: 100000,
			},
			codecVer: CodecVersion0,
		},
		{
			name: "potential delegatee reward",
			value: &validatorMetadata{
				UpDuration:               time.Duration(6000000),
				LastUpdated:              900000,
				PotentialReward:          100000,
				PotentialDelegateeReward: 20000,
			},
			codecVer: CodecVersion0,
		},
	}

	for _, c0 := range cases {
		t.Run(c0.name, func(t *testing.T) {
			b, err := c.Marshal(c0.codecVer, c0.value)
			if err != nil {
				t.Fatalf("Marshal %s: %v", c0.name, err)
			}
			fmt.Printf("EMIT_FIXTURE: %s = %s\n", c0.name, hex.EncodeToString(b))
		})
	}
}
