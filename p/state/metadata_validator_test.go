// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/database"
	"github.com/luxfi/database/memdb"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
)

func TestValidatorUptimes(t *testing.T) {
	require := require.New(t)
	state := newValidatorState()

	// get non-existent uptime
	nodeID := ids.GenerateTestNodeID()
	netID := ids.GenerateTestID()
	_, _, err := state.GetUptime(nodeID, netID)
	require.ErrorIs(err, database.ErrNotFound)

	// set non-existent uptime
	err = state.SetUptime(nodeID, netID, 1, time.Now())
	require.ErrorIs(err, database.ErrNotFound)

	testMetadata := &validatorMetadata{
		UpDuration:  time.Hour,
		lastUpdated: time.Now(),
	}
	// load uptime
	state.LoadValidatorMetadata(nodeID, netID, testMetadata)

	// get uptime
	upDuration, lastUpdated, err := state.GetUptime(nodeID, netID)
	require.NoError(err)
	require.Equal(testMetadata.UpDuration, upDuration)
	require.Equal(testMetadata.lastUpdated, lastUpdated)

	// set uptime
	newUpDuration := testMetadata.UpDuration + 1
	newLastUpdated := testMetadata.lastUpdated.Add(time.Hour)
	require.NoError(state.SetUptime(nodeID, netID, newUpDuration, newLastUpdated))

	// get new uptime
	upDuration, lastUpdated, err = state.GetUptime(nodeID, netID)
	require.NoError(err)
	require.Equal(newUpDuration, upDuration)
	require.Equal(newLastUpdated, lastUpdated)

	// set uptime for non-existent
	err = state.SetUptime(ids.GenerateTestNodeID(), netID, 1, time.Now())
	require.ErrorIs(err, database.ErrNotFound)

	// delete uptime
	state.DeleteValidatorMetadata(nodeID, netID)

	// get deleted uptime
	_, _, err = state.GetUptime(nodeID, netID)
	require.ErrorIs(err, database.ErrNotFound)
}

func TestValidatorDelegateeRewards(t *testing.T) {
	require := require.New(t)
	state := newValidatorState()

	// get non-existent delegatee reward
	nodeID := ids.GenerateTestNodeID()
	netID := ids.GenerateTestID()
	_, err := state.GetDelegateeReward(netID, nodeID)
	require.ErrorIs(err, database.ErrNotFound)

	// set non-existent delegatee reward
	err = state.SetDelegateeReward(netID, nodeID, 100000)
	require.ErrorIs(err, database.ErrNotFound)

	testMetadata := &validatorMetadata{
		PotentialDelegateeReward: 100000,
	}
	// load delegatee reward
	state.LoadValidatorMetadata(nodeID, netID, testMetadata)

	// get delegatee reward
	delegateeReward, err := state.GetDelegateeReward(netID, nodeID)
	require.NoError(err)
	require.Equal(testMetadata.PotentialDelegateeReward, delegateeReward)

	// set delegatee reward
	newDelegateeReward := testMetadata.PotentialDelegateeReward + 100000
	require.NoError(state.SetDelegateeReward(netID, nodeID, newDelegateeReward))

	// get new delegatee reward
	delegateeReward, err = state.GetDelegateeReward(netID, nodeID)
	require.NoError(err)
	require.Equal(newDelegateeReward, delegateeReward)

	// set delegatee reward for non-existent
	err = state.SetDelegateeReward(netID, ids.GenerateTestNodeID(), 1)
	require.ErrorIs(err, database.ErrNotFound)

	// delete delegatee reward
	state.DeleteValidatorMetadata(nodeID, netID)

	// get deleted delegatee reward
	_, err = state.GetDelegateeReward(netID, nodeID)
	require.ErrorIs(err, database.ErrNotFound)
}

func TestWriteValidatorMetadata(t *testing.T) {
	require := require.New(t)
	state := newValidatorState()
	c := pvmcodectest.NewMetadataCodec()

	primaryDB := memdb.New()
	netDB := memdb.New()
	// write empty uptimes
	require.NoError(state.WriteValidatorMetadata(c, primaryDB, netDB, CodecVersion1))

	// load metadata for a chain validator. Load alone does NOT mark the
	// entry as updated — only Set* mutations trigger a subsequent write.
	nodeID := ids.GenerateTestNodeID()
	netID := ids.GenerateTestID()
	testUptimeReward := &validatorMetadata{
		UpDuration:      time.Hour,
		lastUpdated:     time.Now(),
		PotentialReward: 100,
		txID:            ids.GenerateTestID(),
	}
	state.LoadValidatorMetadata(nodeID, netID, testUptimeReward)

	// Without a Set, WriteValidatorMetadata is a no-op for this txID.
	require.NoError(state.WriteValidatorMetadata(c, primaryDB, netDB, CodecVersion1))
	netDBHas, err := netDB.Has(testUptimeReward.txID[:])
	require.NoError(err)
	require.False(netDBHas)

	// Mark the entry as updated via SetUptime, then re-write — now the
	// chain validator's row should land in netDB (not primaryDB).
	require.NoError(state.SetUptime(nodeID, netID, testUptimeReward.UpDuration+1, testUptimeReward.lastUpdated))
	require.NoError(state.WriteValidatorMetadata(c, primaryDB, netDB, CodecVersion1))
	netDBHas, err = netDB.Has(testUptimeReward.txID[:])
	require.NoError(err)
	require.True(netDBHas)
	primaryDBHas, err := primaryDB.Has(testUptimeReward.txID[:])
	require.NoError(err)
	require.False(primaryDBHas)

	// Same for a primary-network validator — Set then Write lands in primaryDB.
	primaryNodeID := ids.GenerateTestNodeID()
	primaryNetID := ids.Empty // primary network ID
	testUptimeReward2 := &validatorMetadata{
		UpDuration:      time.Hour,
		lastUpdated:     time.Now(),
		PotentialReward: 100,
		txID:            ids.GenerateTestID(),
	}
	state.LoadValidatorMetadata(primaryNodeID, primaryNetID, testUptimeReward2)
	require.NoError(state.SetUptime(primaryNodeID, primaryNetID, testUptimeReward2.UpDuration+1, testUptimeReward2.lastUpdated))
	require.NoError(state.WriteValidatorMetadata(c, primaryDB, netDB, CodecVersion1))
	primaryDBHas, err = primaryDB.Has(testUptimeReward2.txID[:])
	require.NoError(err)
	require.True(primaryDBHas)
}

func TestParseValidatorMetadata(t *testing.T) {
	c := pvmcodectest.NewMetadataCodec()
	type test struct {
		name        string
		bytes       []byte
		expected    *validatorMetadata
		expectErr   bool
	}
	tests := []test{
		{
			name:  "nil",
			bytes: nil,
			expected: &validatorMetadata{
				lastUpdated: time.Unix(0, 0),
			},
		},
		{
			name: "potential reward only",
			bytes: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xA0,
			},
			expected: &validatorMetadata{
				PotentialReward: 100000,
				lastUpdated:     time.Unix(0, 0),
			},
		},
		{
			name: "pre-delegatee reward",
			bytes: []byte{
				// codec version
				0x00, 0x00,
				// up duration
				0x00, 0x00, 0x00, 0x00, 0x00, 0x5B, 0x8D, 0x80,
				// last updated
				0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0xBB, 0xA0,
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xA0,
			},
			expected: &validatorMetadata{
				UpDuration:      time.Duration(6000000),
				LastUpdated:     900000,
				PotentialReward: 100000,
				lastUpdated:     time.Unix(900000, 0),
			},
		},
		{
			name: "potential delegatee reward",
			bytes: []byte{
				// codec version
				0x00, 0x00,
				// up duration
				0x00, 0x00, 0x00, 0x00, 0x00, 0x5B, 0x8D, 0x80,
				// last updated
				0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0xBB, 0xA0,
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xA0,
				// potential delegatee reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x4E, 0x20,
			},
			expected: &validatorMetadata{
				UpDuration:               time.Duration(6000000),
				LastUpdated:              900000,
				PotentialReward:          100000,
				PotentialDelegateeReward: 20000,
				lastUpdated:              time.Unix(900000, 0),
			},
		},
		{
			name: "invalid codec version",
			bytes: []byte{
				// codec version
				0x00, 0x02,
				// up duration
				0x00, 0x00, 0x00, 0x00, 0x00, 0x5B, 0x8D, 0x80,
				// last updated
				0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0xBB, 0xA0,
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xA0,
				// potential delegatee reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x4E, 0x20,
			},
			expected:  nil,
			expectErr: true,
		},
		{
			name: "short byte len",
			bytes: []byte{
				// codec version
				0x00, 0x00,
				// up duration
				0x00, 0x00, 0x00, 0x00, 0x00, 0x5B, 0x8D, 0x80,
				// last updated
				0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0xBB, 0xA0,
				// potential reward
				0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xA0,
				// potential delegatee reward
				0x00, 0x00, 0x00, 0x00, 0x4E, 0x20,
			},
			expected:  nil,
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			var metadata validatorMetadata
			err := parseValidatorMetadata(c, tt.bytes, &metadata)
			if tt.expectErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			require.Equal(tt.expected, &metadata)
		})
	}
}
