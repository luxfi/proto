// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package fee

import (
	"errors"
	"fmt"

	"github.com/luxfi/proto/p/txs"
	"github.com/luxfi/proto/p/warp"
	"github.com/luxfi/vm/components/gas"
)

var (
	_ Calculator = (*dynamicCalculator)(nil)

	ErrCalculatingComplexity = errors.New("error calculating complexity")
	ErrCalculatingGas        = errors.New("error calculating gas")
	ErrCalculatingCost       = errors.New("error calculating cost")
)

// NewDynamicCalculator returns a Calculator that prices txs against the
// supplied weights and gas price. The warpCodec is the proto/p/warp
// codec used by the L1-validator family (Register / SetWeight / etc.)
// to parse the warp message embedded in those txs; pass nil if no L1
// validator txs will be priced through this calculator.
func NewDynamicCalculator(
	warpCodec warp.Codec,
	weights gas.Dimensions,
	price gas.Price,
) Calculator {
	return &dynamicCalculator{
		warpCodec: warpCodec,
		weights:   weights,
		price:     price,
	}
}

type dynamicCalculator struct {
	warpCodec warp.Codec
	weights   gas.Dimensions
	price     gas.Price
}

func (c *dynamicCalculator) CalculateFee(tx txs.UnsignedTx) (uint64, error) {
	complexity, err := TxComplexity(c.warpCodec, tx)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrCalculatingComplexity, err)
	}
	gas, err := complexity.ToGas(c.weights)
	if err != nil {
		return 0, fmt.Errorf(
			"%w with complexity (%v) and weights (%v): %w",
			ErrCalculatingGas,
			complexity,
			c.weights,
			err,
		)
	}
	fee, err := gas.Cost(c.price)
	if err != nil {
		return 0, fmt.Errorf(
			"%w with gas (%d) and price (%d): %w",
			ErrCalculatingCost,
			gas,
			c.price,
			err,
		)
	}
	return fee, nil
}
