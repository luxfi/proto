// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package warp_test

import (
	"testing"

	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
	"github.com/luxfi/ids"
	"github.com/luxfi/proto/internal/pvmcodectest"
	"github.com/luxfi/proto/p/warp"
	"github.com/stretchr/testify/require"
)

// TestNewTeleportMessage tests creating a new teleport message
func TestNewTeleportMessage(t *testing.T) {
	require := require.New(t)

	sourceChain := ids.GenerateTestID()
	destChain := ids.GenerateTestID()
	payload := []byte("test transfer payload")
	nonce := uint64(12345)

	msg := warp.NewTeleportMessage(warp.TeleportTransfer, sourceChain, destChain, nonce, payload)

	require.Equal(warp.TeleportVersion, msg.Version)
	require.Equal(warp.TeleportTransfer, msg.MessageType)
	require.Equal(sourceChain, msg.SourceChainID)
	require.Equal(destChain, msg.DestChainID)
	require.Equal(nonce, msg.Nonce)
	require.Equal(payload, msg.Payload)
	require.False(msg.Encrypted)
}

// TestTeleportMessageValidate tests message validation
func TestTeleportMessageValidate(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name        string
		msg         *warp.TeleportMessage
		expectError error
	}{
		{
			name: "valid message",
			msg: &warp.TeleportMessage{
				Version:       warp.TeleportVersion,
				MessageType:   warp.TeleportTransfer,
				SourceChainID: ids.GenerateTestID(),
				DestChainID:   ids.GenerateTestID(),
				Nonce:         1,
				Payload:       []byte("payload"),
			},
			expectError: nil,
		},
		{
			name: "invalid version",
			msg: &warp.TeleportMessage{
				Version:       99, // Wrong version
				MessageType:   warp.TeleportTransfer,
				SourceChainID: ids.GenerateTestID(),
				DestChainID:   ids.GenerateTestID(),
				Nonce:         1,
				Payload:       []byte("payload"),
			},
			expectError: warp.ErrInvalidTeleportVersion,
		},
		{
			name: "invalid message type",
			msg: &warp.TeleportMessage{
				Version:       warp.TeleportVersion,
				MessageType:   99, // Invalid type
				SourceChainID: ids.GenerateTestID(),
				DestChainID:   ids.GenerateTestID(),
				Nonce:         1,
				Payload:       []byte("payload"),
			},
			expectError: warp.ErrInvalidTeleportType,
		},
		{
			name: "empty payload",
			msg: &warp.TeleportMessage{
				Version:       warp.TeleportVersion,
				MessageType:   warp.TeleportTransfer,
				SourceChainID: ids.GenerateTestID(),
				DestChainID:   ids.GenerateTestID(),
				Nonce:         1,
				Payload:       []byte{}, // Empty
			},
			expectError: warp.ErrMissingPayload,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.expectError != nil {
				require.ErrorIs(err, tt.expectError)
			} else {
				require.NoError(err)
			}
		})
	}
}

// TestTeleportMessageToWarpMessage tests conversion to Warp message
func TestTeleportMessageToWarpMessage(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewWarpCodec()

	sourceChain := ids.GenerateTestID()
	destChain := ids.GenerateTestID()
	payload := []byte("test payload for warp")
	networkID := uint32(96369)

	teleport := warp.NewTeleportMessage(warp.TeleportLock, sourceChain, destChain, 100, payload)

	warpMsg, err := teleport.ToWarpMessage(c, networkID)
	require.NoError(err)
	require.NotNil(warpMsg)
	require.Equal(networkID, warpMsg.NetworkID)
	require.Equal(sourceChain, warpMsg.SourceChainID)
	require.NotEmpty(warpMsg.Payload)
}

// TestNewPrivateTeleportMessage tests encrypted message creation
func TestNewPrivateTeleportMessage(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewWarpCodec()

	sourceChain := ids.GenerateTestID()
	destChain := ids.GenerateTestID()
	payload := []byte("confidential cross-chain data")
	nonce := uint64(42)

	// Generate real ML-KEM-768 key pair
	scheme := mlkem768.Scheme()
	pubKey, _, err := scheme.GenerateKeyPair()
	require.NoError(err)
	recipientPubKey, err := pubKey.MarshalBinary()
	require.NoError(err)
	recipientKeyID := []byte("recipient-key-123")

	msg, err := warp.NewPrivateTeleportMessage(c, sourceChain, destChain, nonce, payload, recipientPubKey, recipientKeyID)
	require.NoError(err)
	require.NotNil(msg)

	require.Equal(warp.TeleportVersion, msg.Version)
	require.Equal(warp.TeleportPrivate, msg.MessageType)
	require.True(msg.Encrypted)
	require.NotEmpty(msg.Payload)
	require.NotEqual(payload, msg.Payload) // Should be encrypted
}

// TestPrivateTeleportMessageDecrypt tests decryption of private messages
func TestPrivateTeleportMessageDecrypt(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewWarpCodec()

	sourceChain := ids.GenerateTestID()
	destChain := ids.GenerateTestID()
	originalPayload := []byte("secret message for cross-chain transfer")
	nonce := uint64(999)

	// Generate real ML-KEM-768 key pair
	scheme := mlkem768.Scheme()
	pubKey, privKey, err := scheme.GenerateKeyPair()
	require.NoError(err)
	recipientPubKey, err := pubKey.MarshalBinary()
	require.NoError(err)
	recipientPrivKey, err := privKey.MarshalBinary()
	require.NoError(err)
	recipientKeyID := []byte("test-key")

	// Create encrypted message
	msg, err := warp.NewPrivateTeleportMessage(c, sourceChain, destChain, nonce, originalPayload, recipientPubKey, recipientKeyID)
	require.NoError(err)
	require.True(msg.Encrypted)

	// Decrypt
	decrypted, err := msg.DecryptPayload(c, recipientPrivKey)
	require.NoError(err)
	require.Equal(originalPayload, decrypted)
}

// TestTeleportMessageString tests string representation
func TestTeleportMessageString(t *testing.T) {
	require := require.New(t)

	msg := warp.NewTeleportMessage(
		warp.TeleportSwap,
		ids.GenerateTestID(),
		ids.GenerateTestID(),
		123,
		[]byte("swap data"),
	)

	str := msg.String()
	require.Contains(str, "Teleport")
	require.Contains(str, "Swap")
	require.Contains(str, "nonce=123")
}

// TestTeleportMessageCodecRoundTrip tests serialization
func TestTeleportMessageCodecRoundTrip(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewWarpCodec()

	original := &warp.TeleportMessage{
		Version:       warp.TeleportVersion,
		MessageType:   warp.TeleportGovernance,
		SourceChainID: ids.GenerateTestID(),
		DestChainID:   ids.GenerateTestID(),
		Nonce:         12345,
		Payload:       []byte("governance vote payload"),
		Encrypted:     false,
	}

	// Encode
	encoded, err := c.Marshal(warp.CodecVersion, original)
	require.NoError(err)

	// Decode
	decoded := &warp.TeleportMessage{}
	_, err = c.Unmarshal(encoded, decoded)
	require.NoError(err)

	// Verify
	require.Equal(original.Version, decoded.Version)
	require.Equal(original.MessageType, decoded.MessageType)
	require.Equal(original.SourceChainID, decoded.SourceChainID)
	require.Equal(original.DestChainID, decoded.DestChainID)
	require.Equal(original.Nonce, decoded.Nonce)
	require.Equal(original.Payload, decoded.Payload)
	require.Equal(original.Encrypted, decoded.Encrypted)
}

// TestTeleportTransferPayload tests transfer payload handling
func TestTeleportTransferPayload(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewWarpCodec()

	assetID := ids.GenerateTestID()
	amount := uint64(1000000)
	sender := []byte("0x1234567890abcdef")
	recipient := []byte("0xfedcba0987654321")
	fee := uint64(100)
	memo := []byte("test transfer")

	payload := warp.NewTransferPayload(assetID, amount, sender, recipient, fee, memo)

	require.Equal(assetID, payload.AssetID)
	require.Equal(amount, payload.Amount)
	require.Equal(sender, payload.Sender)
	require.Equal(recipient, payload.Recipient)
	require.Equal(fee, payload.Fee)
	require.Equal(memo, payload.Memo)

	// Test serialization
	encoded, err := payload.Bytes(c)
	require.NoError(err)

	// Test parsing
	parsed, err := warp.ParseTransferPayload(c, encoded)
	require.NoError(err)
	require.Equal(payload.AssetID, parsed.AssetID)
	require.Equal(payload.Amount, parsed.Amount)
	require.Equal(payload.Sender, parsed.Sender)
	require.Equal(payload.Recipient, parsed.Recipient)
	require.Equal(payload.Fee, parsed.Fee)
	require.Equal(payload.Memo, parsed.Memo)
}

// TestTeleportAttestPayload tests attestation payload handling
func TestTeleportAttestPayload(t *testing.T) {
	require := require.New(t)
	c := pvmcodectest.NewWarpCodec()

	payload := &warp.TeleportAttestPayload{
		AttestationType: 1,
		Timestamp:       1234567890,
		Data:            []byte("price: 100.50 USD"),
		AttesterID:      ids.GenerateTestNodeID(),
	}

	// Encode
	encoded, err := c.Marshal(warp.CodecVersion, payload)
	require.NoError(err)

	// Decode
	decoded := &warp.TeleportAttestPayload{}
	_, err = c.Unmarshal(encoded, decoded)
	require.NoError(err)

	// Verify
	require.Equal(payload.AttestationType, decoded.AttestationType)
	require.Equal(payload.Timestamp, decoded.Timestamp)
	require.Equal(payload.Data, decoded.Data)
	require.Equal(payload.AttesterID, decoded.AttesterID)
}

// TestSignatureType tests signature type utilities
func TestSignatureType(t *testing.T) {
	require := require.New(t)

	// Test recommended type
	recommended := warp.RecommendedSignatureType()
	require.Equal(warp.SigTypeCorona, recommended)

	// Test quantum safety
	require.False(warp.SigTypeBLS.IsQuantumSafe())
	require.True(warp.SigTypeCorona.IsQuantumSafe())
	require.True(warp.SigTypeHybrid.IsQuantumSafe())

	// Test string representation
	require.Equal("BLS", warp.SigTypeBLS.String())
	require.Equal("Corona", warp.SigTypeCorona.String())
	require.Equal("Hybrid", warp.SigTypeHybrid.String())
}

// TestTeleportTypes tests all teleport types
func TestTeleportTypes(t *testing.T) {
	require := require.New(t)

	// Verify constants are sequential
	require.Equal(warp.TeleportType(0), warp.TeleportTransfer)
	require.Equal(warp.TeleportType(1), warp.TeleportSwap)
	require.Equal(warp.TeleportType(2), warp.TeleportLock)
	require.Equal(warp.TeleportType(3), warp.TeleportUnlock)
	require.Equal(warp.TeleportType(4), warp.TeleportAttest)
	require.Equal(warp.TeleportType(5), warp.TeleportGovernance)
	require.Equal(warp.TeleportType(6), warp.TeleportPrivate)
}

// TestTeleportMessageAllTypes tests creating messages of all types
func TestTeleportMessageAllTypes(t *testing.T) {
	require := require.New(t)

	types := []warp.TeleportType{
		warp.TeleportTransfer,
		warp.TeleportSwap,
		warp.TeleportLock,
		warp.TeleportUnlock,
		warp.TeleportAttest,
		warp.TeleportGovernance,
	}

	for _, tt := range types {
		msg := warp.NewTeleportMessage(
			tt,
			ids.GenerateTestID(),
			ids.GenerateTestID(),
			uint64(tt), // Use type as nonce
			[]byte("payload"),
		)

		err := msg.Validate()
		require.NoError(err, "type %d should be valid", tt)
	}
}

// TestTeleportVersion tests version constant
func TestTeleportVersion(t *testing.T) {
	require := require.New(t)
	require.Equal(uint8(1), warp.TeleportVersion)
}
