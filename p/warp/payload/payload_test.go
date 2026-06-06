// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package payload_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/warp/payload"
)

var junkBytes = []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}

func TestParseJunk(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewPayloadCodec()
	_, err := payload.Parse(c, junkBytes)
	require.Error(err)
}

func TestParseWrongPayloadType(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewPayloadCodec()
	hashPayload, err := payload.NewHash(c, ids.GenerateTestID())
	require.NoError(err)

	shortID := ids.GenerateTestShortID()
	addressedPayload, err := payload.NewAddressedCall(
		c,
		shortID[:],
		[]byte{1, 2, 3},
	)
	require.NoError(err)

	_, err = payload.ParseAddressedCall(c, hashPayload.Bytes())
	require.ErrorIs(err, payload.ErrWrongType)

	_, err = payload.ParseHash(c, addressedPayload.Bytes())
	require.ErrorIs(err, payload.ErrWrongType)
}

func TestParse(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewPayloadCodec()
	hashPayload, err := payload.NewHash(c, ids.ID{4, 5, 6})
	require.NoError(err)

	parsedHashPayload, err := payload.Parse(c, hashPayload.Bytes())
	require.NoError(err)
	require.Equal(hashPayload, parsedHashPayload)

	addressedPayload, err := payload.NewAddressedCall(
		c,
		[]byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{10, 11, 12},
	)
	require.NoError(err)

	parsedAddressedPayload, err := payload.Parse(c, addressedPayload.Bytes())
	require.NoError(err)
	require.Equal(addressedPayload, parsedAddressedPayload)
}
