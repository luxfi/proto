# Copyright (C) 2026, Lux Industries Inc. All rights reserved.
# See the file LICENSE for licensing terms.
#
# P-chain (platformvm) block envelope schemas.
#
# Consumed by: github.com/luxfi/proto/p/block and downstream node
# package github.com/luxfi/node/vms/platformvm/block. Codegen emits
# *_zap.go siblings next to the generated Go consumers.
#
# Wire envelope convention: [TypeKind:1][ShapeKind:1][ZAP message: N].
# Block envelopes carry their own TypeKind on the outer envelope; the
# inner Txs/Tx fields carry signed-tx envelopes that dispatch on the
# tx ShapeKind set declared in pvm/txs.zap.
#
# Block parser dispatch (node/vms/platformvm/block/parse.go) reads the
# 2-byte codec version prefix and routes either to v0Codec (legacy
# Apricot/Banff layout) or to the current v1 Codec — there is no single
# discriminator that distinguishes Standard / Proposal / Commit / Abort
# above the codec slot map. The schema below mirrors that, declaring
# each block kind as a standalone struct.
#
# Two eras coexist:
#
#   - v1 (current write target) — block/codec.go registers
#     ProposalBlock, AbortBlock, CommitBlock, StandardBlock under
#     CodecVersionV1. Every block produced post-codec-v1 is one of
#     these four kinds.
#   - v0 (read-only) — block/v0/types.go declares the historical
#     Apricot/Banff layout still on disk + still parsed for chain
#     history. ApricotAtomicBlock is dead on the modern P-only network
#     but appears in pre-Banff archives.
#
# Source mapping:
#   node/vms/platformvm/block/common_block.go     -> CommonBlock
#   node/vms/platformvm/block/standard_block.go   -> StandardBlock
#   node/vms/platformvm/block/proposal_block.go   -> ProposalBlock
#   node/vms/platformvm/block/commit_block.go     -> CommitBlock
#   node/vms/platformvm/block/abort_block.go      -> AbortBlock
#   node/vms/platformvm/block/v0/types.go         -> Apricot*/Banff*

package block

# ----------------------------------------------------------------------
# Primitive aliases
# ----------------------------------------------------------------------
type id32 = bytes_fixed[32]

# ----------------------------------------------------------------------
# CommonBlock — shared parent-id + height fields
# ----------------------------------------------------------------------
#
# Embedded by every block kind (v0 and v1). Wire-equal across eras.
# BlockID is derived (hash(blockBytes)) — NOT serialised — so it does
# not appear here.
#
# Fixed section: 32 (PrntID) + 8 (Hght) = 40
struct CommonBlock
    shape_kind = 0x50       # ShapeKindPCommonBlock
{
    # PrntID is the 32-byte BlockID of the parent block.
    PrntID id32                      @0

    # Hght is the block height. Genesis is at height 0.
    Hght   u64                       @32
}

# ======================================================================
# v1 canonical blocks (block/codec.go CodecVersionV1)
# ======================================================================
#
# Each v1 block carries an in-header Time field (Banff-style) and one
# of: a tx list (StandardBlock), a tx list + proposal Tx (ProposalBlock),
# or no payload (CommitBlock / AbortBlock).

# ----------------------------------------------------------------------
# StandardBlock — N decision txs, carries the chain timestamp
# ----------------------------------------------------------------------
#
# Carries an ordered list of standard txs (Add*Validator, CreateChain,
# Import/Export, ...). The block's Time field replaces the legacy
# AdvanceTimeTx flow.
#
# Wire field order matches the embedded-CommonBlock layout the codec
# emits: Time, then CommonBlock fields, then Transactions.
#
# Fixed section: 8 (Time) + 40 (CommonBlock) + 8 (Transactions) = 56
struct StandardBlock
    shape_kind = 0x51       # ShapeKindPStandardBlock
{
    # Time is the unix-seconds timestamp this block proposes the
    # chain advance to.
    Time         u64                 @0

    # CommonBlock fields are embedded inline.
    PrntID       id32                @8
    Hght         u64                 @40

    # Transactions is the ordered list of signed-tx envelopes carried
    # by this block.
    Transactions bytes               @48
}

# ----------------------------------------------------------------------
# ProposalBlock — one proposal tx + tail of decision txs
# ----------------------------------------------------------------------
#
# Carries exactly one proposal Tx (AdvanceTimeTx pre-Banff;
# RewardValidatorTx Banff+) plus a tail of decision Txs that commit
# atomically with the proposal outcome.
#
# A ProposalBlock MUST be followed by either a CommitBlock (the
# proposal's effect is accepted) or an AbortBlock (rejected).
#
# Fixed section: 8 (Time) + 8 (Transactions) + 40 (CommonBlock) +
# 4 (Tx) = 60
struct ProposalBlock
    shape_kind = 0x52       # ShapeKindPProposalBlock
{
    Time         u64                 @0

    # Transactions is the decision-tx tail. May be empty.
    Transactions bytes               @8

    # CommonBlock fields are embedded inline.
    PrntID       id32                @16
    Hght         u64                 @48

    # Tx is the single proposal Tx. Dispatch on the inner Tx
    # ShapeKind: RewardValidatorTx (0x33) on Banff+, AdvanceTimeTx
    # (0x34) on pre-Banff archives.
    Tx           bytes               @56
}

# ----------------------------------------------------------------------
# CommitBlock — accept the parent ProposalBlock's proposal
# ----------------------------------------------------------------------
#
# Carries no payload — its acceptance is the proposal outcome.
#
# Fixed section: 8 (Time) + 40 (CommonBlock) = 48
struct CommitBlock
    shape_kind = 0x53       # ShapeKindPCommitBlock
{
    Time   u64                       @0
    PrntID id32                      @8
    Hght   u64                       @40
}

# ----------------------------------------------------------------------
# AbortBlock — reject the parent ProposalBlock's proposal
# ----------------------------------------------------------------------
#
# Carries no payload.
#
# Fixed section: 8 (Time) + 40 (CommonBlock) = 48
struct AbortBlock
    shape_kind = 0x54       # ShapeKindPAbortBlock
{
    Time   u64                       @0
    PrntID id32                      @8
    Hght   u64                       @40
}

# ======================================================================
# v0 read-only blocks (block/v0/types.go CodecVersionV0)
# ======================================================================
#
# These shapes are decoded ONLY — block/codec.go gcV0 codec is read-only,
# write paths target v1 exclusively. The schemas below mirror the
# pre-codec-v1 v1.23.x Apricot/Banff layout still on disk. Block IDs
# derived from these bytes remain stable across the v0->v1 migration.

# ----------------------------------------------------------------------
# ApricotProposalBlock — pre-Banff proposal block (slot 0)
# ----------------------------------------------------------------------
#
# Apricot-era proposal block. No header timestamp — pre-Banff blocks
# advance time only via the embedded AdvanceTimeTx proposal Tx.
#
# Fixed section: 40 (CommonBlock) + 4 (Tx) = 44
struct ApricotProposalBlock
    shape_kind = 0x55       # ShapeKindPApricotProposalBlock
{
    PrntID id32                      @0
    Hght   u64                       @32

    # Tx is the single proposal Tx (typically AdvanceTimeTx in
    # Apricot history).
    Tx     bytes                     @40
}

# ----------------------------------------------------------------------
# ApricotAbortBlock — pre-Banff abort block (slot 1)
# ----------------------------------------------------------------------
#
# Fixed section: 40 (CommonBlock) = 40
struct ApricotAbortBlock
    shape_kind = 0x56       # ShapeKindPApricotAbortBlock
{
    PrntID id32                      @0
    Hght   u64                       @32
}

# ----------------------------------------------------------------------
# ApricotCommitBlock — pre-Banff commit block (slot 2)
# ----------------------------------------------------------------------
#
# Fixed section: 40 (CommonBlock) = 40
struct ApricotCommitBlock
    shape_kind = 0x57       # ShapeKindPApricotCommitBlock
{
    PrntID id32                      @0
    Hght   u64                       @32
}

# ----------------------------------------------------------------------
# ApricotStandardBlock — pre-Banff standard block (slot 3)
# ----------------------------------------------------------------------
#
# Pre-Banff standard block: no header timestamp, just the embedded
# tx list.
#
# Fixed section: 40 (CommonBlock) + 8 (Transactions) = 48
struct ApricotStandardBlock
    shape_kind = 0x58       # ShapeKindPApricotStandardBlock
{
    PrntID       id32                @0
    Hght         u64                 @32

    Transactions bytes               @40
}

# ----------------------------------------------------------------------
# ApricotAtomicBlock — pre-Banff atomic block (slot 4, DEAD)
# ----------------------------------------------------------------------
#
# Atomic-block flow is dead on the modern P-only network — these
# blocks only appear in pre-Banff archive history. Retained here only
# to keep historical bytes decoding.
#
# Fixed section: 40 (CommonBlock) + 4 (Tx) = 44
struct ApricotAtomicBlock
    shape_kind = 0x59       # ShapeKindPApricotAtomicBlock
{
    PrntID id32                      @0
    Hght   u64                       @32

    # Tx is the single atomic Tx (ImportTx or ExportTx).
    Tx     bytes                     @40
}

# ----------------------------------------------------------------------
# BanffProposalBlock — Banff-era proposal block (slot 29)
# ----------------------------------------------------------------------
#
# Wraps an ApricotProposalBlock with a Banff-era per-block Time
# header + a tail of decision Txs (Transactions). On-wire field order
# is Time, then Transactions, then the embedded ApricotProposalBlock.
#
# Fixed section: 8 (Time) + 8 (Transactions) + 44 (ApricotProposalBlock) = 60
struct BanffProposalBlock
    shape_kind = 0x5A       # ShapeKindPBanffProposalBlock
{
    Time         u64                 @0
    Transactions bytes               @8

    # ApricotProposalBlock embedded inline (PrntID, Hght, Tx).
    PrntID       id32                @16
    Hght         u64                 @48
    Tx           bytes               @56
}

# ----------------------------------------------------------------------
# BanffAbortBlock — Banff-era abort block (slot 30)
# ----------------------------------------------------------------------
#
# Fixed section: 8 (Time) + 40 (ApricotAbortBlock) = 48
struct BanffAbortBlock
    shape_kind = 0x5B       # ShapeKindPBanffAbortBlock
{
    Time   u64                       @0
    PrntID id32                      @8
    Hght   u64                       @40
}

# ----------------------------------------------------------------------
# BanffCommitBlock — Banff-era commit block (slot 31)
# ----------------------------------------------------------------------
#
# Fixed section: 8 (Time) + 40 (ApricotCommitBlock) = 48
struct BanffCommitBlock
    shape_kind = 0x5C       # ShapeKindPBanffCommitBlock
{
    Time   u64                       @0
    PrntID id32                      @8
    Hght   u64                       @40
}

# ----------------------------------------------------------------------
# BanffStandardBlock — Banff-era standard block (slot 32)
# ----------------------------------------------------------------------
#
# Fixed section: 8 (Time) + 48 (ApricotStandardBlock) = 56
struct BanffStandardBlock
    shape_kind = 0x5D       # ShapeKindPBanffStandardBlock
{
    Time         u64                 @0
    PrntID       id32                @8
    Hght         u64                 @40
    Transactions bytes               @48
}
