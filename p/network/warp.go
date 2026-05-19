// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"context"
	"errors"
	"sync"

	"github.com/luxfi/proto/p/state"
	"github.com/luxfi/warp"
)

// Warp validator-registration verification was originally backed by
// protobuf justification messages generated under //go:build grpc.
// luxfi/node/proto/pb/platformvm now ships only a stub package
// declaration there (the actual generated types are not in tree), so
// the historical verification path doesn't compile against the
// canonical default build. Until the ZAP-native equivalent lands,
// keep the struct shape + Verifier interface satisfied with a
// stub Verify that refuses unconditionally. Callers that need the
// protobuf path build with -tags grpc and patch in the legacy file.

var _ warp.Verifier = (*signatureRequestVerifier)(nil)

// ErrWarpVerificationNotImplemented is returned by the default Verify.
// Until the ZAP-native verifier lands, treat this as "no peer will
// satisfy your warp signature request from this node."
var ErrWarpVerificationNotImplemented = errors.New(
	"network: warp signature verification not implemented under default build (rebuild with -tags grpc + regen pb)",
)

type signatureRequestVerifier struct {
	stateLock sync.Locker
	state     state.Chain
}

// Verify is the stub fallback. Returns ErrWarpVerificationNotImplemented
// to make the refusal explicit rather than silent. The pb-backed real
// implementation lives in a separate file under -tags grpc.
func (s signatureRequestVerifier) Verify(
	_ context.Context,
	_ *warp.UnsignedMessage,
	_ []byte,
) error {
	return ErrWarpVerificationNotImplemented
}
