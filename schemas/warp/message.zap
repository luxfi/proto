# Copyright (C) 2026, Lux Industries Inc. All rights reserved.
# See the file LICENSE for licensing terms.
#
# Lux Interchain Messaging (Warp) message schemas.
#
# Consumed by: github.com/luxfi/proto/p/warp and downstream node
# package github.com/luxfi/node/vms/platformvm/warp + Warp-using
# VMs (coreth atomic, xvm import/export). Codegen emits *_zap.go
# siblings next to the generated Go consumers.
#
# Wire envelope convention: [TypeKind:1][ShapeKind:1][ZAP message: N].
# Warp outer envelopes carry their own TypeKind on the discriminator;
# the inner Payload bytes carry an AddressedCall / Hash discriminator
# from the payload codec.
#
# Two layered codec slot maps coexist:
#
#   - Warp signature codec (warp/codec.go) registers (in order):
#       0x00 BitSetSignature
#       0x01 CoronaSignature
#       0x02 EncryptedWarpPayload
#       0x03 HybridBLSCoronaSignature  (DEPRECATED)
#       0x04 TeleportMessage
#       0x05 TeleportTransferPayload
#       0x06 TeleportAttestPayload
#
#   - Warp message codec (warp/message/codec.go) registers (in order):
#       0x00 ChainToL1Conversion
#       0x01 RegisterL1Validator
#       0x02 L1ValidatorRegistration
#       0x03 L1ValidatorWeight
#
#   - Warp payload codec (warp/payload/codec.go) registers (in order):
#       0x00 Hash
#       0x01 AddressedCall
#
# Each ShapeKind value below names the underlying wire type. Reordering
# any of the upstream codec lines is a hard fork — append-only.
#
# Source mapping:
#   node/vms/platformvm/warp/unsigned_message.go      -> UnsignedMessage
#   node/vms/platformvm/warp/signature.go             -> BitSetSignature, CoronaSignature,
#                                                        HybridBLSCoronaSignature, EncryptedWarpPayload
#   node/vms/platformvm/warp/teleport.go              -> TeleportMessage, TeleportTransferPayload,
#                                                        TeleportAttestPayload
#   node/vms/platformvm/warp/payload/addressed_call.go -> AddressedCall
#   node/vms/platformvm/warp/payload/hash.go          -> Hash
#   node/vms/platformvm/warp/message/chain_to_l1_conversion.go -> ChainToL1Conversion + data
#   node/vms/platformvm/warp/message/register_l1_validator.go  -> RegisterL1Validator + PChainOwner
#   node/vms/platformvm/warp/message/l1_validator_registration.go -> L1ValidatorRegistration
#   node/vms/platformvm/warp/message/l1_validator_weight.go    -> L1ValidatorWeight

package message

# ----------------------------------------------------------------------
# Primitive aliases
# ----------------------------------------------------------------------
type id32 = bytes_fixed[32]
type id20 = bytes_fixed[20]
type bls48 = bytes_fixed[48]     # bls.PublicKey compressed (G1)
type sig96 = bytes_fixed[96]     # bls.Signature (G2)

# ======================================================================
# Unsigned + signed Warp messages
# ======================================================================

# ----------------------------------------------------------------------
# UnsignedMessage — outer Warp envelope before signature aggregation
# ----------------------------------------------------------------------
#
# The canonical bytes that BLS / Corona aggregate signatures cover.
# ID = hash(Bytes) where Bytes is the codec marshal of this struct
# under warp.CodecVersion.
#
# Fixed section: 4 (NetworkID) + 32 (SourceChainID) + 8 (Payload) = 44
struct UnsignedMessage
    shape_kind = 0x80       # ShapeKindWarpUnsignedMessage
{
    # NetworkID is the Lux primary network ID the message was emitted
    # on. Verification rejects messages with NetworkID != local
    # (warp.ErrWrongNetworkID).
    NetworkID     u32               @0

    # SourceChainID is the 32-byte chain ID the message originated
    # on. Determines which canonical validator set verifies the
    # aggregate signature.
    SourceChainID id32              @4

    # Payload is the application-level bytes — typically an
    # AddressedCall or Hash payload (see payload codec slots above).
    Payload       bytes             @36
}

# ----------------------------------------------------------------------
# BitSetSignature — Warp 1.0 classical BLS aggregate signature
# ----------------------------------------------------------------------
#
# typeID 0x00 in the warp codec slot map. Classical BLS aggregate
# signature with a signer-bitset selector indexing into the canonical
# validator set at the message's SourceChainID + height.
#
# Fixed section: 8 (Signers) + 96 (Signature) = 104
struct BitSetSignature
    shape_kind = 0x81       # ShapeKindWarpBitSetSignature
{
    # Signers is a big-endian byte slice encoding a bitmap of which
    # canonical validators contributed to the aggregate. Length MUST
    # equal len(set.BitsFromBytes(Signers).Bytes()) — no extraneous
    # leading-zero padding.
    Signers   bytes                 @0

    # Signature is the 96-byte BLS aggregate signature (G2) over the
    # UnsignedMessage bytes.
    Signature sig96                 @8
}

# ----------------------------------------------------------------------
# CoronaSignature — Warp 1.5 post-quantum lattice signature
# ----------------------------------------------------------------------
#
# typeID 0x01 in the warp codec slot map. Lattice-based threshold
# signature (LWE / Ring-LWE; ref github.com/luxfi/corona). Replaces
# BitSetSignature for new Warp messages.
#
# Fixed section: 8 (Signers) + 8 (Signature) = 16
struct CoronaSignature
    shape_kind = 0x82       # ShapeKindWarpCoronaSignature
{
    # Signers is a big-endian bitmap of contributing validators.
    Signers   bytes                 @0

    # Signature is the Corona threshold-signature blob. Variable
    # length — depends on threshold parameters (M, N, Dbar, Kappa).
    # Carries (c challenge polynomial, z response vector, Delta hint
    # vector) per the Corona protocol.
    Signature bytes                 @8
}

# ----------------------------------------------------------------------
# HybridBLSCoronaSignature — DEPRECATED transitional hybrid
# ----------------------------------------------------------------------
#
# typeID 0x03 in the warp codec slot map. Both BLS and Corona
# signatures MUST verify for the hybrid to be accepted.
#
# Deprecated: use CoronaSignature. Retained because the typeID is
# committed to history — reordering this line is a hard fork.
#
# Fixed section: 8 (Signers) + 96 (BLSSignature) +
# 8 (CoronaSignature) + 8 (CoronaPublicKeys) = 120
struct HybridBLSCoronaSignature
    shape_kind = 0x83       # ShapeKindWarpHybridBLSCoronaSignature
{
    Signers          bytes          @0

    # BLSSignature is the 96-byte BLS aggregate signature.
    BLSSignature     sig96          @8

    # CoronaSignature is the aggregated Corona lattice signature.
    CoronaSignature  bytes          @104

    # CoronaPublicKeys is the list of per-signer Corona public keys,
    # in the same order as the Signers bitset enumerates. Needed
    # because validators may have distinct Corona vs BLS keys.
    CoronaPublicKeys bytes          @112
}

# ----------------------------------------------------------------------
# EncryptedWarpPayload — ML-KEM-768 + AES-256-GCM sealed payload
# ----------------------------------------------------------------------
#
# typeID 0x02 in the warp codec slot map. Wraps a confidential
# cross-chain payload (private bridges, sealed bids, MEV-protected
# intents). Carried inside a TeleportMessage when
# MessageType=TeleportPrivate.
#
# Fixed section: 8 (EncapsulatedKey) + 8 (Nonce) +
# 8 (Ciphertext) + 8 (RecipientKeyID) = 32
struct EncryptedWarpPayload
    shape_kind = 0x84       # ShapeKindWarpEncryptedPayload
{
    # EncapsulatedKey is the ML-KEM-768 ciphertext carrying the
    # encapsulated shared secret. Length MUST equal 1088 bytes
    # (MLKEM768CiphertextLen) at runtime.
    EncapsulatedKey bytes           @0

    # Nonce is the 12-byte AES-GCM nonce.
    Nonce           bytes           @8

    # Ciphertext is the AES-256-GCM encrypted payload, including the
    # trailing 16-byte authentication tag.
    Ciphertext      bytes           @16

    # RecipientKeyID identifies the recipient's ML-KEM public key
    # (typically a hash of the public key) so recipients know which
    # private key to use for decryption.
    RecipientKeyID  bytes           @24
}

# ======================================================================
# Payload codec slot map (warp/payload/codec.go)
# ======================================================================

# ----------------------------------------------------------------------
# AddressedCall — addressed cross-chain call payload
# ----------------------------------------------------------------------
#
# typeID 0x01 in the warp payload codec. Carries a source address +
# arbitrary application payload. Destination address (if any) is
# encoded inside Payload, not at this layer.
#
# Fixed section: 8 (SourceAddress) + 8 (Payload) = 16
struct AddressedCall
    shape_kind = 0x85       # ShapeKindWarpAddressedCall
{
    # SourceAddress is the issuing address on the source chain
    # (chain-specific encoding).
    SourceAddress bytes             @0

    # Payload is the inner application payload bytes.
    Payload       bytes             @8
}

# ----------------------------------------------------------------------
# Hash — finality-relay block-hash payload
# ----------------------------------------------------------------------
#
# typeID 0x00 in the warp payload codec. Single-field payload carrying
# a block hash to attest finality of.
#
# Fixed section: 32 (Hash) = 32
struct Hash
    shape_kind = 0x86       # ShapeKindWarpHash
{
    # Hash is the 32-byte block-hash being attested.
    Hash id32                       @0
}

# ======================================================================
# Message codec slot map (warp/message/codec.go) — L1 lifecycle messages
# ======================================================================

# ----------------------------------------------------------------------
# ChainToL1Conversion — primary-net summary of a chain conversion
# ----------------------------------------------------------------------
#
# typeID 0x00 in the warp message codec. P-chain emits this message
# after a successful ConvertNetworkToL1Tx, communicating the conversion
# to the receiving L1.
#
# Carried payload is the ID (hash of the ChainToL1ConversionData blob).
#
# Fixed section: 32 (ID) = 32
struct ChainToL1Conversion
    shape_kind = 0x87       # ShapeKindWarpChainToL1Conversion
{
    # ID is the canonical conversion ID — hash(ChainToL1ConversionData
    # bytes). Receiving L1s compute the same hash from their stored
    # conversion data to verify.
    ID id32                         @0
}

# ----------------------------------------------------------------------
# ChainToL1ConversionData — payload hashed into ChainToL1Conversion.ID
# ----------------------------------------------------------------------
#
# Not a wire-codec slot — this is the canonical pre-image whose hash
# becomes ChainToL1Conversion.ID. Both sides marshal this and compare
# the hash to verify.
#
# Fixed section: 32 (ChainID) + 32 (ManagerChainID) +
# 8 (ManagerAddress) + 8 (Validators) = 80
struct ChainToL1ConversionData
    shape_kind = 0x88       # ShapeKindWarpChainToL1ConversionData
{
    ChainID        id32             @0
    ManagerChainID id32             @32
    ManagerAddress bytes            @64

    # Validators: list of ChainToL1ConversionValidatorData.
    Validators     bytes            @72
}

# ----------------------------------------------------------------------
# ChainToL1ConversionValidatorData — one validator entry
# ----------------------------------------------------------------------
#
# Per-validator data hashed into the conversion ID. Wire-encoded as a
# JSONByteSlice NodeID + fixed-length BLS pubkey + weight.
#
# Fixed section: 8 (NodeID) + 48 (BLSPublicKey) + 8 (Weight) = 64
struct ChainToL1ConversionValidatorData
    shape_kind = 0x89       # ShapeKindWarpChainToL1ConversionValidatorData
{
    # NodeID is the 20-byte short-id carried as variable bytes
    # (JSONByteSlice in the Go type).
    NodeID       bytes              @0

    # BLSPublicKey is the validator's 48-byte compressed BLS public
    # key. Fixed-length wire field.
    BLSPublicKey bls48              @8

    # Weight is the validator's sampling weight at conversion time.
    Weight       u64                @56
}

# ----------------------------------------------------------------------
# RegisterL1Validator — request to add a validator to an L1
# ----------------------------------------------------------------------
#
# typeID 0x01 in the warp message codec. Issued by an L1's
# validator-manager contract to request a new validator be registered
# on the P-chain. Carried as the inner payload of an AddressedCall
# carried by a Warp UnsignedMessage carried as the Message field of a
# RegisterL1ValidatorTx.
#
# ValidationID = hash(this payload bytes) — used to key the L1Validator
# state record.
#
# Fixed section: 32 (ChainID) + 8 (NodeID) + 48 (BLSPublicKey) +
# 8 (Expiry) + 12 (RemainingBalanceOwner) + 12 (DisableOwner) +
# 8 (Weight) = 128
struct RegisterL1Validator
    shape_kind = 0x8A       # ShapeKindWarpRegisterL1Validator
{
    # ChainID is the L1 chain ID this validator is being registered
    # on. MUST NOT be PrimaryNetworkID (warp.ErrInvalidChainID).
    ChainID               id32      @0

    # NodeID is the validator's 20-byte short-id (variable-length
    # JSONByteSlice in the Go type).
    NodeID                bytes     @32

    # BLSPublicKey is the validator's 48-byte compressed BLS public
    # key.
    BLSPublicKey          bls48     @40

    # Expiry is the unix-seconds deadline after which this
    # registration request becomes invalid (replay-protected via the
    # state ExpiryEntry record).
    Expiry                u64       @88

    # RemainingBalanceOwner is the PChainOwner that receives leftover
    # balance when the validator exits the set.
    RemainingBalanceOwner bytes     @96

    # DisableOwner is the PChainOwner with authority to manually
    # deactivate this validator.
    DisableOwner          bytes     @108

    # Weight is the validator's sampling weight. MUST be > 0
    # (warp.ErrInvalidWeight).
    Weight                u64       @120
}

# ----------------------------------------------------------------------
# PChainOwner — threshold + addresses owner record (warp side)
# ----------------------------------------------------------------------
#
# Carried inside RegisterL1Validator (twice — RemainingBalanceOwner +
# DisableOwner) and on the P-chain by ConvertNetworkToL1Validator.
# Identical shape and semantics to pvm/txs.PChainOwner.
#
# Fixed section: 4 (Threshold) + 8 (Addresses) = 12
struct PChainOwner
    shape_kind = 0x8B       # ShapeKindWarpPChainOwner
{
    Threshold u32                   @0

    # Addresses is the list of 20-byte ShortIDs authorised on this
    # owner. Sorted+unique.
    Addresses bytes                 @4
}

# ----------------------------------------------------------------------
# L1ValidatorRegistration — registration ack from the P-chain
# ----------------------------------------------------------------------
#
# typeID 0x02 in the warp message codec. P-chain emits this to ack
# whether a ValidationID is currently a validator on the P-chain side.
#
# Registered=false is FINAL — that ValidationID can never become a
# validator on the P-chain (the request has expired or been replaced).
# It is still possible that ValidationID was previously a validator.
#
# Fixed section: 32 (ValidationID) + 1 (Registered) = 33
struct L1ValidatorRegistration
    shape_kind = 0x8C       # ShapeKindWarpL1ValidatorRegistration
{
    ValidationID id32               @0

    # Registered: 1 = currently registered, 0 = never will be.
    # Encoded as a Go bool / wire u8.
    Registered   u8                 @32
}

# ----------------------------------------------------------------------
# L1ValidatorWeight — bidirectional weight-update message
# ----------------------------------------------------------------------
#
# typeID 0x03 in the warp message codec.
#
# When the P-chain RECEIVES this from an L1's manager: treat as a
# command to set the validator's weight to Weight at nonce Nonce.
#
# When the P-chain SENDS this to an L1 / observer: report the current
# Nonce+Weight pair.
#
# Reserved: Nonce == MaxUint64 is removal-only (Weight MUST be 0).
# Verify() returns ErrNonceReservedForRemoval otherwise.
#
# Fixed section: 32 (ValidationID) + 8 (Nonce) + 8 (Weight) = 48
struct L1ValidatorWeight
    shape_kind = 0x8D       # ShapeKindWarpL1ValidatorWeight
{
    ValidationID id32               @0
    Nonce        u64                @32
    Weight       u64                @40
}

# ======================================================================
# Teleport: cross-chain bridge messages (warp codec slots 0x04-0x06)
# ======================================================================

# ----------------------------------------------------------------------
# TeleportMessage — Warp wrapper for cross-chain bridge ops
# ----------------------------------------------------------------------
#
# typeID 0x04 in the warp codec slot map. High-level cross-chain
# messaging primitive sitting on top of Warp signatures.
#
# Version is currently TeleportVersion=1; MessageType selects the
# operation kind. When Encrypted=1 (TeleportPrivate), Payload is
# a codec-marshalled EncryptedWarpPayload.
#
# Fixed section: 1 (Version) + 1 (MessageType) + 32 (SourceChainID) +
# 32 (DestChainID) + 8 (Nonce) + 8 (Payload) + 1 (Encrypted) = 83
struct TeleportMessage
    shape_kind = 0x8E       # ShapeKindWarpTeleportMessage
{
    # Version is the Teleport protocol version (currently 1).
    Version       u8                @0

    # MessageType selects the cross-chain operation:
    #   0 Transfer    — standard asset transfer
    #   1 Swap        — atomic swap
    #   2 Lock        — lock on source
    #   3 Unlock      — unlock on destination
    #   4 Attest      — oracle attestation
    #   5 Governance  — cross-chain governance
    #   6 Private     — encrypted transfer
    MessageType   u8                @1

    SourceChainID id32              @2
    DestChainID   id32              @34

    # Nonce prevents replay across (Source, Dest) pairs.
    Nonce         u64               @66

    # Payload is the application payload — bytes of an
    # EncryptedWarpPayload when Encrypted=1, otherwise free-form.
    Payload       bytes             @74

    # Encrypted: 1 indicates Payload is an EncryptedWarpPayload, 0
    # plain bytes.
    Encrypted     u8                @82
}

# ----------------------------------------------------------------------
# TeleportTransferPayload — asset-transfer Teleport payload
# ----------------------------------------------------------------------
#
# typeID 0x05 in the warp codec slot map. Carried as the Payload of a
# TeleportMessage when MessageType=TeleportTransfer.
#
# Fixed section: 32 (AssetID) + 8 (Amount) + 8 (Sender) +
# 8 (Recipient) + 8 (Fee) + 8 (Memo) = 72
struct TeleportTransferPayload
    shape_kind = 0x8F       # ShapeKindWarpTeleportTransferPayload
{
    # AssetID is the 32-byte asset ID being transferred.
    AssetID   id32                  @0

    # Amount is the quantity being transferred (asset-defined units).
    Amount    u64                   @32

    # Sender is the chain-specific source address bytes.
    Sender    bytes                 @40

    # Recipient is the chain-specific destination address bytes.
    Recipient bytes                 @48

    # Fee paid for the bridge operation.
    Fee       u64                   @56

    # Memo is optional metadata.
    Memo      bytes                 @64
}

# ----------------------------------------------------------------------
# TeleportAttestPayload — oracle-attestation Teleport payload
# ----------------------------------------------------------------------
#
# typeID 0x06 in the warp codec slot map. Carried as the Payload of a
# TeleportMessage when MessageType=TeleportAttest (oracle feed,
# compute attestation, ...).
#
# Fixed section: 1 (AttestationType) + 8 (Timestamp) + 8 (Data) +
# 20 (AttesterID) = 37
struct TeleportAttestPayload
    shape_kind = 0x90       # ShapeKindWarpTeleportAttestPayload
{
    # AttestationType identifies what is being attested
    # (application-defined; e.g. 1=price feed, 2=compute result, ...).
    AttestationType u8              @0

    # Timestamp is the unix-seconds time of the attestation.
    Timestamp       u64             @1

    # Data is the application-defined attestation payload (price,
    # compute result, ...).
    Data            bytes           @9

    # AttesterID is the NodeID of the attesting validator.
    AttesterID      id20            @17
}
