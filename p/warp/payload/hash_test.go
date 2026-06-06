// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package payload_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pcodectest"
	"github.com/luxfi/proto/p/warp/payload"
)

func TestHash(t *testing.T) {
	require := require.New(t)
	c := pcodectest.NewPayloadCodec()

	hashPayload, err := payload.NewHash(c, ids.GenerateTestID())
	require.NoError(err)

	hashPayloadBytes := hashPayload.Bytes()
	parsedHashPayload, err := payload.ParseHash(c, hashPayloadBytes)
	require.NoError(err)
	require.Equal(hashPayload, parsedHashPayload)
}

func TestParseHashJunk(t *testing.T) {
	c := pcodectest.NewPayloadCodec()
	_, err := payload.ParseHash(c, junkBytes)
	require.Error(t, err)
}

func TestHashBytes(t *testing.T) {
	require := require.New(t)
	c := pcodectest.NewPayloadCodec()
	base64Payload := "AAAAAAAABAUGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	hashPayload, err := payload.NewHash(c, ids.ID{4, 5, 6})
	require.NoError(err)
	require.Equal(base64Payload, base64.StdEncoding.EncodeToString(hashPayload.Bytes()))
}
