// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package pcodectest is the canonical test wiring for the proto/p PVM
// codec set. proto/p carries no github.com/luxfi/codec import after the
// Wave 2A rip (#101); this helper package is the bridge that lets test
// suites under proto/p construct linearcodec-backed codecs without
// duplicating wire-registration logic across every test file.
//
// Production callers (luxfi/node/vms/platformvm/...) construct their
// codecs inline. This helper exists strictly so the in-tree test files
// don't need to duplicate the wiring.
package pcodectest

import (
	"math"

	"github.com/luxfi/codec"
	"github.com/luxfi/codec/linearcodec"

	"github.com/luxfi/proto/p/warp"
	warpmsg "github.com/luxfi/proto/p/warp/message"
	"github.com/luxfi/proto/p/warp/payload"
)

// NewPayloadCodec returns a codec.Manager + linearcodec pair with the
// canonical warp payload types (Hash, AddressedCall) registered. Used by
// proto/p/warp/payload tests in lieu of the legacy package-level
// payload.Codec singleton.
func NewPayloadCodec() payload.Codec {
	c := linearcodec.NewDefault()
	cm := codec.NewManager(payload.MaxMessageSize)
	if err := payload.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(payload.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm
}

// NewMessageCodec returns a codec.Manager + linearcodec pair with the
// canonical warp/message types registered.
func NewMessageCodec() warpmsg.Codec {
	c := linearcodec.NewDefault()
	cm := codec.NewManager(math.MaxInt)
	if err := warpmsg.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(warpmsg.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm
}

// NewWarpCodec returns a codec.Manager + linearcodec pair with the
// canonical proto/p/warp signature + teleport types registered.
func NewWarpCodec() warp.Codec {
	c := linearcodec.NewDefault()
	cm := codec.NewManager(math.MaxInt)
	if err := warp.RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(warp.CodecVersion, c); err != nil {
		panic(err)
	}
	return cm
}
