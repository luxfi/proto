// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pcodectest"
	"github.com/luxfi/proto/p/block"
)

func TestNewBanffCommitBlock(t *testing.T) {
	require := require.New(t)
	c := pcodectest.NewPVMCodecs().GenesisCodec

	timestamp := time.Now().Truncate(time.Second)
	parentID := ids.GenerateTestID()
	height := uint64(1337)
	blk, err := block.NewBanffCommitBlock(c,
		timestamp,
		parentID,
		height,
	)
	require.NoError(err)

	// Make sure the block is initialized
	require.NotEmpty(blk.Bytes())

	require.Equal(timestamp, blk.Timestamp())
	require.Equal(parentID, blk.Parent())
	require.Equal(height, blk.Height())
}

func TestNewApricotCommitBlock(t *testing.T) {
	require := require.New(t)
	c := pcodectest.NewPVMCodecs().GenesisCodec

	parentID := ids.GenerateTestID()
	height := uint64(1337)
	blk, err := block.NewApricotCommitBlock(c,
		parentID,
		height,
	)
	require.NoError(err)

	// Make sure the block is initialized
	require.NotEmpty(blk.Bytes())

	require.Equal(parentID, blk.Parent())
	require.Equal(height, blk.Height())
}
