// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p2p

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// The handshake wire format is positional: fields are identified by their order,
// not by a tag, so the only safe way to extend it is to append at the tail and
// guard the read with HasMore. These tests hold that property for Chains, because
// getting it wrong does not produce a decode error — it produces two nodes that
// read each other's fields off by one and disagree about what was said.

func handshake(chains []*ChainIdentity) *Handshake {
	return &Handshake{
		NetworkId:  96369,
		MyTime:     1735084800,
		IpAddr:     []byte{127, 0, 0, 1},
		IpPort:     9651,
		Client:     &Client{Name: "luxd", Major: 1, Minor: 36, Patch: 58},
		IpBlsSig:   bytes.Repeat([]byte{0xbb}, 96),
		IpMldsaSig: bytes.Repeat([]byte{0xdd}, 3309),
		Chains:     chains,
	}
}

func roundTrip(t *testing.T, h *Handshake) *Handshake {
	t.Helper()
	raw, err := Marshal(&Message{Message: &Message_Handshake{Handshake: h}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back Message
	if err := Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := back.GetHandshake()
	if got == nil {
		t.Fatal("decoded message is not a handshake")
	}
	return got
}

func TestHandshakeChainsRoundTrip(t *testing.T) {
	sent := []*ChainIdentity{
		{
			NetworkId:     96369,
			ChainId:       bytes.Repeat([]byte{0x11}, 32),
			VmId:          bytes.Repeat([]byte{0x22}, 32),
			GenesisDigest: bytes.Repeat([]byte{0x33}, 32),
			RulesId:       bytes.Repeat([]byte{0x44}, 32),
		},
		{
			NetworkId:     96369,
			ChainId:       bytes.Repeat([]byte{0x55}, 32),
			VmId:          bytes.Repeat([]byte{0x66}, 32),
			GenesisDigest: bytes.Repeat([]byte{0x77}, 32),
			// A chain that declares no rule generation sends an empty RulesId,
			// which must survive as empty rather than as a 32-byte zero.
		},
	}

	got := roundTrip(t, handshake(sent))
	if len(got.Chains) != len(sent) {
		t.Fatalf("decoded %d chains, sent %d", len(got.Chains), len(sent))
	}
	for i, want := range sent {
		have := got.Chains[i]
		if have.NetworkId != want.NetworkId ||
			!bytes.Equal(have.ChainId, want.ChainId) ||
			!bytes.Equal(have.VmId, want.VmId) ||
			!bytes.Equal(have.GenesisDigest, want.GenesisDigest) ||
			!bytes.Equal(have.RulesId, want.RulesId) {
			t.Errorf("chain %d round-tripped as %+v, sent %+v", i, have, want)
		}
	}

	// Fields ahead of Chains must be untouched: a positional encoder that
	// mis-sizes the tail corrupts everything after it, and a decoder that reads
	// the tail early corrupts everything before.
	if !bytes.Equal(got.IpMldsaSig, handshake(nil).IpMldsaSig) {
		t.Error("IpMldsaSig did not survive the appended field")
	}
	if got.Client.GetPatch() != 58 || got.NetworkId != 96369 {
		t.Error("leading fields moved")
	}
}

// TestHandshakeChainsEmptyIsNotAbsent is the distinction the fail-closed rule
// rests on. A peer that runs no chains says so with a zero count; a peer too old
// to have the field says nothing and its frame ends early. Both decode to an
// empty slice, so the two are told apart by frame length, and that must hold on
// the wire even though the decoded values look alike.
func TestHandshakeChainsEmptyIsNotAbsent(t *testing.T) {
	stated, err := Marshal(&Message{Message: &Message_Handshake{Handshake: handshake([]*ChainIdentity{})}})
	if err != nil {
		t.Fatal(err)
	}
	// Chains is the append-only tail: its 4-byte count is written last, so a
	// peer too old to say it ends the frame before it. AllChains is not part
	// of that tail — it sits mid-frame at a declared offset and every peer
	// writes it — so truncating before it is not a vintage the wire promises.
	countAt := len(stated) - 4
	if !bytes.Equal(stated[countAt:], []byte{0, 0, 0, 0}) {
		t.Error("an empty chain list is not encoded as a zero count")
	}

	for name, legacy := range map[string][]byte{
		"before Chains": stated[:countAt],
	} {
		var back Message
		if err := Unmarshal(legacy, &back); err != nil {
			t.Fatalf("a frame ending %s must still decode: %v", name, err)
		}
		if n := len(back.GetHandshake().Chains); n != 0 {
			t.Errorf("frame ending %s decoded %d chains, want 0", name, n)
		}
		if back.GetHandshake().AllChains {
			t.Errorf("frame ending %s decoded AllChains as set", name)
		}
		if !bytes.Equal(back.GetHandshake().IpMldsaSig, handshake(nil).IpMldsaSig) {
			t.Errorf("frame ending %s lost a field it did carry", name)
		}
	}
}

// TestHandshakeChainsLyingCountIsBounded: the count is attacker controlled and is
// read before any entry, so a decoder that preallocates from it can be made to
// allocate gigabytes from a frame of a few dozen bytes.
func TestHandshakeChainsLyingCountIsBounded(t *testing.T) {
	raw, err := Marshal(&Message{Message: &Message_Handshake{Handshake: handshake(nil)}})
	if err != nil {
		t.Fatal(err)
	}
	// The count sits ahead of the 1-byte AllChains flag that follows it.
	binary.BigEndian.PutUint32(raw[len(raw)-5:], 0xFFFFFFFF)

	var back Message
	if err := Unmarshal(raw, &back); err == nil {
		t.Fatal("a frame claiming 4294967295 chains decoded without error")
	}
}
