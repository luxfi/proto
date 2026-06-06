// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"github.com/luxfi/proto/internal/pcodectest"
)

// testCodec is the package-internal test codec. The legacy package
// surface exported a `Codec` codec.Manager singleton; after the
// Wave 2A rip (#101) `Codec` is the *interface type* and there is no
// package-level codec value. Tests inside this package use testCodec
// when they need a working codec.Manager-shaped instance for
// Marshal/Unmarshal/Initialize/NewSigned/Parse.
//
// Constructed once at package init using proto/internal/pcodectest
// (the codec-agnostic helper, which does not import proto/p/txs and
// thus does not close a test-time import cycle). The PVM-aware
// pvmcodectest helper IS NOT used here for the same reason — it
// imports proto/p/{block,txs,...}, which would close the cycle.
var testCodec Codec = func() Codec {
	c := pcodectest.NewLinearCodec()
	cm := pcodectest.NewDefaultCodecManager()
	if err := RegisterTypes(c); err != nil {
		panic(err)
	}
	if err := cm.RegisterCodec(CodecVersion, c); err != nil {
		panic(err)
	}
	return cm
}()
