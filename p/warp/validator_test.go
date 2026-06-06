// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package warp_test

import (
	"context"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/luxfi/constants"
	"github.com/luxfi/crypto/bls"
	"github.com/luxfi/crypto/bls/signer/localsigner"
	"github.com/luxfi/ids"
	"github.com/luxfi/math/set"
	"github.com/luxfi/metric"
	"github.com/luxfi/proto/p/warp"
	"github.com/luxfi/upgrade"
	validators "github.com/luxfi/validators"
	"github.com/luxfi/validators/validatorsmock"
)

var (
	chainID = ids.GenerateTestID()
)

// testValidatorStateAdapter wraps validators.State to implement ValidatorState
// converting GetValidatorOutput to ValidatorData
type testValidatorStateAdapter struct {
	validators.State
}

func (t *testValidatorStateAdapter) GetValidatorSet(ctx context.Context, height uint64, chainID ids.ID) (map[ids.NodeID]*warp.ValidatorData, error) {
	validatorSet, err := t.State.GetValidatorSet(ctx, height, chainID)
	if err != nil {
		return nil, err
	}

	result := make(map[ids.NodeID]*warp.ValidatorData, len(validatorSet))
	for nodeID, validator := range validatorSet {
		result[nodeID] = &warp.ValidatorData{
			NodeID:    validator.NodeID,
			PublicKey: validator.PublicKey,
			Weight:    validator.Weight,
		}
	}
	return result, nil
}

func TestGetCanonicalValidatorSet(t *testing.T) {
	type test struct {
		name           string
		stateF         func(*gomock.Controller) validators.State
		expectedVdrs   []*warp.Validator
		expectedWeight uint64
		expectedErr    error
	}

	tests := []test{
		{
			name: "can't get validator set",
			stateF: func(ctrl *gomock.Controller) validators.State {
				state := validatorsmock.NewState(ctrl)
				state.EXPECT().GetValidatorSet(gomock.Any(), pChainHeight, chainID).Return(nil, errTest)
				return state
			},
			expectedErr: errTest,
		},
		{
			name: "all validators have public keys; no duplicate pub keys",
			stateF: func(ctrl *gomock.Controller) validators.State {
				state := validatorsmock.NewState(ctrl)
				state.EXPECT().GetValidatorSet(gomock.Any(), pChainHeight, chainID).Return(
					map[ids.NodeID]*validators.GetValidatorOutput{
						testVdrs[0].nodeID: {
							NodeID:    testVdrs[0].nodeID,
							PublicKey: bls.PublicKeyToUncompressedBytes(testVdrs[0].vdr.PublicKey),
							Weight:    testVdrs[0].vdr.Weight,
						},
						testVdrs[1].nodeID: {
							NodeID:    testVdrs[1].nodeID,
							PublicKey: bls.PublicKeyToUncompressedBytes(testVdrs[1].vdr.PublicKey),
							Weight:    testVdrs[1].vdr.Weight,
						},
					},
					nil,
				)
				return state
			},
			expectedVdrs:   []*warp.Validator{testVdrs[0].vdr, testVdrs[1].vdr},
			expectedWeight: 6,
			expectedErr:    nil,
		},
		{
			name: "all validators have public keys; duplicate pub keys",
			stateF: func(ctrl *gomock.Controller) validators.State {
				state := validatorsmock.NewState(ctrl)
				state.EXPECT().GetValidatorSet(gomock.Any(), pChainHeight, chainID).Return(
					map[ids.NodeID]*validators.GetValidatorOutput{
						testVdrs[0].nodeID: {
							NodeID:    testVdrs[0].nodeID,
							PublicKey: bls.PublicKeyToUncompressedBytes(testVdrs[0].vdr.PublicKey),
							Weight:    testVdrs[0].vdr.Weight,
						},
						testVdrs[1].nodeID: {
							NodeID:    testVdrs[1].nodeID,
							PublicKey: bls.PublicKeyToUncompressedBytes(testVdrs[1].vdr.PublicKey),
							Weight:    testVdrs[1].vdr.Weight,
						},
						testVdrs[2].nodeID: {
							NodeID:    testVdrs[2].nodeID,
							PublicKey: bls.PublicKeyToUncompressedBytes(testVdrs[0].vdr.PublicKey),
							Weight:    testVdrs[0].vdr.Weight,
						},
					},
					nil,
				)
				return state
			},
			expectedVdrs: []*warp.Validator{
				{
					PublicKey:      testVdrs[0].vdr.PublicKey,
					PublicKeyBytes: testVdrs[0].vdr.PublicKeyBytes,
					Weight:         testVdrs[0].vdr.Weight * 2,
					NodeIDs: []ids.NodeID{
						testVdrs[0].nodeID,
						testVdrs[2].nodeID,
					},
				},
				testVdrs[1].vdr,
			},
			expectedWeight: 9,
			expectedErr:    nil,
		},
		{
			name: "validator without public key; no duplicate pub keys",
			stateF: func(ctrl *gomock.Controller) validators.State {
				state := validatorsmock.NewState(ctrl)
				state.EXPECT().GetValidatorSet(gomock.Any(), pChainHeight, chainID).Return(
					map[ids.NodeID]*validators.GetValidatorOutput{
						testVdrs[0].nodeID: {
							NodeID:    testVdrs[0].nodeID,
							PublicKey: nil,
							Weight:    testVdrs[0].vdr.Weight,
						},
						testVdrs[1].nodeID: {
							NodeID:    testVdrs[1].nodeID,
							PublicKey: bls.PublicKeyToUncompressedBytes(testVdrs[1].vdr.PublicKey),
							Weight:    testVdrs[1].vdr.Weight,
						},
					},
					nil,
				)
				return state
			},
			expectedVdrs:   []*warp.Validator{testVdrs[1].vdr},
			expectedWeight: 6,
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			ctrl := gomock.NewController(t)

			state := tt.stateF(ctrl)
			// Wrap validators.State to implement ValidatorState
			wrappedState := &testValidatorStateAdapter{
				State: state,
			}

			validators, err := warp.GetCanonicalValidatorSetFromSubchainID(t.Context(), wrappedState, pChainHeight, chainID)
			require.ErrorIs(err, tt.expectedErr)
			if err != nil {
				return
			}
			require.Equal(tt.expectedWeight, validators.TotalWeight)

			// These are pointers so have to test equality like this
			require.Len(validators.Validators, len(tt.expectedVdrs))
			for i, expectedVdr := range tt.expectedVdrs {
				gotVdr := validators.Validators[i]
				expectedPKBytes := bls.PublicKeyToUncompressedBytes(expectedVdr.PublicKey)
				gotPKBytes := bls.PublicKeyToUncompressedBytes(gotVdr.PublicKey)
				require.Equal(expectedPKBytes, gotPKBytes)
				require.Equal(expectedVdr.PublicKeyBytes, gotVdr.PublicKeyBytes)
				require.Equal(expectedVdr.Weight, gotVdr.Weight)
				require.ElementsMatch(expectedVdr.NodeIDs, gotVdr.NodeIDs)
			}
		})
	}
}

func TestFilterValidators(t *testing.T) {
	sk0, err := localsigner.New()
	require.NoError(t, err)
	pk0 := sk0.PublicKey()
	vdr0 := &warp.Validator{
		PublicKey:      pk0,
		PublicKeyBytes: bls.PublicKeyToUncompressedBytes(pk0),
		Weight:         1,
	}

	sk1, err := localsigner.New()
	require.NoError(t, err)
	pk1 := sk1.PublicKey()
	vdr1 := &warp.Validator{
		PublicKey:      pk1,
		PublicKeyBytes: bls.PublicKeyToUncompressedBytes(pk1),
		Weight:         2,
	}

	type test struct {
		name         string
		indices      set.Bits
		vdrs         []*warp.Validator
		expectedVdrs []*warp.Validator
		expectedErr  error
	}

	tests := []test{
		{
			name:         "empty",
			indices:      set.NewBits(),
			vdrs:         []*warp.Validator{},
			expectedVdrs: []*warp.Validator{},
			expectedErr:  nil,
		},
		{
			name:        "unknown validator",
			indices:     set.NewBits(2),
			vdrs:        []*warp.Validator{vdr0, vdr1},
			expectedErr: warp.ErrUnknownValidator,
		},
		{
			name:    "two filtered out",
			indices: set.NewBits(),
			vdrs: []*warp.Validator{
				vdr0,
				vdr1,
			},
			expectedVdrs: []*warp.Validator{},
			expectedErr:  nil,
		},
		{
			name:    "one filtered out",
			indices: set.NewBits(1),
			vdrs: []*warp.Validator{
				vdr0,
				vdr1,
			},
			expectedVdrs: []*warp.Validator{
				vdr1,
			},
			expectedErr: nil,
		},
		{
			name:    "none filtered out",
			indices: set.NewBits(0, 1),
			vdrs: []*warp.Validator{
				vdr0,
				vdr1,
			},
			expectedVdrs: []*warp.Validator{
				vdr0,
				vdr1,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			vdrs, err := warp.FilterValidators(tt.indices, tt.vdrs)
			require.ErrorIs(err, tt.expectedErr)
			if tt.expectedErr != nil {
				return
			}
			require.Equal(tt.expectedVdrs, vdrs)
		})
	}
}

func TestSumWeight(t *testing.T) {
	vdr0 := &warp.Validator{
		Weight: 1,
	}
	vdr1 := &warp.Validator{
		Weight: 2,
	}
	vdr2 := &warp.Validator{
		Weight: math.MaxUint64,
	}

	type test struct {
		name        string
		vdrs        []*warp.Validator
		expectedSum uint64
		expectedErr error
	}

	tests := []test{
		{
			name:        "empty",
			vdrs:        []*warp.Validator{},
			expectedSum: 0,
		},
		{
			name:        "one",
			vdrs:        []*warp.Validator{vdr0},
			expectedSum: 1,
		},
		{
			name:        "two",
			vdrs:        []*warp.Validator{vdr0, vdr1},
			expectedSum: 3,
		},
		{
			name:        "overflow",
			vdrs:        []*warp.Validator{vdr0, vdr2},
			expectedErr: warp.ErrWeightOverflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			sum, err := warp.SumWeight(tt.vdrs)
			require.ErrorIs(err, tt.expectedErr)
			if tt.expectedErr != nil {
				return
			}
			require.Equal(tt.expectedSum, sum)
		})
	}
}

func BenchmarkGetCanonicalValidatorSet(b *testing.B) {
	pChainHeight := uint64(1)
	chainID := ids.GenerateTestID()
	numNodes := 10_000
	getValidatorOutputs := make([]*validators.GetValidatorOutput, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		nodeID := ids.GenerateTestNodeID()
		blsPrivateKey, err := localsigner.New()
		require.NoError(b, err)
		blsPublicKey := blsPrivateKey.PublicKey()
		getValidatorOutputs = append(getValidatorOutputs, &validators.GetValidatorOutput{
			NodeID:    nodeID,
			PublicKey: bls.PublicKeyToUncompressedBytes(blsPublicKey),
			Weight:    20,
		})
	}

	for _, size := range []int{0, 1, 10, 100, 1_000, 10_000} {
		getValidatorsOutput := make(map[ids.NodeID]*validators.GetValidatorOutput)
		for i := 0; i < size; i++ {
			validator := getValidatorOutputs[i]
			getValidatorsOutput[validator.NodeID] = validator
		}
		// Create a simple validator state for benchmarking
		wrappedState := newMockValidatorState(
			func() map[ids.NodeID]*warp.ValidatorData {
				result := make(map[ids.NodeID]*warp.ValidatorData, len(getValidatorsOutput))
				for nodeID, vdr := range getValidatorsOutput {
					result[nodeID] = &warp.ValidatorData{
						NodeID:    vdr.NodeID,
						PublicKey: vdr.PublicKey,
						Weight:    vdr.Weight,
					}
				}
				return result
			}(),
			nil,
		)

		b.Run(strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := warp.GetCanonicalValidatorSetFromSubchainID(b.Context(), wrappedState, pChainHeight, chainID)
				require.NoError(b, err)
			}
		})
	}
}

// mockValidatorState is a test mock that tracks call counts
type mockValidatorState struct {
	callCount int
	data      map[ids.NodeID]*warp.ValidatorData
	err       error
}

func (m *mockValidatorState) GetValidatorSet(ctx context.Context, height uint64, chainID ids.ID) (map[ids.NodeID]*warp.ValidatorData, error) {
	m.callCount++
	return m.data, m.err
}

func newMockValidatorState(data map[ids.NodeID]*warp.ValidatorData, err error) *mockValidatorState {
	return &mockValidatorState{data: data, err: err}
}

func TestCachedValidatorState(t *testing.T) {
	ctx := context.Background()
	height := uint64(100)
	chain1 := ids.GenerateTestID()
	chain2 := ids.GenerateTestID()

	// Create test validator data
	nodeID1 := ids.GenerateTestNodeID()
	nodeID2 := ids.GenerateTestNodeID()
	testData := map[ids.NodeID]*warp.ValidatorData{
		nodeID1: {
			NodeID:    nodeID1,
			PublicKey: bls.PublicKeyToUncompressedBytes(testVdrs[0].vdr.PublicKey),
			Weight:    100,
		},
		nodeID2: {
			NodeID:    nodeID2,
			PublicKey: bls.PublicKeyToUncompressedBytes(testVdrs[1].vdr.PublicKey),
			Weight:    200,
		},
	}

	type test struct {
		name              string
		state             *mockValidatorState
		upgradeConfig     *upgrade.Config
		networkID         uint32
		expectedCallCount int
		operations        func(*testing.T, *warp.CachedValidatorState)
	}

	tests := []test{
		{
			name:              "pre-Granite no caching",
			state:             newMockValidatorState(testData, nil),
			upgradeConfig:     &upgrade.Config{GraniteTime: time.Now().Add(1 * time.Hour)},
			networkID:         constants.MainnetID,
			expectedCallCount: 2, // Should call underlying state twice (no caching)
			operations: func(t *testing.T, cached *warp.CachedValidatorState) {
				vdrs1, err := cached.GetValidatorSet(ctx, height, chain1)
				require.NoError(t, err)
				require.Equal(t, testData, vdrs1)

				vdrs2, err := cached.GetValidatorSet(ctx, height, chain1)
				require.NoError(t, err)
				require.Equal(t, testData, vdrs2)
			},
		},
		{
			name:              "post-Granite with caching",
			state:             newMockValidatorState(testData, nil),
			upgradeConfig:     &upgrade.Config{GraniteTime: time.Now().Add(-1 * time.Hour)},
			networkID:         constants.MainnetID,
			expectedCallCount: 1, // Should call underlying state once, then use cache
			operations: func(t *testing.T, cached *warp.CachedValidatorState) {
				vdrs1, err := cached.GetValidatorSet(ctx, height, chain1)
				require.NoError(t, err)
				require.Equal(t, testData, vdrs1)

				vdrs2, err := cached.GetValidatorSet(ctx, height, chain1)
				require.NoError(t, err)
				require.Equal(t, testData, vdrs2)
			},
		},
		{
			name:              "different heights cached separately",
			state:             newMockValidatorState(testData, nil),
			upgradeConfig:     &upgrade.Config{GraniteTime: time.Now().Add(-1 * time.Hour)},
			networkID:         constants.MainnetID,
			expectedCallCount: 2, // Two different heights = two calls
			operations: func(t *testing.T, cached *warp.CachedValidatorState) {
				vdrs1, err := cached.GetValidatorSet(ctx, height, chain1)
				require.NoError(t, err)
				require.Equal(t, testData, vdrs1)

				vdrs2, err := cached.GetValidatorSet(ctx, height+1, chain1)
				require.NoError(t, err)
				require.Equal(t, testData, vdrs2)
			},
		},
		{
			name:              "different chains cached separately",
			state:             newMockValidatorState(testData, nil),
			upgradeConfig:     &upgrade.Config{GraniteTime: time.Now().Add(-1 * time.Hour)},
			networkID:         constants.MainnetID,
			expectedCallCount: 2, // Two different chains = two calls
			operations: func(t *testing.T, cached *warp.CachedValidatorState) {
				vdrs1, err := cached.GetValidatorSet(ctx, height, chain1)
				require.NoError(t, err)
				require.Equal(t, testData, vdrs1)

				vdrs2, err := cached.GetValidatorSet(ctx, height, chain2)
				require.NoError(t, err)
				require.Equal(t, testData, vdrs2)
			},
		},
		{
			name:              "error propagates without caching",
			state:             newMockValidatorState(nil, errTest),
			upgradeConfig:     &upgrade.Config{GraniteTime: time.Now().Add(-1 * time.Hour)},
			networkID:         constants.MainnetID,
			expectedCallCount: 1,
			operations: func(t *testing.T, cached *warp.CachedValidatorState) {
				_, err := cached.GetValidatorSet(ctx, height, chain1)
				require.ErrorIs(t, err, errTest)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			registerer := metric.NewRegistry()

			cached, err := warp.NewCachedValidatorState(tt.state, tt.upgradeConfig, tt.networkID, registerer)
			require.NoError(err)
			require.NotNil(cached)

			// Run test operations
			tt.operations(t, cached)

			// Verify call count
			require.Equal(tt.expectedCallCount, tt.state.callCount, "unexpected number of calls to underlying state")
		})
	}
}
