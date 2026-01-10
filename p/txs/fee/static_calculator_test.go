// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package fee

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/protocol/p/txs"
)

func TestStaticCalculator(t *testing.T) {
	calculator := NewSimpleStaticCalculator(StaticConfig{})
	for _, test := range txTests {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)

			txBytes, err := hex.DecodeString(test.tx)
			require.NoError(err)

			tx, err := txs.Parse(txs.Codec, txBytes)
			if err != nil {
				t.Skipf("skipping invalid tx encoding: %v", err)
			}

			_, err = calculator.CalculateFee(tx.Unsigned)
			require.ErrorIs(err, test.expectedStaticFeeErr)
		})
	}
}
