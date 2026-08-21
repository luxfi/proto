package conformance

import (
	"testing"

	"github.com/luxfi/ids"
	nodestakeable "github.com/luxfi/node/vms/platformvm/stakeable"
	protostakeable "github.com/luxfi/proto/p/stakeable"
	"github.com/luxfi/utxo/secp256k1fx"
)

// A stakeable-locked output is serialized by proto when the wallet BUILDS a tx
// and by the node when it PARSES one. The canonical output sort compares
// exactly those bytes, so a second encoding is not a cosmetic difference: the
// two sides order the outputs differently and the chain rejects the tx with
// "outputs not sorted". That is what happened to the first real
// AddPermissionlessDelegatorTx issued against mainnet — proto emitted an ad-hoc
// 0xFF||locktime||inner marker while the chain emits the shared
// wire.LockedOutput envelope.
//
// TestWire's fixtures are all unlocked outputs, so nothing covered this.
func TestLockOutWireMatchesChain(t *testing.T) {
	inner := &secp256k1fx.TransferOutput{
		Amt: 4998000000000000,
		OutputOwners: secp256k1fx.OutputOwners{
			Threshold: 1,
			Addrs:     []ids.ShortID{{1, 2, 3}},
		},
	}
	const locktime = uint64(0x90cc4f80)

	proto := (&protostakeable.LockOut{Locktime: locktime, TransferableOut: inner}).Bytes()
	node := (&nodestakeable.LockOut{Locktime: locktime, TransferableOut: inner}).Bytes()

	t.Logf("proto len=%d %x", len(proto), proto)
	t.Logf("node  len=%d %x", len(node), node)

	if string(proto) != string(node) {
		t.Fatalf("LockOut wire fork: proto %d bytes, node %d bytes", len(proto), len(node))
	}

	// Positive control: the comparison can fail. A different locktime must
	// produce different bytes, otherwise the check above proves nothing.
	other := (&nodestakeable.LockOut{Locktime: locktime + 1, TransferableOut: inner}).Bytes()
	if string(node) == string(other) {
		t.Fatal("control failed: locktime is not reflected in the wire bytes")
	}
}
