// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package payload

import (
	"errors"
	"fmt"
)

var ErrWrongType = errors.New("wrong payload type")

// Payload provides a common interface for all payloads implemented by this
// package.
type Payload interface {
	// Bytes returns the binary representation of this payload.
	Bytes() []byte

	// initialize the payload with the provided binary representation.
	initialize(b []byte)
}

// Parse decodes a payload from wire bytes using the supplied Codec. The
// codec must have RegisterTypes (or its equivalent) called against its
// registry before invocation.
func Parse(c Codec, bytes []byte) (Payload, error) {
	var payload Payload
	if _, err := c.Unmarshal(bytes, &payload); err != nil {
		return nil, err
	}
	payload.initialize(bytes)
	return payload, nil
}

// initialize marshals p through the supplied Codec and stores the wire
// representation on p via its private initialize() method.
func initialize(c Codec, p Payload) error {
	bytes, err := c.Marshal(CodecVersion, &p)
	if err != nil {
		return fmt.Errorf("couldn't marshal %T payload: %w", p, err)
	}
	p.initialize(bytes)
	return nil
}
