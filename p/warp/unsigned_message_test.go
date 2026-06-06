// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package warp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/constants"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/warp"
)

func TestUnsignedMessage(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewWarpCodec()

	msg, err := warp.NewUnsignedMessage(
		c,
		constants.UnitTestID,
		ids.GenerateTestID(),
		[]byte("payload"),
	)
	require.NoError(err)

	msgBytes := msg.Bytes()
	msg2, err := warp.ParseUnsignedMessage(c, msgBytes)
	require.NoError(err)
	require.Equal(msg, msg2)
}

func TestParseUnsignedMessageJunk(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewWarpCodec()

	bytes := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	_, err := warp.ParseUnsignedMessage(c, bytes)
	require.Error(err)
}
