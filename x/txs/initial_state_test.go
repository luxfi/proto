// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/xcodectest"
	"github.com/luxfi/proto/x/txs"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"
	"github.com/luxfi/vm/components/verify"
)

var errTest = errors.New("non-nil error")

func TestInitialStateVerifySerialization(t *testing.T) {
	require := require.New(t)

	m, c := xcodectest.NewRuntimeCodec()
	require.NoError(c.RegisterType(&secp256k1fx.TransferOutput{}))

	expected := []byte{
		// Codec version:
		0x00, 0x00,
		// fxID:
		0x00, 0x00, 0x00, 0x00,
		// num outputs:
		0x00, 0x00, 0x00, 0x01,
		// output:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x30, 0x39, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0xd4, 0x31, 0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x02, 0x51, 0x02, 0x5c, 0x61,
		0xfb, 0xcf, 0xc0, 0x78, 0xf6, 0x93, 0x34, 0xf8,
		0x34, 0xbe, 0x6d, 0xd2, 0x6d, 0x55, 0xa9, 0x55,
		0xc3, 0x34, 0x41, 0x28, 0xe0, 0x60, 0x12, 0x8e,
		0xde, 0x35, 0x23, 0xa2, 0x4a, 0x46, 0x1c, 0x89,
		0x43, 0xab, 0x08, 0x59,
	}

	is := &txs.InitialState{
		FxIndex: 0,
		Outs: []verify.State{
			&secp256k1fx.TransferOutput{
				Amt: 12345,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  54321,
					Threshold: 1,
					Addrs: []ids.ShortID{
						{
							0x51, 0x02, 0x5c, 0x61, 0xfb, 0xcf, 0xc0, 0x78,
							0xf6, 0x93, 0x34, 0xf8, 0x34, 0xbe, 0x6d, 0xd2,
							0x6d, 0x55, 0xa9, 0x55,
						},
						{
							0xc3, 0x34, 0x41, 0x28, 0xe0, 0x60, 0x12, 0x8e,
							0xde, 0x35, 0x23, 0xa2, 0x4a, 0x46, 0x1c, 0x89,
							0x43, 0xab, 0x08, 0x59,
						},
					},
				},
			},
		},
	}

	isBytes, err := m.Marshal(txs.CodecVersion, is)
	require.NoError(err)
	require.Equal(expected, isBytes)
}

func TestInitialStateVerifyNil(t *testing.T) {
	require := require.New(t)

	m, _ := xcodectest.NewRuntimeCodec()
	numFxs := 1

	is := (*txs.InitialState)(nil)
	err := is.Verify(m, numFxs)
	require.ErrorIs(err, txs.ErrNilInitialState)
}

func TestInitialStateVerifyUnknownFxID(t *testing.T) {
	require := require.New(t)

	m, _ := xcodectest.NewRuntimeCodec()
	numFxs := 1

	is := txs.InitialState{
		FxIndex: 1,
	}
	err := is.Verify(m, numFxs)
	require.ErrorIs(err, txs.ErrUnknownFx)
}

func TestInitialStateVerifyNilOutput(t *testing.T) {
	require := require.New(t)

	m, _ := xcodectest.NewRuntimeCodec()
	numFxs := 1

	is := txs.InitialState{
		FxIndex: 0,
		Outs:    []verify.State{nil},
	}
	err := is.Verify(m, numFxs)
	require.ErrorIs(err, txs.ErrNilFxOutput)
}

func TestInitialStateVerifyInvalidOutput(t *testing.T) {
	require := require.New(t)

	m, c := xcodectest.NewRuntimeCodec()
	require.NoError(c.RegisterType(&lux.TestState{}))
	numFxs := 1

	is := txs.InitialState{
		FxIndex: 0,
		Outs:    []verify.State{&lux.TestState{Err: errTest}},
	}
	err := is.Verify(m, numFxs)
	require.ErrorIs(err, errTest)
}

func TestInitialStateVerifyUnsortedOutputs(t *testing.T) {
	require := require.New(t)

	m, c := xcodectest.NewRuntimeCodec()
	require.NoError(c.RegisterType(&lux.TestTransferable{}))
	numFxs := 1

	is := txs.InitialState{
		FxIndex: 0,
		Outs: []verify.State{
			&lux.TestTransferable{Val: 1},
			&lux.TestTransferable{Val: 0},
		},
	}
	err := is.Verify(m, numFxs)
	require.ErrorIs(err, txs.ErrOutputsNotSorted)
	is.Sort(m)
	require.NoError(is.Verify(m, numFxs))
}

func TestInitialStateCompare(t *testing.T) {
	tests := []struct {
		a        *txs.InitialState
		b        *txs.InitialState
		expected int
	}{
		{
			a:        &txs.InitialState{},
			b:        &txs.InitialState{},
			expected: 0,
		},
		{
			a: &txs.InitialState{
				FxIndex: 1,
			},
			b:        &txs.InitialState{},
			expected: 1,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%d_%d_%d", test.a.FxIndex, test.b.FxIndex, test.expected), func(t *testing.T) {
			require := require.New(t)

			require.Equal(test.expected, test.a.Compare(test.b))
			require.Equal(-test.expected, test.b.Compare(test.a))
		})
	}
}
