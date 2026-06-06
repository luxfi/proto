// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message

import (
	"errors"
)

const CodecVersion = 0

// Codec is the proto/p/warp/message-local wire codec interface. It is
// structurally identical to the legacy codec.Manager surface that lived
// in `github.com/luxfi/codec`, but defined locally so proto/p carries no
// import of that package. Any concrete codec (linearcodec via
// codec.Manager) satisfies this interface by shape, and callers from
// outside proto/p inject the implementation at parser construction time
// (see proto/internal/pcodectest).
//
// Wave 2A of the codec rip (#101). The shape of this interface MUST
// stay byte-compatible with codec.Manager so existing wire bytes
// continue to roundtrip during the multi-wave migration.
type Codec interface {
	Marshal(version uint16, source interface{}) ([]byte, error)
	Unmarshal(bytes []byte, dest interface{}) (uint16, error)
	Size(version uint16, value interface{}) (int, error)
}

// Registry is the proto/p/warp/message-local type-registration surface.
// Same rationale as Codec — structurally mirrors codec.Registry.
type Registry interface {
	RegisterType(interface{}) error
}

// RegisterTypes registers the warp-message wire payload types onto the
// supplied registry. Callers wire this through the registry they hand
// to their Codec implementation before invoking any Parse*.
func RegisterTypes(r Registry) error {
	return errors.Join(
		r.RegisterType(&ChainToL1Conversion{}),
		r.RegisterType(&RegisterL1Validator{}),
		r.RegisterType(&L1ValidatorRegistration{}),
		r.RegisterType(&L1ValidatorWeight{}),
	)
}
