// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package state

const (
	CodecVersion0Tag        = "v0"
	CodecVersion0    uint16 = 0

	CodecVersion1Tag        = "v1"
	CodecVersion1    uint16 = 1
)

// MetadataCodec is the wire codec interface used by metadata_validator
// and metadata_delegator to marshal validator/delegator state. It is
// structurally identical to codec.Manager — proto/p carries no
// github.com/luxfi/codec import after the Wave 2A rip (#101), so the
// concrete codec is constructed externally (production: PVM wiring;
// tests: proto/internal/pcodectest.NewMetadataCodec) and threaded into
// state.New as a constructor parameter.
//
// The metadata codec is a SEPARATE codec from the block/txs codecs —
// it has its own version-tagged registration sequence keyed off the
// v0:"true" / v1:"true" struct tags on validatorMetadata /
// delegatorMetadata. Do NOT pass the block.Codec here.
type MetadataCodec interface {
	Marshal(version uint16, source interface{}) ([]byte, error)
	Unmarshal(bytes []byte, dest interface{}) (uint16, error)
	Size(version uint16, value interface{}) (int, error)
}
