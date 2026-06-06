// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package warp

import "fmt"

// Message defines the standard format for a Warp message.
type Message struct {
	UnsignedMessage `serialize:"true"`
	Signature       Signature `serialize:"true"`

	bytes []byte
}

// NewMessage creates a new *Message and initializes it against the
// supplied Codec.
func NewMessage(
	c Codec,
	unsignedMsg *UnsignedMessage,
	signature Signature,
) (*Message, error) {
	msg := &Message{
		UnsignedMessage: *unsignedMsg,
		Signature:       signature,
	}
	return msg, msg.Initialize(c)
}

// ParseMessage converts a slice of bytes into an initialized *Message
// using the supplied Codec.
func ParseMessage(c Codec, b []byte) (*Message, error) {
	msg := &Message{
		bytes: b,
	}
	_, err := c.Unmarshal(b, msg)
	if err != nil {
		return nil, err
	}
	return msg, msg.UnsignedMessage.Initialize(c)
}

// Initialize recalculates the result of Bytes() using the supplied
// Codec. It does not call Initialize() on the UnsignedMessage.
func (m *Message) Initialize(c Codec) error {
	bytes, err := c.Marshal(CodecVersion, m)
	m.bytes = bytes
	return err
}

// Bytes returns the binary representation of this message. It assumes that the
// message is initialized from either New, Parse, or an explicit call to
// Initialize.
func (m *Message) Bytes() []byte {
	return m.bytes
}

func (m *Message) String() string {
	return fmt.Sprintf("WarpMessage(%s, %s)", &m.UnsignedMessage, m.Signature)
}
