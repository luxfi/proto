// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/warp/message"
)

func TestL1ValidatorWeight(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewMessageCodec()

	msg, err := message.NewL1ValidatorWeight(
		c,
		ids.GenerateTestID(),
		rand.Uint64(), //#nosec G404
		rand.Uint64(), //#nosec G404
	)
	require.NoError(err)

	parsed, err := message.ParseL1ValidatorWeight(c, msg.Bytes())
	require.NoError(err)
	require.Equal(msg, parsed)
}

func TestL1ValidatorWeight_Verify(t *testing.T) {
	c := pvmcodectest.NewMessageCodec()
	mustCreate := func(msg *message.L1ValidatorWeight, err error) *message.L1ValidatorWeight {
		require.NoError(t, err)
		return msg
	}
	tests := []struct {
		name     string
		msg      *message.L1ValidatorWeight
		expected error
	}{
		{
			name: "Invalid Nonce",
			msg: mustCreate(message.NewL1ValidatorWeight(
				c,
				ids.GenerateTestID(),
				math.MaxUint64,
				1,
			)),
			expected: message.ErrNonceReservedForRemoval,
		},
		{
			name: "Valid",
			msg: mustCreate(message.NewL1ValidatorWeight(
				c,
				ids.GenerateTestID(),
				math.MaxUint64,
				0,
			)),
			expected: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.msg.Verify()
			require.ErrorIs(t, err, test.expected)
		})
	}
}
