// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"reflect"

	log "github.com/luxfi/log"
	"github.com/luxfi/timer/mockable"
	"github.com/luxfi/utxo/secp256k1fx"
)

var _ secp256k1fx.VM = (*fxVM)(nil)

// Codec is the proto/x-local wire codec interface. It is structurally
// identical to the legacy codec.Manager surface that lived in
// `github.com/luxfi/codec`, but defined locally so proto/x carries no
// import of that package. Any concrete codec (linearcodec via
// codec.Manager, or zapcodec) satisfies this interface by shape, and
// callers from outside proto/x inject the implementation at parser
// construction time (see NewParser).
//
// Wave 1A of the codec rip (#101). The shape of this interface MUST
// stay byte-compatible with codec.Manager so existing wire bytes
// continue to roundtrip during the multi-wave migration.
type Codec interface {
	Marshal(version uint16, source interface{}) ([]byte, error)
	Unmarshal(bytes []byte, dest interface{}) (uint16, error)
	Size(version uint16, value interface{}) (int, error)
}

// Registry is the proto/x-local type-registration surface. Same
// rationale as Codec — structurally mirrors codec.Registry.
type Registry interface {
	RegisterType(interface{}) error
}

// codecRegistry fans a single RegisterType call across a slice of
// Registry instances. Used by the fxVM to register fx-owned types
// across the codec and genesis codec simultaneously, while also
// recording the fx-index for the reflect type so VerifyOperation can
// route polymorphic operations to the correct fx.
type codecRegistry struct {
	codecs      []Registry
	index       int
	typeToIndex map[reflect.Type]int
}

func (cr *codecRegistry) RegisterType(val interface{}) error {
	valType := reflect.TypeOf(val)
	cr.typeToIndex[valType] = cr.index

	var firstErr error
	for _, c := range cr.codecs {
		if err := c.RegisterType(val); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type fxVM struct {
	typeToFxIndex map[reflect.Type]int

	clock         *mockable.Clock
	log           log.Logger
	codecRegistry Registry
}

func (vm *fxVM) Clock() *mockable.Clock {
	return vm.clock
}

func (vm *fxVM) Logger() log.Logger {
	return vm.log
}

// CodecRegistry exposes the fx-bound registry to downstream Fx machinery.
// The downstream Fx contract in luxfi/utxo currently doesn't require it
// on the VM interface (see secp256k1fx.VM — codec-free post-Wave-1A),
// but some Fx implementations cast their VM to a registry-bearing type
// when registering their own typed payloads. Returning the local
// Registry interface keeps the proto/x package free of any direct
// github.com/luxfi/codec import.
func (vm *fxVM) CodecRegistry() Registry {
	return vm.codecRegistry
}
