// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message

import (
	"errors"
	"fmt"
)

var ErrWrongType = errors.New("wrong payload type")

// Payload provides a common interface for all payloads implemented by this
// package.
type Payload interface {
	// Bytes returns the binary representation of this payload.
	//
	// If the payload is not initialized, this method will return nil.
	Bytes() []byte

	// initialize the payload with the provided binary representation.
	initialize(b []byte)
}

// payload is embedded by all the payloads to provide the common implementation
// of Payload.
type payload []byte

func (p payload) Bytes() []byte {
	return p
}

func (p *payload) initialize(bytes []byte) {
	*p = bytes
}

// Parse decodes a payload from wire bytes using the supplied Codec.
func Parse(c Codec, bytes []byte) (Payload, error) {
	var p Payload
	if _, err := c.Unmarshal(bytes, &p); err != nil {
		return nil, err
	}
	p.initialize(bytes)
	return p, nil
}

// Initialize marshals p through the supplied Codec and stores the wire
// representation on p via its private initialize() method.
func Initialize(c Codec, p Payload) error {
	bytes, err := c.Marshal(CodecVersion, &p)
	if err != nil {
		return fmt.Errorf("couldn't marshal %T payload: %w", p, err)
	}
	p.initialize(bytes)
	return nil
}
