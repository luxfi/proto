// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import "github.com/luxfi/proto/p/block"

const CodecVersion = block.CodecVersion

// Codec is an alias for block.Codec — same wire codec interface,
// re-exported under the genesis package name for callers that want to
// reference the genesis-bound codec by its semantic name. The block
// genesis codec (block.RegisterTypes bound through a
// codec.NewManager(math.MaxInt32)) is the canonical concrete instance.
type Codec = block.Codec
