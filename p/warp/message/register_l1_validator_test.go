// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/constants"
	"github.com/luxfi/crypto/bls"
	"github.com/luxfi/crypto/bls/signer/localsigner"
	"github.com/luxfi/crypto/hash"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/warp/message"
)

func newBLSPublicKey(t *testing.T) [bls.PublicKeyLen]byte {
	sk, err := localsigner.New()
	require.NoError(t, err)

	pk := sk.PublicKey()
	pkBytes := bls.PublicKeyToCompressedBytes(pk)
	return [bls.PublicKeyLen]byte(pkBytes)
}

func TestRegisterL1Validator(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewMessageCodec()

	msg, err := message.NewRegisterL1Validator(
		c,
		ids.GenerateTestID(),
		ids.GenerateTestNodeID(),
		newBLSPublicKey(t),
		rand.Uint64(), //#nosec G404
		message.PChainOwner{
			Threshold: rand.Uint32(), //#nosec G404
			Addresses: []ids.ShortID{
				ids.GenerateTestShortID(),
			},
		},
		message.PChainOwner{
			Threshold: rand.Uint32(), //#nosec G404
			Addresses: []ids.ShortID{
				ids.GenerateTestShortID(),
			},
		},
		rand.Uint64(), //#nosec G404
	)
	require.NoError(err)

	bytes := msg.Bytes()
	var expectedValidationID ids.ID = hash.ComputeHash256Array(bytes)
	require.Equal(expectedValidationID, msg.ValidationID())

	parsed, err := message.ParseRegisterL1Validator(c, bytes)
	require.NoError(err)
	require.Equal(msg, parsed)
}

func TestRegisterL1Validator_Verify(t *testing.T) {
	c := pvmcodectest.NewMessageCodec()
	mustCreate := func(msg *message.RegisterL1Validator, err error) *message.RegisterL1Validator {
		require.NoError(t, err)
		return msg
	}
	tests := []struct {
		name     string
		msg      *message.RegisterL1Validator
		expected error
	}{
		{
			name: "PrimaryNetworkID",
			msg: mustCreate(message.NewRegisterL1Validator(
				c,
				constants.PrimaryNetworkID,
				ids.GenerateTestNodeID(),
				newBLSPublicKey(t),
				rand.Uint64(), //#nosec G404
				message.PChainOwner{
					Threshold: 1,
					Addresses: []ids.ShortID{
						ids.GenerateTestShortID(),
					},
				},
				message.PChainOwner{
					Threshold: 0,
				},
				1,
			)),
			expected: message.ErrInvalidChainID,
		},
		{
			name: "Weight = 0",
			msg: mustCreate(message.NewRegisterL1Validator(
				c,
				ids.GenerateTestID(),
				ids.GenerateTestNodeID(),
				newBLSPublicKey(t),
				rand.Uint64(), //#nosec G404
				message.PChainOwner{
					Threshold: 1,
					Addresses: []ids.ShortID{
						ids.GenerateTestShortID(),
					},
				},
				message.PChainOwner{
					Threshold: 0,
				},
				0,
			)),
			expected: message.ErrInvalidWeight,
		},
		{
			name: "Invalid NodeID Length",
			msg: &message.RegisterL1Validator{
				ChainID:      ids.GenerateTestID(),
				NodeID:       nil,
				BLSPublicKey: newBLSPublicKey(t),
				Expiry:       rand.Uint64(), //#nosec G404
				RemainingBalanceOwner: message.PChainOwner{
					Threshold: 1,
					Addresses: []ids.ShortID{
						ids.GenerateTestShortID(),
					},
				},
				DisableOwner: message.PChainOwner{
					Threshold: 0,
				},
				Weight: 1,
			},
			expected: message.ErrInvalidNodeID,
		},
		{
			name: "Invalid NodeID",
			msg: mustCreate(message.NewRegisterL1Validator(
				c,
				ids.GenerateTestID(),
				ids.EmptyNodeID,
				newBLSPublicKey(t),
				rand.Uint64(), //#nosec G404
				message.PChainOwner{
					Threshold: 1,
					Addresses: []ids.ShortID{
						ids.GenerateTestShortID(),
					},
				},
				message.PChainOwner{
					Threshold: 0,
				},
				1,
			)),
			expected: message.ErrInvalidNodeID,
		},
		{
			name: "Invalid Owner",
			msg: mustCreate(message.NewRegisterL1Validator(
				c,
				ids.GenerateTestID(),
				ids.GenerateTestNodeID(),
				newBLSPublicKey(t),
				rand.Uint64(), //#nosec G404
				message.PChainOwner{
					Threshold: 0,
					Addresses: []ids.ShortID{
						ids.GenerateTestShortID(),
					},
				},
				message.PChainOwner{
					Threshold: 0,
				},
				1,
			)),
			expected: message.ErrInvalidOwner,
		},
		{
			name: "Valid",
			msg: mustCreate(message.NewRegisterL1Validator(
				c,
				ids.GenerateTestID(),
				ids.GenerateTestNodeID(),
				newBLSPublicKey(t),
				rand.Uint64(), //#nosec G404
				message.PChainOwner{
					Threshold: 1,
					Addresses: []ids.ShortID{
						ids.GenerateTestShortID(),
					},
				},
				message.PChainOwner{
					Threshold: 0,
				},
				1,
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
