// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/database/memdb"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pcodectest"
)

func TestParseDelegatorMetadata(t *testing.T) {
	c := pcodectest.NewMetadataCodec()
	type test struct {
		name      string
		bytes     []byte
		expected  *delegatorMetadata
		expectErr bool
	}
	tests := []test{
		{
			name: "potential reward only no codec",
			bytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7b,
			},
			expected: &delegatorMetadata{
				PotentialReward: 123,
				StakerStartTime: 0,
			},
		},
		{
			name: "potential reward + staker start time with codec v1",
			bytes: []byte{
				// codec version
				0x00, 0x01,
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7b,
				// staker start time
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xc8,
			},
			expected: &delegatorMetadata{
				PotentialReward: 123,
				StakerStartTime: 456,
			},
		},
		{
			name: "invalid codec version",
			bytes: []byte{
				// codec version
				0x00, 0x02,
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7b,
				// staker start time
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xc8,
			},
			expected:  nil,
			expectErr: true,
		},
		{
			name: "short byte len",
			bytes: []byte{
				// codec version
				0x00, 0x01,
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7b,
				// staker start time
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expected:  nil,
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			var metadata delegatorMetadata
			err := parseDelegatorMetadata(c, tt.bytes, &metadata)
			if tt.expectErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			require.Equal(tt.expected, &metadata)
		})
	}
}

func TestWriteDelegatorMetadata(t *testing.T) {
	c := pcodectest.NewMetadataCodec()
	type test struct {
		name     string
		version  uint16
		metadata *delegatorMetadata
		expected []byte
	}
	tests := []test{
		{
			name:    CodecVersion0Tag,
			version: CodecVersion0,
			metadata: &delegatorMetadata{
				PotentialReward: 123,
				StakerStartTime: 456,
			},
			expected: []byte{
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7b,
			},
		},
		{
			name:    CodecVersion1Tag,
			version: CodecVersion1,
			metadata: &delegatorMetadata{
				PotentialReward: 123,
				StakerStartTime: 456,
			},
			expected: []byte{
				// codec version
				0x00, 0x01,
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7b,
				// staker start time
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0xc8,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			db := memdb.New()
			tt.metadata.txID = ids.GenerateTestID()
			require.NoError(writeDelegatorMetadata(c, db, tt.metadata, tt.version))
			bytes, err := db.Get(tt.metadata.txID[:])
			require.NoError(err)
			require.Equal(tt.expected, bytes)
		})
	}
}
