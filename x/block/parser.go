// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block

import (
	"errors"
	"reflect"

	log "github.com/luxfi/log"
	"github.com/luxfi/proto/x/fxs"
	"github.com/luxfi/proto/x/txs"
	"github.com/luxfi/timer/mockable"
)

// CodecVersion is the current default codec version
const CodecVersion = txs.CodecVersion

var _ Parser = (*parser)(nil)

type Parser interface {
	txs.Parser

	ParseBlock(bytes []byte) (Block, error)
	ParseGenesisBlock(bytes []byte) (Block, error)
}

type parser struct {
	txs.Parser
}

// NewParser wires the block-level type registry on top of a tx-level
// parser. The caller injects the four ParserCodecs (codec.Manager /
// linearcodec or zapcodec instances). proto/x/block stays free of
// any github.com/luxfi/codec import — Wave 1A of the codec rip (#101).
func NewParser(codecs txs.ParserCodecs, fxList []fxs.Fx) (Parser, error) {
	p, err := txs.NewParser(codecs, fxList)
	if err != nil {
		return nil, err
	}
	c := p.CodecRegistry()
	gc := p.GenesisCodecRegistry()

	err = errors.Join(
		c.RegisterType(&StandardBlock{}),
		gc.RegisterType(&StandardBlock{}),
	)
	return &parser{
		Parser: p,
	}, err
}

// NewCustomParser is NewParser with explicit clock and logger
// injection.
func NewCustomParser(
	codecs txs.ParserCodecs,
	typeToFxIndex map[reflect.Type]int,
	clock *mockable.Clock,
	logger log.Logger,
	fxList []fxs.Fx,
) (Parser, error) {
	p, err := txs.NewCustomParser(codecs, typeToFxIndex, clock, logger, fxList)
	if err != nil {
		return nil, err
	}
	c := p.CodecRegistry()
	gc := p.GenesisCodecRegistry()

	err = errors.Join(
		c.RegisterType(&StandardBlock{}),
		gc.RegisterType(&StandardBlock{}),
	)
	return &parser{
		Parser: p,
	}, err
}

func (p *parser) ParseBlock(bytes []byte) (Block, error) {
	return parse(p.Codec(), bytes)
}

func (p *parser) ParseGenesisBlock(bytes []byte) (Block, error) {
	return parse(p.GenesisCodec(), bytes)
}

func parse(cm txs.Codec, bytes []byte) (Block, error) {
	var blk Block
	if _, err := cm.Unmarshal(bytes, &blk); err != nil {
		return nil, err
	}
	return blk, blk.initialize(bytes, cm)
}
