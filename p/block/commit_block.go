// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"context"
	"time"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/p/txs"
	"github.com/luxfi/runtime"
)

var (
	_ BanffBlock = (*BanffCommitBlock)(nil)
	_ Block      = (*ApricotCommitBlock)(nil)
)

type BanffCommitBlock struct {
	Time               uint64 `serialize:"true" json:"time"`
	ApricotCommitBlock `serialize:"true"`
}

func (b *BanffCommitBlock) Timestamp() time.Time {
	return time.Unix(int64(b.Time), 0)
}

func (b *BanffCommitBlock) Visit(v Visitor) error {
	return v.BanffCommitBlock(b)
}

// NewBanffCommitBlock builds and initializes a BanffCommitBlock against
// the supplied block Codec.
func NewBanffCommitBlock(
	c Codec,
	timestamp time.Time,
	parentID ids.ID,
	height uint64,
) (*BanffCommitBlock, error) {
	blk := &BanffCommitBlock{
		Time: uint64(timestamp.Unix()),
		ApricotCommitBlock: ApricotCommitBlock{
			CommonBlock: CommonBlock{
				PrntID: parentID,
				Hght:   height,
			},
		},
	}
	return blk, initialize(c, blk, &blk.CommonBlock)
}

type ApricotCommitBlock struct {
	CommonBlock `serialize:"true"`
}

func (b *ApricotCommitBlock) initialize(bytes []byte, _ Codec) error {
	b.CommonBlock.initialize(bytes)
	return nil
}

func (*ApricotCommitBlock) InitRuntime(*runtime.Runtime) {}

func (*ApricotCommitBlock) Txs() []*txs.Tx {
	return nil
}

func (b *ApricotCommitBlock) Visit(v Visitor) error {
	return v.ApricotCommitBlock(b)
}

// NewApricotCommitBlock builds and initializes an ApricotCommitBlock
// against the supplied block Codec.
func NewApricotCommitBlock(
	c Codec,
	parentID ids.ID,
	height uint64,
) (*ApricotCommitBlock, error) {
	blk := &ApricotCommitBlock{
		CommonBlock: CommonBlock{
			PrntID: parentID,
			Hght:   height,
		},
	}
	return blk, initialize(c, blk, &blk.CommonBlock)
}

// InitializeWithContext initializes the block with consensus context
func (b *BanffCommitBlock) InitializeWithContext(ctx context.Context) error {
	// Initialize any context-dependent fields here
	return nil
}

// InitializeWithContext initializes the block with consensus context
func (b *ApricotCommitBlock) InitializeWithContext(ctx context.Context) error {
	// Initialize any context-dependent fields here
	return nil
}
