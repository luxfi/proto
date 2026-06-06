// Copyright (C) 2019-2026, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

//go:build emit_fixtures

package txs_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/luxfi/crypto/secp256k1"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/xcodectest"
	"github.com/luxfi/proto/x/fxs"
	"github.com/luxfi/proto/x/txs"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"
)

// TestEmitTxCredBytes emits the credBytes portion of the signed Import/Export
// transactions used by TestImportTxSerialization and TestExportTxSerialization.
// The credBytes are signed bytes, deterministically derived from
// SignSECP256K1Fx — they change whenever the wire format the signer
// signs over changes.
//
// Run via:
//
//	go test -tags emit_fixtures -run TestEmitTxCredBytes -v ./x/txs
//
// then paste the hex output into the respective credBytes literal in
// import_tx_test.go / export_tx_test.go.
func TestEmitTxCredBytes(t *testing.T) {
	parser, err := txs.NewParser(
		xcodectest.New(),
		[]fxs.Fx{&secp256k1fx.Fx{}},
	)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	// Use the same shared TestKeys() that the package's *_test.go
	// files use (defined in base_tx_test.go as `keys = secp256k1.TestKeys()`).
	allKeys := secp256k1.TestKeys()
	k := allKeys[0]

	type emit struct {
		name string
		tx   *txs.Tx
	}

	cases := []emit{
		{
			name: "ImportTx",
			tx: &txs.Tx{Unsigned: &txs.ImportTx{
				BaseTx: txs.BaseTx{BaseTx: lux.BaseTx{
					NetworkID: 2,
					BlockchainID: ids.ID{
						0xff, 0xff, 0xff, 0xff, 0xee, 0xee, 0xee, 0xee,
						0xdd, 0xdd, 0xdd, 0xdd, 0xcc, 0xcc, 0xcc, 0xcc,
						0xbb, 0xbb, 0xbb, 0xbb, 0xaa, 0xaa, 0xaa, 0xaa,
						0x99, 0x99, 0x99, 0x99, 0x88, 0x88, 0x88, 0x88,
					},
					Memo: []byte{0x00, 0x01, 0x02, 0x03},
				}},
				SourceChain: ids.ID{
					0x1f, 0x8f, 0x9f, 0x0f, 0x1e, 0x8e, 0x9e, 0x0e,
					0x2d, 0x7d, 0xad, 0xfd, 0x2c, 0x7c, 0xac, 0xfc,
					0x3b, 0x6b, 0xbb, 0xeb, 0x3a, 0x6a, 0xba, 0xea,
					0x49, 0x59, 0xc9, 0xd9, 0x48, 0x58, 0xc8, 0xd8,
				},
				ImportedIns: []*lux.TransferableInput{{
					UTXOID: lux.UTXOID{TxID: ids.ID{
						0x0f, 0x2f, 0x4f, 0x6f, 0x8e, 0xae, 0xce, 0xee,
						0x0d, 0x2d, 0x4d, 0x6d, 0x8c, 0xac, 0xcc, 0xec,
						0x0b, 0x2b, 0x4b, 0x6b, 0x8a, 0xaa, 0xca, 0xea,
						0x09, 0x29, 0x49, 0x69, 0x88, 0xa8, 0xc8, 0xe8,
					}},
					Asset: lux.Asset{ID: ids.ID{
						0x1f, 0x3f, 0x5f, 0x7f, 0x9e, 0xbe, 0xde, 0xfe,
						0x1d, 0x3d, 0x5d, 0x7d, 0x9c, 0xbc, 0xdc, 0xfc,
						0x1b, 0x3b, 0x5b, 0x7b, 0x9a, 0xba, 0xda, 0xfa,
						0x19, 0x39, 0x59, 0x79, 0x98, 0xb8, 0xd8, 0xf8,
					}},
					In: &secp256k1fx.TransferInput{
						Amt:   1000,
						Input: secp256k1fx.Input{SigIndices: []uint32{0}},
					},
				}},
			}},
		},
		{
			name: "ExportTx",
			tx: &txs.Tx{Unsigned: &txs.ExportTx{
				BaseTx: txs.BaseTx{BaseTx: lux.BaseTx{
					NetworkID: 2,
					BlockchainID: ids.ID{
						0xff, 0xff, 0xff, 0xff, 0xee, 0xee, 0xee, 0xee,
						0xdd, 0xdd, 0xdd, 0xdd, 0xcc, 0xcc, 0xcc, 0xcc,
						0xbb, 0xbb, 0xbb, 0xbb, 0xaa, 0xaa, 0xaa, 0xaa,
						0x99, 0x99, 0x99, 0x99, 0x88, 0x88, 0x88, 0x88,
					},
					Ins: []*lux.TransferableInput{{
						UTXOID: lux.UTXOID{TxID: ids.ID{
							0x0f, 0x2f, 0x4f, 0x6f, 0x8e, 0xae, 0xce, 0xee,
							0x0d, 0x2d, 0x4d, 0x6d, 0x8c, 0xac, 0xcc, 0xec,
							0x0b, 0x2b, 0x4b, 0x6b, 0x8a, 0xaa, 0xca, 0xea,
							0x09, 0x29, 0x49, 0x69, 0x88, 0xa8, 0xc8, 0xe8,
						}},
						Asset: lux.Asset{ID: ids.ID{
							0x1f, 0x3f, 0x5f, 0x7f, 0x9e, 0xbe, 0xde, 0xfe,
							0x1d, 0x3d, 0x5d, 0x7d, 0x9c, 0xbc, 0xdc, 0xfc,
							0x1b, 0x3b, 0x5b, 0x7b, 0x9a, 0xba, 0xda, 0xfa,
							0x19, 0x39, 0x59, 0x79, 0x98, 0xb8, 0xd8, 0xf8,
						}},
						In: &secp256k1fx.TransferInput{
							Amt:   1000,
							Input: secp256k1fx.Input{SigIndices: []uint32{0}},
						},
					}},
					Memo: []byte{0x00, 0x01, 0x02, 0x03},
				}},
				DestinationChain: ids.ID{
					0x1f, 0x8f, 0x9f, 0x0f, 0x1e, 0x8e, 0x9e, 0x0e,
					0x2d, 0x7d, 0xad, 0xfd, 0x2c, 0x7c, 0xac, 0xfc,
					0x3b, 0x6b, 0xbb, 0xeb, 0x3a, 0x6a, 0xba, 0xea,
					0x49, 0x59, 0xc9, 0xd9, 0x48, 0x58, 0xc8, 0xd8,
				},
			}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.tx.Initialize(parser.Codec()); err != nil {
				t.Fatalf("Initialize: %v", err)
			}
			unsignedLen := len(c.tx.Bytes())
			if err := c.tx.SignSECP256K1Fx(
				parser.Codec(),
				[][]*secp256k1.PrivateKey{{k, k}, {k, k}},
			); err != nil {
				t.Fatalf("SignSECP256K1Fx: %v", err)
			}
			signed := c.tx.Bytes()
			// credBytes = everything after the unsigned body. But the
			// signed tx also bumps the last "creds count" byte at
			// position (unsignedLen-1) from 0 to 2; the test mirrors
			// this with `expected[len(expected)-1] = 0x02`.
			credBytes := signed[unsignedLen:]
			fmt.Printf("EMIT_CREDBYTES: %s tx.ID=%s unsignedLen=%d signedLen=%d\n",
				c.name, c.tx.ID().String(), unsignedLen, len(signed))
			fmt.Printf("  credBytes hex: %s\n", hex.EncodeToString(credBytes))
			fmt.Printf("  credBytes go : %s\n", goByteLiteral(credBytes))
		})
	}
}

func goByteLiteral(b []byte) string {
	parts := make([]string, len(b))
	for i, v := range b {
		parts[i] = fmt.Sprintf("0x%02x", v)
	}
	return "[]byte{" + strings.Join(parts, ", ") + "}"
}
