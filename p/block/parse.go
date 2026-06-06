// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

// Parse decodes a Block from wire bytes using the supplied Codec. The
// Codec must have block.RegisterTypes (or its equivalent) called against
// its registry before invocation.
func Parse(c Codec, b []byte) (Block, error) {
	var blk Block
	if _, err := c.Unmarshal(b, &blk); err != nil {
		return nil, err
	}
	return blk, blk.initialize(b, c)
}
