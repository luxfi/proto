// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package fee

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/protocol/p/txs"
)

func TestDynamicCalculator(t *testing.T) {
	calculator := NewDynamicCalculator(testDynamicWeights, testDynamicPrice)
	for _, test := range txTests {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)

			txBytes, err := hex.DecodeString(test.tx)
			require.NoError(err)

			tx, err := txs.Parse(txs.Codec, txBytes)
			if err != nil {
				t.Skipf("skipping invalid tx encoding: %v", err)
			}

			fee, err := calculator.CalculateFee(tx.Unsigned)
			require.Equal(int(test.expectedDynamicFee), int(fee))
			require.ErrorIs(err, test.expectedDynamicFeeErr)
		})
	}
}
