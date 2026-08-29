// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p2p

import (
	_ "embed"
	"fmt"
	"reflect"
	"testing"

	"github.com/luxfi/proto/node/zap/schema"
)

//go:embed p2p.zap
var schemaSrc []byte

// The wire is stated twice: as Go in codec.go and as a schema in p2p.zap.
// This file is what keeps the two honest. For every message it fills each
// field with a distinct value, encodes with the shipped codec, and reads the
// bytes back using nothing but the schema. A field the codec forgets, a
// width it gets wrong, a field the schema does not declare, and a field
// order the two disagree on all surface here as a failure.

// reader walks a frame under the stream framing that p2p.zap describes.
type reader struct {
	r *Reader
	f *schema.File
}

// value reads one field and returns it as the Go value it should equal.
func (w *reader) value(f schema.Field) (any, error) {
	switch f.Kind {
	case "bool":
		v, err := w.r.ReadUint8()
		return v == 1, err
	case "u8":
		return w.r.ReadUint8()
	case "u32":
		return w.r.ReadUint32()
	case "i32":
		return w.r.ReadInt32()
	case "u64":
		return w.r.ReadUint64()
	case "text":
		return w.r.ReadString()
	case "bytes":
		return w.r.ReadBytes()

	case "bytes_fixed":
		// The stream framing length-prefixes even a fixed field. The
		// prefix is redundant, so it is also a check: an id declared
		// fixed that arrives some other width is a schema that lied.
		b, err := w.r.ReadBytes()
		if err == nil && len(b) != f.Fixed {
			err = fmt.Errorf("declared bytes_fixed[%d], wire carried %d bytes", f.Fixed, len(b))
		}
		return b, err

	case "list":
		n, err := w.r.ReadUint32()
		if err != nil {
			return nil, err
		}
		out := make([]any, n)
		elem := schema.Field{Kind: f.Elem, Fixed: f.Fixed}
		for i := range out {
			if out[i], err = w.value(elem); err != nil {
				return nil, err
			}
		}
		return out, nil
	}
	// Anything else names a struct: inline here, since only a struct-typed
	// FIELD carries a presence byte (see optional).
	s, ok := w.f.Struct(string(f.Kind))
	if !ok {
		return nil, fmt.Errorf("schema declares no type %q", f.Kind)
	}
	return w.fields(s)
}

// fields reads every field of s in declaration order.
func (w *reader) fields(s *schema.Struct) ([]any, error) {
	out := make([]any, len(s.Fields))
	for i, f := range s.Fields {
		v, err := w.value(f)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", s.Name, f.Name, err)
		}
		out[i] = v
	}
	return out, nil
}

// optional reads a struct field written as a presence byte then its fields.
func (w *reader) optional(s *schema.Struct) ([]any, error) {
	present, err := w.r.ReadUint8()
	if err != nil || present == 0 {
		return nil, err
	}
	return w.fields(s)
}

// requireSameShape asserts the Go type and the schema declare the same
// fields in the same order. A field added to one and not the other fails
// here rather than becoming a silent hole on the wire.
func requireSameShape(t *testing.T, s *schema.Struct, typ reflect.Type) {
	t.Helper()
	var names []string
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).PkgPath == "" {
			names = append(names, typ.Field(i).Name)
		}
	}
	if len(names) != len(s.Fields) {
		t.Fatalf("%s: Go declares %v, schema declares %d fields", s.Name, names, len(s.Fields))
	}
	for i, f := range s.Fields {
		if names[i] != f.Name {
			t.Fatalf("%s field %d: Go says %q, schema says %q", s.Name, i, names[i], f.Name)
		}
	}
}

// requireSameValue compares what the schema read against what was sent.
func requireSameValue(t *testing.T, path string, sent reflect.Value, got any, f schema.Field, file *schema.File) {
	t.Helper()
	for sent.Kind() == reflect.Ptr {
		if sent.IsNil() {
			return // an absent optional struct reads back as nothing
		}
		sent = sent.Elem()
	}
	switch sent.Kind() {
	case reflect.Bool:
		if sent.Bool() != got.(bool) {
			t.Errorf("%s: sent %v, wire carried %v", path, sent.Bool(), got)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if sent.Uint() != reflect.ValueOf(got).Convert(sent.Type()).Uint() {
			t.Errorf("%s: sent %v, wire carried %v", path, sent.Uint(), got)
		}
	case reflect.Int32, reflect.Int64:
		if sent.Int() != reflect.ValueOf(got).Convert(sent.Type()).Int() {
			t.Errorf("%s: sent %v, wire carried %v", path, sent.Int(), got)
		}
	case reflect.String:
		if sent.String() != got.(string) {
			t.Errorf("%s: sent %q, wire carried %q", path, sent.String(), got)
		}
	case reflect.Slice:
		if sent.Type().Elem().Kind() == reflect.Uint8 {
			if string(sent.Bytes()) != string(got.([]byte)) {
				t.Errorf("%s: sent %x, wire carried %x", path, sent.Bytes(), got)
			}
			return
		}
		list, ok := got.([]any)
		if !ok || sent.Len() != len(list) {
			t.Errorf("%s: sent %d elements, wire carried %v", path, sent.Len(), got)
			return
		}
		elem := schema.Field{Kind: f.Elem, Fixed: f.Fixed}
		for i := range list {
			requireSameValue(t, fmt.Sprintf("%s[%d]", path, i), sent.Index(i), list[i], elem, file)
		}
	case reflect.Struct:
		s, ok := file.Struct(string(f.Kind))
		if !ok {
			s, _ = file.Struct(sent.Type().Name())
		}
		vals, ok := got.([]any)
		if !ok {
			t.Errorf("%s: sent a struct, wire carried %v", path, got)
			return
		}
		for i, sf := range s.Fields {
			requireSameValue(t, path+"."+sf.Name, sent.Field(i), vals[i], sf, file)
		}
	}
}

// fill puts a distinctive non-zero value in every field of v, recursively.
func fill(v reflect.Value) {
	switch v.Kind() {
	case reflect.Uint32, reflect.Uint64, reflect.Uint16, reflect.Uint8:
		v.SetUint(7)
	case reflect.Int32, reflect.Int64:
		v.SetInt(7)
	case reflect.Bool:
		v.SetBool(true)
	case reflect.String:
		v.SetString("x")
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// 32 bytes: every id on this wire is an ids.ID, and the
			// schema declares those bytes_fixed[32].
			b := make([]byte, 32)
			for i := range b {
				b[i] = byte(i + 1)
			}
			v.SetBytes(b)
			return
		}
		e := reflect.New(v.Type().Elem()).Elem()
		fill(e)
		v.Set(reflect.Append(v, e))
	case reflect.Ptr:
		v.Set(reflect.New(v.Type().Elem()))
		fill(v.Elem())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).PkgPath != "" {
				continue // unexported (the union interface)
			}
			fill(v.Field(i))
		}
	}
}

type messageCase struct {
	name string
	tag  int
	wrap func(reflect.Value) isMessage_Message
	typ  reflect.Type
}

func messageCases() []messageCase {
	return []messageCase{
		{"Ping", 2, func(v reflect.Value) isMessage_Message { return &Message_Ping{Ping: v.Interface().(*Ping)} }, reflect.TypeOf(&Ping{})},
		{"Pong", 3, func(v reflect.Value) isMessage_Message { return &Message_Pong{Pong: v.Interface().(*Pong)} }, reflect.TypeOf(&Pong{})},
		{"Handshake", 4, func(v reflect.Value) isMessage_Message {
			return &Message_Handshake{Handshake: v.Interface().(*Handshake)}
		}, reflect.TypeOf(&Handshake{})},
		{"GetPeerList", 5, func(v reflect.Value) isMessage_Message {
			return &Message_GetPeerList{GetPeerList: v.Interface().(*GetPeerList)}
		}, reflect.TypeOf(&GetPeerList{})},
		{"PeerList", 6, func(v reflect.Value) isMessage_Message {
			return &Message_PeerList_{PeerList_: v.Interface().(*PeerList)}
		}, reflect.TypeOf(&PeerList{})},
		{"GetStateSummaryFrontier", 7, func(v reflect.Value) isMessage_Message {
			return &Message_GetStateSummaryFrontier{GetStateSummaryFrontier: v.Interface().(*GetStateSummaryFrontier)}
		}, reflect.TypeOf(&GetStateSummaryFrontier{})},
		{"StateSummaryFrontier", 8, func(v reflect.Value) isMessage_Message {
			return &Message_StateSummaryFrontier_{StateSummaryFrontier_: v.Interface().(*StateSummaryFrontier)}
		}, reflect.TypeOf(&StateSummaryFrontier{})},
		{"GetAcceptedStateSummary", 9, func(v reflect.Value) isMessage_Message {
			return &Message_GetAcceptedStateSummary{GetAcceptedStateSummary: v.Interface().(*GetAcceptedStateSummary)}
		}, reflect.TypeOf(&GetAcceptedStateSummary{})},
		{"AcceptedStateSummary", 10, func(v reflect.Value) isMessage_Message {
			return &Message_AcceptedStateSummary_{AcceptedStateSummary_: v.Interface().(*AcceptedStateSummary)}
		}, reflect.TypeOf(&AcceptedStateSummary{})},
		{"GetAcceptedFrontier", 11, func(v reflect.Value) isMessage_Message {
			return &Message_GetAcceptedFrontier{GetAcceptedFrontier: v.Interface().(*GetAcceptedFrontier)}
		}, reflect.TypeOf(&GetAcceptedFrontier{})},
		{"AcceptedFrontier", 12, func(v reflect.Value) isMessage_Message {
			return &Message_AcceptedFrontier_{AcceptedFrontier_: v.Interface().(*AcceptedFrontier)}
		}, reflect.TypeOf(&AcceptedFrontier{})},
		{"GetAccepted", 13, func(v reflect.Value) isMessage_Message {
			return &Message_GetAccepted{GetAccepted: v.Interface().(*GetAccepted)}
		}, reflect.TypeOf(&GetAccepted{})},
		{"Accepted", 14, func(v reflect.Value) isMessage_Message {
			return &Message_Accepted_{Accepted_: v.Interface().(*Accepted)}
		}, reflect.TypeOf(&Accepted{})},
		{"GetAncestors", 15, func(v reflect.Value) isMessage_Message {
			return &Message_GetAncestors{GetAncestors: v.Interface().(*GetAncestors)}
		}, reflect.TypeOf(&GetAncestors{})},
		{"Ancestors", 16, func(v reflect.Value) isMessage_Message {
			return &Message_Ancestors_{Ancestors_: v.Interface().(*Ancestors)}
		}, reflect.TypeOf(&Ancestors{})},
		{"Get", 17, func(v reflect.Value) isMessage_Message { return &Message_Get{Get: v.Interface().(*Get)} }, reflect.TypeOf(&Get{})},
		{"Put", 18, func(v reflect.Value) isMessage_Message { return &Message_Put{Put: v.Interface().(*Put)} }, reflect.TypeOf(&Put{})},
		{"PushQuery", 19, func(v reflect.Value) isMessage_Message {
			return &Message_PushQuery{PushQuery: v.Interface().(*PushQuery)}
		}, reflect.TypeOf(&PushQuery{})},
		{"PullQuery", 20, func(v reflect.Value) isMessage_Message {
			return &Message_PullQuery{PullQuery: v.Interface().(*PullQuery)}
		}, reflect.TypeOf(&PullQuery{})},
		{"Chits", 21, func(v reflect.Value) isMessage_Message { return &Message_Chits{Chits: v.Interface().(*Chits)} }, reflect.TypeOf(&Chits{})},
		{"Request", 22, func(v reflect.Value) isMessage_Message { return &Message_Request{Request: v.Interface().(*Request)} }, reflect.TypeOf(&Request{})},
		{"Response", 23, func(v reflect.Value) isMessage_Message { return &Message_Response{Response: v.Interface().(*Response)} }, reflect.TypeOf(&Response{})},
		{"Gossip", 24, func(v reflect.Value) isMessage_Message { return &Message_Gossip{Gossip: v.Interface().(*Gossip)} }, reflect.TypeOf(&Gossip{})},
		{"Error", 25, func(v reflect.Value) isMessage_Message { return &Message_Error{Error: v.Interface().(*Error)} }, reflect.TypeOf(&Error{})},
	}
}

func TestSchemaDescribesTheWire(t *testing.T) {
	file, err := schema.Parse(schemaSrc)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range messageCases() {
		t.Run(c.name, func(t *testing.T) {
			s, ok := file.Struct(c.name)
			if !ok {
				t.Fatalf("p2p.zap declares no struct %s", c.name)
			}
			sent := reflect.New(c.typ.Elem())
			fill(sent.Elem())
			requireSameShape(t, s, c.typ.Elem())

			b, err := Marshal(&Message{Message: c.wrap(sent)})
			if err != nil {
				t.Fatal(err)
			}

			w := &reader{r: NewReader(b), f: file}
			tag, err := w.r.ReadUint8()
			if err != nil {
				t.Fatal(err)
			}
			if int(tag) != c.tag {
				t.Fatalf("discriminator: schema says %d, codec wrote %d", c.tag, tag)
			}

			for i, f := range s.Fields {
				var got any
				var err error
				if _, nested := file.Struct(string(f.Kind)); nested {
					got, err = w.optional(mustStruct(t, file, string(f.Kind)))
				} else {
					got, err = w.value(f)
				}
				if err != nil {
					t.Fatalf("%s.%s: %v", c.name, f.Name, err)
				}
				requireSameValue(t, c.name+"."+f.Name, sent.Elem().Field(i), got, f, file)
			}
			if w.r.HasMore() {
				t.Errorf("%s: codec wrote %d bytes the schema does not declare", c.name, len(b)-w.r.offset)
			}

			// The schema reading the frame is half the claim; the codec
			// reading its own frame back is the other half. An encoder
			// that writes a field its decoder skips passes the walk and
			// fails here.
			var back Message
			if err := Unmarshal(b, &back); err != nil {
				t.Fatalf("%s: codec cannot read its own frame: %v", c.name, err)
			}
			got := reflect.ValueOf(back.Message).Elem().Field(0)
			if !reflect.DeepEqual(sent.Interface(), got.Interface()) {
				t.Errorf("%s: sent %+v, decoded %+v", c.name, sent.Elem(), got.Elem())
			}
		})
	}
}

func mustStruct(t *testing.T, f *schema.File, name string) *schema.Struct {
	t.Helper()
	s, ok := f.Struct(name)
	if !ok {
		t.Fatalf("p2p.zap declares no struct %s", name)
	}
	return s
}

// TestBFTVariantsCrossTheWire pins the second discriminator. The union has
// no IDL form, so p2p.zap states the byte table in prose and this is what
// holds codec.go to it: every variant the Go type declares must survive.
func TestBFTVariantsCrossTheWire(t *testing.T) {
	for _, c := range []struct {
		name string
		msg  isBFT_Message
	}{
		{"BlockProposal", &BFT_BlockProposal{BlockProposal: &BlockProposal{Block: []byte("b")}}},
		{"Vote", &BFT_Vote{Vote: &Vote{BlockHash: []byte("h"), Signature: []byte("s")}}},
		{"EmptyVote", &BFT_EmptyVote{EmptyVote: &EmptyVote{View: 1, Seq: 2, Signature: []byte("s")}}},
		{"FinalizeVote", &BFT_FinalizeVote{FinalizeVote: &Vote{BlockHash: []byte("h"), Signature: []byte("s")}}},
		{"Notarization", &BFT_Notarization{Notarization: &QuorumCertificate{BlockHash: []byte("h"), View: 1, Seq: 2, AggregatedSignature: []byte("a"), Signers: []byte("x")}}},
		{"EmptyNotarization", &BFT_EmptyNotarization{EmptyNotarization: &EmptyNotarization{View: 1, Seq: 2, AggregatedSignature: []byte("a"), Signers: []byte("x")}}},
		{"Finalization", &BFT_Finalization{Finalization: &QuorumCertificate{BlockHash: []byte("h"), View: 1, Seq: 2, AggregatedSignature: []byte("a"), Signers: []byte("x")}}},
		{"ReplicationRequest", &BFT_ReplicationRequest{ReplicationRequest: &ReplicationRequest{Seqs: []uint64{1}, LatestRound: 3}}},
		{"ReplicationResponse", &BFT_ReplicationResponse{ReplicationResponse: &ReplicationResponse{Messages: [][]byte{[]byte("m")}}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			sent := &BFT{ChainId: make([]byte, 32), Message: c.msg}
			b, err := Marshal(&Message{Message: &Message_BFT{BFT: sent}})
			if err != nil {
				t.Fatal(err)
			}
			var out Message
			if err := Unmarshal(b, &out); err != nil {
				t.Fatal(err)
			}
			got := out.GetBFT()
			if got == nil || got.Message == nil {
				t.Fatalf("variant did not cross: %d bytes encoded, nothing decoded, no error", len(b))
			}
			if !reflect.DeepEqual(sent.Message, got.Message) {
				t.Errorf("sent %+v, received %+v", sent.Message, got.Message)
			}
		})
	}
}

// TestUnsetMessageIsRefused: a Message whose union is empty encodes to
// nothing, and nothing is not a frame.
func TestUnsetMessageIsRefused(t *testing.T) {
	if b, err := Marshal(&Message{}); err == nil {
		t.Errorf("marshal of an empty message returned %d bytes and no error", len(b))
	}
}
