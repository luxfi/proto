// Package conformance pins proto's P-Chain wire to the chain's. Each case
// builds one tx twice — once with proto, once with the node's own package —
// asserts the bytes are identical, then asserts the node's Parse accepts
// proto's signed bytes and recovers the right type and credentials. A stride
// or offset that drifts yields a silently empty list rather than an error, so
// this comparison is the only thing that catches it.
package conformance

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/luxfi/crypto/secp256k1"
	"github.com/luxfi/ids"
	lux "github.com/luxfi/utxo"
	"github.com/luxfi/utxo/secp256k1fx"

	nodefx "github.com/luxfi/node/vms/platformvm/fx"
	nodesec "github.com/luxfi/node/vms/platformvm/security"
	nodesigner "github.com/luxfi/node/vms/platformvm/signer"
	nodetxs "github.com/luxfi/node/vms/platformvm/txs"
	nodemsg "github.com/luxfi/node/vms/platformvm/warp/message"

	protosec "github.com/luxfi/proto/p/security"
	protosigner "github.com/luxfi/proto/p/signer"
	prototxs "github.com/luxfi/proto/p/txs"
	protomsg "github.com/luxfi/proto/p/warp/message"
)

// t is the active test; every case runs as a subtest of TestWire.
var t *testing.T

func check(name string, cond bool, detail string) {
	if !cond {
		t.Errorf("%s: %s", name, detail)
		return
	}
	t.Logf("%s: %s", name, detail)
}

func id(b byte) ids.ID {
	var i ids.ID
	for k := range i {
		i[k] = b
	}
	return i
}

func addr(b byte) ids.ShortID {
	var a ids.ShortID
	for k := range a {
		a[k] = b
	}
	return a
}

func nodeID(b byte) ids.NodeID {
	var n ids.NodeID
	for k := range n {
		n[k] = b
	}
	return n
}

// envelope returns the same spending envelope value both trees share.
func envelope() lux.BaseTx {
	return lux.BaseTx{
		NetworkID:    96369,
		BlockchainID: id(0x11),
		Outs: []*lux.TransferableOutput{{
			Asset: lux.Asset{ID: id(0x22)},
			Out: &secp256k1fx.TransferOutput{
				Amt: 7_000_000,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Threshold: 1,
					Addrs:     []ids.ShortID{addr(0x33)},
				},
			},
		}},
		Ins: []*lux.TransferableInput{{
			UTXOID: lux.UTXOID{TxID: id(0x44), OutputIndex: 3},
			Asset:  lux.Asset{ID: id(0x22)},
			In: &secp256k1fx.TransferInput{
				Amt:   9_000_000,
				Input: secp256k1fx.Input{SigIndices: []uint32{0}},
			},
		}},
		Memo: []byte("lux"),
	}
}

func owners() *secp256k1fx.OutputOwners {
	return &secp256k1fx.OutputOwners{Threshold: 1, Addrs: []ids.ShortID{addr(0x55)}}
}

// compare asserts proto's unsigned bytes equal the node's, then signs with
// proto and asserts the node's Parse accepts the result.
func compare(name string, protoUnsigned prototxs.UnsignedTx, nodeUnsigned nodetxs.UnsignedTx, key *secp256k1.PrivateKey) {
	got, err := prototxs.Marshal(protoUnsigned)
	if err != nil {
		check(name+"/marshal", false, err.Error())
		return
	}
	want := nodeUnsigned.Bytes()
	check(name+"/bytes", bytes.Equal(got, want),
		fmt.Sprintf("proto=%d node=%d bytes", len(got), len(want)))
	if !bytes.Equal(got, want) {
		for i := range got {
			if i >= len(want) || got[i] != want[i] {
				fmt.Printf("      first divergence at offset %d\n", i)
				break
			}
		}
	}

	tx := &prototxs.Tx{Unsigned: protoUnsigned}
	if err := tx.Sign([][]*secp256k1.PrivateKey{{key}}); err != nil {
		check(name+"/sign", false, err.Error())
		return
	}
	signed := tx.Bytes()
	check(name+"/prefix", bytes.HasPrefix(signed, got),
		fmt.Sprintf("signed=%d unsigned=%d", len(signed), len(got)))

	parsed, err := nodetxs.Parse(signed)
	if err != nil {
		check(name+"/node.Parse", false, err.Error())
		return
	}
	sameKind := fmt.Sprintf("%T", parsed.Unsigned) == fmt.Sprintf("%T", nodeUnsigned)
	check(name+"/node.Parse", sameKind && len(parsed.Creds) == 1,
		fmt.Sprintf("type=%T creds=%d id=%s", parsed.Unsigned, len(parsed.Creds), parsed.ID()))

	// proto must also read back its own bytes.
	back, err := prototxs.Parse(signed)
	if err != nil {
		check(name+"/proto.Parse", false, err.Error())
		return
	}
	rt, err := prototxs.Marshal(back.Unsigned)
	check(name+"/roundtrip", err == nil && bytes.Equal(rt, got),
		fmt.Sprintf("creds=%d", len(back.Creds)))
}

func TestWire(outer *testing.T) {
	t = outer

	key, err := secp256k1.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	// kind 3 — BaseTx
	{
		base := envelope()
		n, err := nodetxs.NewBaseTx(&base)
		if err != nil {
			t.Fatal(err)
		}
		p := &prototxs.BaseTx{BaseTx: envelope()}
		compare("BaseTx", p, n, key)
	}

	// kind 4 — ImportTx
	{
		base := envelope()
		imported := []*lux.TransferableInput{{
			UTXOID: lux.UTXOID{TxID: id(0x66), OutputIndex: 1},
			Asset:  lux.Asset{ID: id(0x22)},
			In: &secp256k1fx.TransferInput{
				Amt:   1_000,
				Input: secp256k1fx.Input{SigIndices: []uint32{0}},
			},
		}}
		n, err := nodetxs.NewImportTx(&base, id(0x77), imported)
		if err != nil {
			t.Fatal(err)
		}
		p := &prototxs.ImportTx{
			BaseTx:         prototxs.BaseTx{BaseTx: envelope()},
			SourceChain:    id(0x77),
			ImportedInputs: imported,
		}
		compare("ImportTx", p, n, key)
	}

	// kind 5 — ExportTx
	{
		base := envelope()
		exported := []*lux.TransferableOutput{{
			Asset: lux.Asset{ID: id(0x22)},
			Out: &secp256k1fx.TransferOutput{
				Amt:          2_000,
				OutputOwners: *owners(),
			},
		}}
		n, err := nodetxs.NewExportTx(&base, id(0x88), exported)
		if err != nil {
			t.Fatal(err)
		}
		p := &prototxs.ExportTx{
			BaseTx:           prototxs.BaseTx{BaseTx: envelope()},
			DestinationChain: id(0x88),
			ExportedOutputs:  exported,
		}
		compare("ExportTx", p, n, key)
	}

	// kind 6 — CreateNetworkTx, legacy-permissioned (restake parent, no own set)
	{
		base := envelope()
		var nodeOwner nodefx.Owner = owners()
		n, err := nodetxs.NewCreateNetworkTx(
			&base,
			ids.Empty,
			nodeOwner,
			nodesec.Mode{RestakeParent: true},
			nil,
			ids.Empty,
			nil,
		)
		if err != nil {
			t.Fatal(err)
		}
		p := &prototxs.CreateNetworkTx{
			BaseTx:   prototxs.BaseTx{BaseTx: envelope()},
			Owner:    owners(),
			Security: protosec.Mode{RestakeParent: true},
		}
		compare("CreateNetworkTx", p, n, key)
	}

	// kind 6 — CreateNetworkTx, sovereign with a genesis validator
	{
		base := envelope()
		var pop nodesigner.ProofOfPossession
		for i := range pop.PublicKey {
			pop.PublicKey[i] = byte(i)
		}
		for i := range pop.ProofOfPossession {
			pop.ProofOfPossession[i] = byte(255 - i)
		}
		nvdr := &nodetxs.NetworkValidator{
			NodeID:  nodeID(0x99).Bytes(),
			Weight:  100,
			Balance: 50,
			Signer:  pop,
			RemainingBalanceOwner: nodemsg.PChainOwner{
				Threshold: 1, Addresses: []ids.ShortID{addr(0xaa)},
			},
			DeactivationOwner: nodemsg.PChainOwner{
				Threshold: 1, Addresses: []ids.ShortID{addr(0xbb)},
			},
		}
		var nodeOwner nodefx.Owner = owners()
		n, err := nodetxs.NewCreateNetworkTx(
			&base,
			id(0x01),
			nodeOwner,
			nodesec.Mode{Admission: nodesec.Gated, Manager: nodesec.PChain},
			[]*nodetxs.NetworkValidator{nvdr},
			id(0x02),
			[]byte{0xde, 0xad},
		)
		if err != nil {
			t.Fatal(err)
		}
		var ppop protosigner.ProofOfPossession
		ppop.PublicKey = pop.PublicKey
		ppop.ProofOfPossession = pop.ProofOfPossession
		p := &prototxs.CreateNetworkTx{
			BaseTx:   prototxs.BaseTx{BaseTx: envelope()},
			Parent:   id(0x01),
			Owner:    owners(),
			Security: protosec.Mode{Admission: protosec.Gated, Manager: protosec.PChain},
			Validators: []*prototxs.NetworkValidator{{
				NodeID:  nodeID(0x99).Bytes(),
				Weight:  100,
				Balance: 50,
				Signer:  ppop,
				RemainingBalanceOwner: protomsg.PChainOwner{
					Threshold: 1, Addresses: []ids.ShortID{addr(0xaa)},
				},
				DeactivationOwner: protomsg.PChainOwner{
					Threshold: 1, Addresses: []ids.ShortID{addr(0xbb)},
				},
			}},
			ManagerChainID: id(0x02),
			ManagerAddress: []byte{0xde, 0xad},
		}
		compare("CreateNetworkTx/sovereign", p, n, key)
	}

	// kind 7 — CreateChainTx
	{
		base := envelope()
		auth := &secp256k1fx.Input{SigIndices: []uint32{0, 2}}
		fxIDs := []ids.ID{id(0x0a), id(0x0b)}
		genesis := []byte(`{"config":{}}`)
		n, err := nodetxs.NewCreateChainTx(&base, id(0xcc), "osage l2", id(0xdd), fxIDs, genesis, auth)
		if err != nil {
			t.Fatal(err)
		}
		p := &prototxs.CreateChainTx{
			BaseTx:            prototxs.BaseTx{BaseTx: envelope()},
			ValidateNetworkID: id(0xcc),
			BlockchainName:    "osage l2",
			VMID:              id(0xdd),
			FxIDs:             fxIDs,
			GenesisData:       genesis,
			ChainAuth:         auth,
		}
		compare("CreateChainTx", p, n, key)
	}

	// kind 14 — AddPermissionlessValidatorTx
	{
		base := envelope()
		var pop nodesigner.ProofOfPossession
		for i := range pop.PublicKey {
			pop.PublicKey[i] = byte(i + 1)
		}
		for i := range pop.ProofOfPossession {
			pop.ProofOfPossession[i] = byte(i + 2)
		}
		stake := []*lux.TransferableOutput{{
			Asset: lux.Asset{ID: id(0x22)},
			Out: &secp256k1fx.TransferOutput{
				Amt:          2_000_000,
				OutputOwners: *owners(),
			},
		}}
		nv := nodetxs.Validator{NodeID: nodeID(0xee), Start: 100, End: 200, Wght: 300}
		var valOwner, delOwner nodefx.Owner = owners(), owners()
		n, err := nodetxs.NewAddPermissionlessValidatorTx(
			&base, nv, id(0x00), &pop, stake, valOwner, delOwner, 20_000,
		)
		if err != nil {
			t.Fatal(err)
		}
		var ppop protosigner.ProofOfPossession
		ppop.PublicKey = pop.PublicKey
		ppop.ProofOfPossession = pop.ProofOfPossession
		p := &prototxs.AddPermissionlessValidatorTx{
			BaseTx:                prototxs.BaseTx{BaseTx: envelope()},
			Validator:             prototxs.Validator{NodeID: nodeID(0xee), Start: 100, End: 200, Wght: 300},
			Chain:                 id(0x00),
			Signer:                &ppop,
			StakeOuts:             stake,
			ValidatorRewardsOwner: owners(),
			DelegatorRewardsOwner: owners(),
			DelegationShares:      20_000,
		}
		compare("AddPermissionlessValidatorTx", p, n, key)
	}

	// kind 18 — IncreaseL1ValidatorBalanceTx
	{
		base := envelope()
		n, err := nodetxs.NewIncreaseL1ValidatorBalanceTx(&base, id(0x0f), 12_345)
		if err != nil {
			t.Fatal(err)
		}
		p := &prototxs.IncreaseL1ValidatorBalanceTx{
			BaseTx:       prototxs.BaseTx{BaseTx: envelope()},
			ValidationID: id(0x0f),
			Balance:      12_345,
		}
		compare("IncreaseL1ValidatorBalanceTx", p, n, key)
	}

	// Negative control: the probe must reject bytes that are not this wire.
	{
		_, err := nodetxs.Parse([]byte("not a zap buffer at all, definitely"))
		check("negative-control", err != nil, fmt.Sprintf("node rejects non-wire bytes: %v", err))
	}
}
