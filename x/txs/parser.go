// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"fmt"
	"reflect"

	log "github.com/luxfi/log"
	"github.com/luxfi/proto/x/fxs"
	"github.com/luxfi/timer/mockable"
	"github.com/luxfi/utxo/nftfx"
	"github.com/luxfi/utxo/propertyfx"
	"github.com/luxfi/utxo/secp256k1fx"
)

// CodecVersion is the current default codec version
const CodecVersion = 0

var _ Parser = (*parser)(nil)

// Parser owns the wire codecs for both regular and genesis txs and the
// per-fx type registries used at fx initialization. It is constructed
// once at consumer boot (see luxfi/sdk/wallet/chain/x/builder for the
// canonical wiring) and reused for the life of the chain.
type Parser interface {
	Codec() Codec
	GenesisCodec() Codec

	CodecRegistry() Registry
	GenesisCodecRegistry() Registry

	ParseTx(bytes []byte) (*Tx, error)
	ParseGenesisTx(bytes []byte) (*Tx, error)
}

// ParserCodecs bundles the four codec-shaped dependencies the parser
// needs. The caller (outside proto/x) constructs concrete linearcodec
// or zapcodec instances and a codec.Manager wrapping each, then hands
// the bundle in. proto/x stays free of any github.com/luxfi/codec
// import.
//
//   - Codec / GenesisCodec are the wire codec managers (the legacy
//     codec.Manager). GenesisCodec typically allows a larger maximum
//     blob to accommodate the genesis tx, while Codec uses the default
//     limit for runtime txs.
//   - Registry / GenesisRegistry are the per-codec type registries
//     (legacy linearcodec.Codec / codec.Registry). They are exposed so
//     fx packages can register their own typed payloads at
//     initialization, and so the parser can also seed BaseTx /
//     CreateAssetTx / OperationTx / ImportTx / ExportTx.
type ParserCodecs struct {
	Codec            Codec
	GenesisCodec     Codec
	Registry         Registry
	GenesisRegistry  Registry
}

type parser struct {
	cm  Codec
	gcm Codec
	c   Registry
	gc  Registry
}

// NewParser registers the five XVM tx types onto both supplied
// registries, initializes the supplied fxs against a derived
// codec-registry adapter, and returns a Parser. The caller is
// responsible for ensuring the registries are wired into the
// corresponding codec managers before invocation — both registries
// must already accept RegisterType calls, and any subsequent Marshal
// / Unmarshal through Codec / GenesisCodec must dispatch through
// those same registries.
func NewParser(codecs ParserCodecs, fxList []fxs.Fx) (Parser, error) {
	return NewCustomParser(
		codecs,
		make(map[reflect.Type]int),
		&mockable.Clock{},
		log.Noop(),
		fxList,
	)
}

// NewCustomParser is NewParser with explicit clock and logger
// injection — used by VMs that want to share their consensus clock
// and structured logger with the fx machinery.
func NewCustomParser(
	codecs ParserCodecs,
	typeToFxIndex map[reflect.Type]int,
	clock *mockable.Clock,
	logger log.Logger,
	fxList []fxs.Fx,
) (Parser, error) {
	if codecs.Codec == nil || codecs.GenesisCodec == nil ||
		codecs.Registry == nil || codecs.GenesisRegistry == nil {
		return nil, fmt.Errorf("parser: all four ParserCodecs fields must be non-nil")
	}

	registries := []Registry{codecs.GenesisRegistry, codecs.Registry}
	for _, r := range registries {
		for _, typ := range []interface{}{
			&BaseTx{},
			&CreateAssetTx{},
			&OperationTx{},
			&ImportTx{},
			&ExportTx{},
		} {
			if err := r.RegisterType(typ); err != nil {
				return nil, fmt.Errorf("parser: register tx type: %w", err)
			}
		}
	}

	vm := &fxVM{
		typeToFxIndex: typeToFxIndex,
		clock:         clock,
		log:           logger,
	}
	for i, fx := range fxList {
		registry := &codecRegistry{
			codecs:      registries,
			index:       i,
			typeToIndex: vm.typeToFxIndex,
		}
		vm.codecRegistry = registry
		if err := fx.Initialize(vm); err != nil {
			return nil, err
		}
		// The fx packages went ZAP-native upstream and dropped their
		// codec self-registration. Until proto/x finishes its own ZAP
		// migration in a later wave, the parser has to register the
		// fx-owned wire types here so that linearcodec slot IDs stay
		// stable across the recognized fxs (secp256k1fx, nftfx,
		// propertyfx). Each call below routes through the per-fx
		// registry adapter, so the typeToFxIndex map is also populated
		// correctly for the executor's polymorphic dispatch.
		for _, typ := range fxOwnedTypes(fx) {
			if err := registry.RegisterType(typ); err != nil {
				return nil, fmt.Errorf("parser: register fx type: %w", err)
			}
		}
	}
	return &parser{
		cm:  codecs.Codec,
		gcm: codecs.GenesisCodec,
		c:   codecs.Registry,
		gc:  codecs.GenesisRegistry,
	}, nil
}

func (p *parser) Codec() Codec {
	return p.cm
}

func (p *parser) GenesisCodec() Codec {
	return p.gcm
}

func (p *parser) CodecRegistry() Registry {
	return p.c
}

func (p *parser) GenesisCodecRegistry() Registry {
	return p.gc
}

func (p *parser) ParseTx(bytes []byte) (*Tx, error) {
	return parse(p.cm, bytes)
}

func (p *parser) ParseGenesisTx(bytes []byte) (*Tx, error) {
	return parse(p.gcm, bytes)
}

// fxOwnedTypes returns the wire payload types historically registered
// by each known fx's Initialize. Order matches the legacy registration
// order so linearcodec slot IDs are stable: TransferInput before
// MintOutput before TransferOutput etc. — see luxfi/utxo/secp256k1fx/
// fx.go pre-138a575 for the canonical reference list.
//
// Unknown fxs return nil — they're expected to register their own
// types via their Initialize implementation (kept for forward
// compatibility with future fxs that may still need codec dispatch).
func fxOwnedTypes(fx fxs.Fx) []interface{} {
	switch fx.(type) {
	case *secp256k1fx.Fx:
		return []interface{}{
			&secp256k1fx.TransferInput{},
			&secp256k1fx.MintOutput{},
			&secp256k1fx.TransferOutput{},
			&secp256k1fx.MintOperation{},
			&secp256k1fx.Credential{},
		}
	case *nftfx.Fx:
		return []interface{}{
			&nftfx.MintOutput{},
			&nftfx.TransferOutput{},
			&nftfx.MintOperation{},
			&nftfx.TransferOperation{},
			&nftfx.Credential{},
		}
	case *propertyfx.Fx:
		return []interface{}{
			&propertyfx.MintOutput{},
			&propertyfx.OwnedOutput{},
			&propertyfx.MintOperation{},
			&propertyfx.BurnOperation{},
			&propertyfx.Credential{},
		}
	default:
		return nil
	}
}

func parse(cm Codec, signedBytes []byte) (*Tx, error) {
	tx := &Tx{}
	parsedVersion, err := cm.Unmarshal(signedBytes, tx)
	if err != nil {
		return nil, err
	}
	if parsedVersion != CodecVersion {
		return nil, fmt.Errorf("expected codec version %d but got %d", CodecVersion, parsedVersion)
	}

	unsignedBytesLen, err := cm.Size(CodecVersion, &tx.Unsigned)
	if err != nil {
		return nil, fmt.Errorf("couldn't calculate UnsignedTx marshal length: %w", err)
	}

	unsignedBytes := signedBytes[:unsignedBytesLen]
	tx.SetBytes(unsignedBytes, signedBytes)
	return tx, nil
}
