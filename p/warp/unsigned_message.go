// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package warp

import (
	"fmt"

	"github.com/luxfi/crypto/hash"
	"github.com/luxfi/ids"
)

// UnsignedMessage defines the standard format for an unsigned Warp message.
type UnsignedMessage struct {
	NetworkID     uint32 `serialize:"true"`
	SourceChainID ids.ID `serialize:"true"`
	Payload       []byte `serialize:"true"`

	bytes []byte
	id    ids.ID
}

// NewUnsignedMessage creates a new *UnsignedMessage and initializes it
// against the supplied Codec.
func NewUnsignedMessage(
	c Codec,
	networkID uint32,
	sourceChainID ids.ID,
	payload []byte,
) (*UnsignedMessage, error) {
	msg := &UnsignedMessage{
		NetworkID:     networkID,
		SourceChainID: sourceChainID,
		Payload:       payload,
	}
	return msg, msg.Initialize(c)
}

// ParseUnsignedMessage converts a slice of bytes into an initialized
// *UnsignedMessage using the supplied Codec.
func ParseUnsignedMessage(c Codec, b []byte) (*UnsignedMessage, error) {
	msg := &UnsignedMessage{
		bytes: b,
		id:    hash.ComputeHash256Array(b),
	}
	_, err := c.Unmarshal(b, msg)
	return msg, err
}

// Initialize recalculates the result of Bytes() using the supplied
// Codec.
func (m *UnsignedMessage) Initialize(c Codec) error {
	bytes, err := c.Marshal(CodecVersion, m)
	if err != nil {
		return fmt.Errorf("couldn't marshal warp unsigned message: %w", err)
	}
	m.bytes = bytes
	m.id = hash.ComputeHash256Array(m.bytes)
	return nil
}

// Bytes returns the binary representation of this message. It assumes that the
// message is initialized from either New, Parse, or an explicit call to
// Initialize.
func (m *UnsignedMessage) Bytes() []byte {
	return m.bytes
}

// ID returns an identifier for this message. It assumes that the
// message is initialized from either New, Parse, or an explicit call to
// Initialize.
func (m *UnsignedMessage) ID() ids.ID {
	return m.id
}

func (m *UnsignedMessage) String() string {
	return fmt.Sprintf("UnsignedMessage(NetworkID = %d, SourceChainID = %s, Payload = %x)", m.NetworkID, m.SourceChainID, m.Payload)
}
