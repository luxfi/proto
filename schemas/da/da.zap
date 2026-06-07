# Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
# See the file LICENSE for licensing terms.
#
# Wire schemas for DA blobs + certificates, replacing the legacy json/v2
# database persistence. The hand-rolled codec lives in
# node/vms/da/codec_zap.go; this schema records the wire shape for the
# centralized registry and future codegen.

package da

type id32 = bytes_fixed[32]

# Wire layout per DABlob envelope (120 bytes fixed + variable tails).
struct DABlobWire {
    Kind         u8    @0       # = 1 (blob)
    Height       u64   @8
    TimestampNs  i64   @16
    ChunkCount   u32   @24
    ChunkSize    u32   @28
    ID           id32  @32
    Submitter    id32  @64
    Commitment   bytes @96
    Data         bytes @104
    Chunks       bytes @112      # concatenated ChunkEntry records
}

# DACertificate envelope (40 bytes fixed + variable tails).
struct DACertWire {
    Kind         u8    @0        # = 2 (cert)
    Threshold    u32   @4
    TimestampS   i64   @8
    Commitment   bytes @16        # inner DACommitmentWire
    SignerBitmap bytes @24
    Signatures   bytes @32        # u32-prefix-list of sig blobs
}

# DACommitment payload (nested inside DACertWire.Commitment).
struct DACommitmentWire {
    BlobID        id32  @0
    Height        u64   @32
    ChunkCount    u32   @40
    Commitment    bytes @44
    DataRoot      bytes @52
    ErasureRoot   bytes @60
    ValidatorSigs bytes @68
}
