// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

import (
	"github.com/luxfi/proto/p/block"
)

const (
	CodecVersion0Tag        = "v0"
	CodecVersion0    uint16 = 0

	CodecVersion1Tag        = "v1"
	CodecVersion1    uint16 = 1
)

// MetadataCodec is the codec interface used by metadata_validator and
// metadata_delegator to marshal validator/delegator state. It mirrors
// block.Codec — the wire codec.Manager surface — and is set on the
// state struct at construction so that callers can inject the
// versioned-tag codec (v0 + v1) that the legacy state layout requires.
//
// Wave 2A of the codec rip (#101). After the rip, proto/p no longer
// holds a package-level MetadataCodec singleton; the field on *state
// is supplied by the constructor.
type MetadataCodec = block.Codec
