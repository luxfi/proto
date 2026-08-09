// Warp, X-chain and genesis conformance. Same discipline as wire_test.go: a
// value is encoded through proto and handed to the chain's own parser. Warp
// additionally asserts the message ID, because signature aggregation keys on
// it — bytes that parse but hash differently still cannot reach a quorum.
package conformance

import (
	"bytes"
	"testing"

	"github.com/luxfi/crypto/secp256k1"
	"github.com/luxfi/ids"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"
	"github.com/luxfi/utxo/wire"

	nodewarp "github.com/luxfi/node/vms/platformvm/warp"
	nodewarpmsg "github.com/luxfi/node/vms/platformvm/warp/message"
	nodewarppayload "github.com/luxfi/node/vms/platformvm/warp/payload"
	nodextxs "github.com/luxfi/node/vms/xvm/txs"

	protowarp "github.com/luxfi/proto/p/warp"
	protowarpmsg "github.com/luxfi/proto/p/warp/message"
	protowarppayload "github.com/luxfi/proto/p/warp/payload"
	protoxtxs "github.com/luxfi/proto/x/txs"
)

func TestWarpWire(t *testing.T) {
	t.Run("AddressedCall", func(t *testing.T) {
		src, payload := []byte{0xaa, 0xbb}, []byte{0x01, 0x02, 0x03}

		got, err := protowarppayload.NewAddressedCall(src, payload)
		if err != nil {
			t.Fatal(err)
		}
		want, err := nodewarppayload.NewAddressedCall(src, payload)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.Bytes(), want.Bytes()) {
			t.Fatalf("bytes differ\n proto %x\n node  %x", got.Bytes(), want.Bytes())
		}
		parsed, err := nodewarppayload.ParseAddressedCall(got.Bytes())
		if err != nil {
			t.Fatalf("node rejected proto bytes: %v", err)
		}
		if !bytes.Equal(parsed.Payload, payload) || !bytes.Equal(parsed.SourceAddress, src) {
			t.Fatalf("round trip lost fields: %+v", parsed)
		}
		// A parser that accepts anything proves nothing.
		if _, err := nodewarppayload.ParseAddressedCall([]byte{0, 0, 1, 0, 0, 0, 4}); err == nil {
			t.Fatal("node accepted bytes that are not this wire")
		}
	})

	t.Run("UnsignedMessage", func(t *testing.T) {
		chainID := ids.GenerateTestID()
		payload := []byte("payload")

		got, err := protowarp.NewUnsignedMessage(9, chainID, payload)
		if err != nil {
			t.Fatal(err)
		}
		want, err := nodewarp.NewUnsignedMessage(9, chainID, payload)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.Bytes(), want.Bytes()) {
			t.Fatalf("bytes differ\n proto %x\n node  %x", got.Bytes(), want.Bytes())
		}
		parsed, err := nodewarp.ParseUnsignedMessage(got.Bytes())
		if err != nil {
			t.Fatalf("node rejected proto bytes: %v", err)
		}
		// The ID is what aggregation keys on.
		if parsed.ID() != got.ID() {
			t.Fatalf("message ID diverges: proto %s node %s", got.ID(), parsed.ID())
		}
	})

	t.Run("RegisterL1Validator", func(t *testing.T) {
		chainID, nodeID := ids.GenerateTestID(), ids.GenerateTestNodeID()
		var pk [48]byte
		pk[0] = 0x91

		got, err := protowarpmsg.NewRegisterL1Validator(
			chainID, nodeID, pk, 1234,
			protowarpmsg.PChainOwner{Threshold: 1, Addresses: []ids.ShortID{{0x01}}},
			protowarpmsg.PChainOwner{},
			77,
		)
		if err != nil {
			t.Fatal(err)
		}
		want, err := nodewarpmsg.NewRegisterL1Validator(
			chainID, nodeID, pk, 1234,
			nodewarpmsg.PChainOwner{Threshold: 1, Addresses: []ids.ShortID{{0x01}}},
			nodewarpmsg.PChainOwner{},
			77,
		)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.Bytes(), want.Bytes()) {
			t.Fatalf("bytes differ\n proto %x\n node  %x", got.Bytes(), want.Bytes())
		}
		parsed, err := nodewarpmsg.ParseRegisterL1Validator(got.Bytes())
		if err != nil {
			t.Fatalf("node rejected proto bytes: %v", err)
		}
		// ValidationID is hash(bytes) and names the validator on chain.
		if parsed.ValidationID() != got.ValidationID() {
			t.Fatalf("validation ID diverges: proto %s node %s", got.ValidationID(), parsed.ValidationID())
		}
	})
}

func TestXChainWire(t *testing.T) {
	key, err := secp256k1.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	blockchainID, assetID, inTxID := ids.GenerateTestID(), ids.GenerateTestID(), ids.GenerateTestID()

	base := func() lux.BaseTx {
		return lux.BaseTx{
			NetworkID:    5,
			BlockchainID: blockchainID,
			Ins: []*lux.TransferableInput{{
				UTXOID: lux.UTXOID{TxID: inTxID},
				Asset:  lux.Asset{ID: assetID},
				In: &secp256k1fx.TransferInput{
					Amt:   1000,
					Input: secp256k1fx.Input{SigIndices: []uint32{0}},
				},
			}},
			Outs: []*lux.TransferableOutput{{
				Asset: lux.Asset{ID: assetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 900,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{key.Address()},
					},
				},
			}},
		}
	}

	t.Run("BaseTxUnsigned", func(t *testing.T) {
		got, err := protoxtxs.UnsignedBytes(&protoxtxs.BaseTx{BaseTx: base()})
		if err != nil {
			t.Fatal(err)
		}
		want, err := nodextxs.UnsignedBytes(&nodextxs.BaseTx{BaseTx: base()})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("bytes differ\n proto %x\n node  %x", got, want)
		}
	})

	t.Run("BaseTxSigned", func(t *testing.T) {
		tx := &protoxtxs.Tx{Unsigned: &protoxtxs.BaseTx{BaseTx: base()}}
		if err := tx.SignSECP256K1Fx([][]*secp256k1.PrivateKey{{key}}); err != nil {
			t.Fatal(err)
		}
		signed := tx.Bytes()

		// The envelope the chain's parser opens first.
		if _, err := wire.WrapSignedTx(signed); err != nil {
			t.Fatalf("node envelope rejected proto bytes: %v", err)
		}
		parsed, err := nodextxs.Parse(signed)
		if err != nil {
			t.Fatalf("node rejected proto bytes: %v", err)
		}
		if parsed.ID() != tx.ID() {
			t.Fatalf("tx ID diverges: proto %s node %s", tx.ID(), parsed.ID())
		}
		if len(parsed.Creds) != 1 {
			t.Fatalf("credentials lost: got %d want 1", len(parsed.Creds))
		}
		if _, err := nodextxs.Parse([]byte{0, 0, 0, 0, 0, 0, 0x71, 0x78}); err == nil {
			t.Fatal("node accepted bytes that are not this wire")
		}
	})
}
