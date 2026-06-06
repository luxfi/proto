// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package payload_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/warp/payload"
)

func TestAddressedCall(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewPayloadCodec()
	shortID := ids.GenerateTestShortID()

	addressedPayload, err := payload.NewAddressedCall(
		c,
		shortID[:],
		[]byte{1, 2, 3},
	)
	require.NoError(err)

	addressedPayloadBytes := addressedPayload.Bytes()
	parsedAddressedPayload, err := payload.ParseAddressedCall(c, addressedPayloadBytes)
	require.NoError(err)
	require.Equal(addressedPayload, parsedAddressedPayload)
}

func TestParseAddressedCallJunk(t *testing.T) {
	c := pvmcodectest.NewPayloadCodec()
	_, err := payload.ParseAddressedCall(c, junkBytes)
	require.Error(t, err)
}

func TestAddressedCallBytes(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewPayloadCodec()
	base64Payload := "AAAAAAABAAAAEAECAwAAAAAAAAAAAAAAAAAAAAADCgsM"
	addressedPayload, err := payload.NewAddressedCall(
		c,
		[]byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{10, 11, 12},
	)
	require.NoError(err)
	require.Equal(base64Payload, base64.StdEncoding.EncodeToString(addressedPayload.Bytes()))
}
