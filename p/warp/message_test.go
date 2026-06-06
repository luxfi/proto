// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package warp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/constants"
	"github.com/luxfi/crypto/bls"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pcodectest"
	"github.com/luxfi/proto/p/warp"
)

func TestMessage(t *testing.T) {
	require := require.New(t)
	c := pcodectest.NewWarpCodec()

	payload := []byte("payload")

	unsignedMsg, err := warp.NewUnsignedMessage(
		c,
		constants.UnitTestID,
		ids.GenerateTestID(),
		payload,
	)
	require.NoError(err)
	require.Len(unsignedMsg.Bytes(), 42+len(payload))

	msg, err := warp.NewMessage(
		c,
		unsignedMsg,
		&warp.BitSetSignature{
			Signers:   []byte{1, 2, 3},
			Signature: [bls.SignatureLen]byte{4, 5, 6},
		},
	)
	require.NoError(err)

	msgBytes := msg.Bytes()
	msg2, err := warp.ParseMessage(c, msgBytes)
	require.NoError(err)
	require.Equal(msg, msg2)
}

func TestParseMessageJunk(t *testing.T) {
	require := require.New(t)
	c := pcodectest.NewWarpCodec()

	bytes := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	_, err := warp.ParseMessage(c, bytes)
	require.Error(err)
}
