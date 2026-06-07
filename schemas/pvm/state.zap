# Copyright (C) 2026, Lux Industries Inc. All rights reserved.
# See the file LICENSE for licensing terms.
#
# P-chain (platformvm) state-value schemas.
#
# These are the on-disk state records the platformvm writes to its
# state DB. Distinct from txs.zap (wire/tx envelopes) and block.zap
# (block envelopes) — these are KV-store value shapes only.
#
# Consumed by: github.com/luxfi/proto/p/state and downstream node
# package github.com/luxfi/node/vms/platformvm/state. Codegen emits
# *_zap.go siblings next to the generated Go consumers.
#
# Wire envelope convention: [TypeKind:1][ShapeKind:1][ZAP message: N].
# State values use TypeKindReserved (0x00) on the outer envelope.
#
# UTXO values are referenced — full UTXO schema lives in
# utxo/utxo.zap. Per-validator-set indexing uses the (chainID, nodeID)
# compound key; the layout below records that key shape so the on-disk
# format is documented in one place.
#
# Source mapping:
#   node/vms/platformvm/state/l1_validator.go         -> L1Validator
#   node/vms/platformvm/state/expiry.go               -> ExpiryEntry
#   node/vms/platformvm/state/chain_id_node_id.go     -> ChainIDNodeID
#   node/vms/platformvm/state/metadata_validator.go   -> ValidatorMetadata + variants
#   node/vms/platformvm/state/metadata_delegator.go   -> DelegatorMetadata
#   node/vms/platformvm/state/state.go ~L255          -> StateBlk (legacy)
#   node/vms/platformvm/state/state.go ~L481          -> HeightRange
#   node/vms/platformvm/state/state.go ~L486          -> ValidatorWeightDiff
#   node/vms/platformvm/state/state.go ~L516          -> TxBytesAndStatus
#   node/vms/platformvm/state/state.go ~L531          -> NetToL1Conversion

package state

# ----------------------------------------------------------------------
# Primitive aliases
# ----------------------------------------------------------------------
type id32 = bytes_fixed[32]
type id20 = bytes_fixed[20]

# ----------------------------------------------------------------------
# L1Validator — LP-77 L1 validator on-disk record
# ----------------------------------------------------------------------
#
# Stored under the L1 active/inactive prefixes keyed by ValidationID.
# ValidationID is the KV key and is NOT serialised inside the value.
#
# For a given ValidationID, the constant fields (ChainID, NodeID,
# PublicKey, RemainingBalanceOwner, DeactivationOwner, StartTime) MUST
# remain unchanged across writes (state.ErrMutatedL1Validator otherwise).
#
# Fixed section: 32 (ChainID) + 20 (NodeID) + 8 (PublicKey) +
# 8 (RemainingBalanceOwner) + 8 (DeactivationOwner) + 8 (StartTime) +
# 8 (Weight) + 8 (MinNonce) + 8 (EndAccumulatedFee) = 108
struct L1Validator
    shape_kind = 0x70       # ShapeKindPStateL1Validator
{
    # ChainID is the 32-byte L1 chain ID this validator is in.
    ChainID               id32       @0

    # NodeID is the 20-byte short-id of the validator's node.
    NodeID                id20       @32

    # PublicKey is the validator's uncompressed BLS public key.
    # Guaranteed populated for active L1 validators.
    PublicKey             bytes      @52

    # RemainingBalanceOwner is the codec-marshalled fx.Owner that
    # receives the validator's remaining balance after fee accrual,
    # at the time the validator is removed from the set.
    RemainingBalanceOwner bytes      @60

    # DeactivationOwner is the codec-marshalled fx.Owner authorised
    # to manually deactivate this validator.
    DeactivationOwner     bytes      @68

    # StartTime is the unix-seconds timestamp this validator was
    # added to the set.
    StartTime             u64        @76

    # Weight is the validator's sampling weight. Updates require
    # MinNonce++. Setting Weight=0 removes the validator (and uses
    # the reserved MaxUint64 nonce).
    Weight                u64        @84

    # MinNonce is the smallest nonce the next weight-update may use.
    # Reserved MaxUint64 = removal-only.
    MinNonce              u64        @92

    # EndAccumulatedFee is the total accrued fees per validator at
    # which this validator must be deactivated. 0 = inactive.
    EndAccumulatedFee     u64        @100
}

# ----------------------------------------------------------------------
# ExpiryEntry — RegisterL1Validator replay-protection record
# ----------------------------------------------------------------------
#
# Stored under ExpiryReplayProtectionPrefix. The (Timestamp,
# ValidationID) pair forms both the key AND value identity — values
# are written empty; the key bytes are this struct's marshal.
#
# Fixed section: 8 (Timestamp) + 32 (ValidationID) = 40
struct ExpiryEntry
    shape_kind = 0x71       # ShapeKindPStateExpiryEntry
{
    # Timestamp is the unix-seconds expiry deadline.
    Timestamp    u64                 @0

    # ValidationID is the hashed-payload ID of the
    # RegisterL1Validator warp message this expiry guards.
    ValidationID id32                @8
}

# ----------------------------------------------------------------------
# ChainIDNodeID — (chainID, nodeID) compound key
# ----------------------------------------------------------------------
#
# Compound key for the per-chain validator-set index
# (L1Validators.HasL1Validator lookup, chain validator state
# membership). Total 52 bytes (chainID 32 + nodeID 20).
#
# Fixed section: 32 (ChainID) + 20 (NodeID) = 52
struct ChainIDNodeID
    shape_kind = 0x72       # ShapeKindPStateChainIDNodeID
{
    ChainID id32                     @0
    NodeID  id20                     @32
}

# ----------------------------------------------------------------------
# ValidatorMetadata — v1+ validator on-chain stats record
# ----------------------------------------------------------------------
#
# Stored per (txID -> validatorMetadata) under the current-validator
# linkedDB. Holds uptime + reward accumulators + the staker-start
# timestamp.
#
# Parse path tolerates four historical sizes (parseValidatorMetadata):
#   0                                  — nothing stored (legacy nil)
#   8 (Uint64Size)                     — PotentialReward only
#   preDelegateeRewardSize             — UpDuration + LastUpdated + PotentialReward
#   preStakerStartTimeSize             — adds PotentialDelegateeReward
#   default (v1+)                      — adds StakerStartTime
#
# The struct below is the v1+ wire layout. The legacy short forms are
# declared as separate structs so codegen accessors exist for both.
#
# Fixed section: 8 (UpDuration) + 8 (LastUpdated) +
# 8 (PotentialReward) + 8 (PotentialDelegateeReward) +
# 8 (StakerStartTime) = 40
struct ValidatorMetadata
    shape_kind = 0x73       # ShapeKindPStateValidatorMetadata
{
    # UpDuration is the cumulative uptime in nanoseconds (Go
    # time.Duration). LastUpdated tracks when this was last bumped.
    UpDuration               u64     @0
    LastUpdated              u64     @8

    # PotentialReward is the accrued validation reward (in nLUX) the
    # validator is eligible for at exit-time, contingent on uptime
    # meeting reward.Calculator's UptimeRequirement.
    PotentialReward          u64     @16

    # PotentialDelegateeReward is the validator's share of accrued
    # delegation rewards.
    PotentialDelegateeReward u64     @24

    # StakerStartTime is the unix-seconds time the validator entered
    # the active set. Distinct from the Validator.Start field on the
    # original AddValidatorTx (which may pre-date the actual entry
    # for pending validators).
    StakerStartTime          u64     @32
}

# ----------------------------------------------------------------------
# PreDelegateeRewardMetadata — pre-PotentialDelegateeReward layout
# ----------------------------------------------------------------------
#
# Apricot-era validator metadata. Lacks PotentialDelegateeReward and
# StakerStartTime. Decoded read-only by parseValidatorMetadata when
# the stored value has size preDelegateeRewardSize.
#
# Fixed section: 8 (UpDuration) + 8 (LastUpdated) +
# 8 (PotentialReward) = 24
struct PreDelegateeRewardMetadata
    shape_kind = 0x74       # ShapeKindPStatePreDelegateeRewardMetadata
{
    UpDuration      u64              @0
    LastUpdated     u64              @8
    PotentialReward u64              @16
}

# ----------------------------------------------------------------------
# PreStakerStartTimeMetadata — pre-StakerStartTime layout
# ----------------------------------------------------------------------
#
# Banff-era validator metadata with delegatee reward but without
# StakerStartTime. Decoded read-only by parseValidatorMetadata when
# the stored value has size preStakerStartTimeSize.
#
# Fixed section: 8 (UpDuration) + 8 (LastUpdated) +
# 8 (PotentialReward) + 8 (PotentialDelegateeReward) = 32
struct PreStakerStartTimeMetadata
    shape_kind = 0x75       # ShapeKindPStatePreStakerStartTimeMetadata
{
    UpDuration               u64     @0
    LastUpdated              u64     @8
    PotentialReward          u64     @16
    PotentialDelegateeReward u64     @24
}

# ----------------------------------------------------------------------
# DelegatorMetadata — delegator on-chain stats record
# ----------------------------------------------------------------------
#
# Stored per (txID -> delegatorMetadata) under the current-delegator
# linkedDB. Parse path also tolerates the legacy 8-byte-only form
# (PotentialReward as a bare uint64) for pre-StakerStartTime
# delegators.
#
# Fixed section: 8 (PotentialReward) + 8 (StakerStartTime) = 16
struct DelegatorMetadata
    shape_kind = 0x76       # ShapeKindPStateDelegatorMetadata
{
    # PotentialReward is the accrued delegation reward (in nLUX) the
    # delegator is eligible for at exit-time.
    PotentialReward u64              @0

    # StakerStartTime is the unix-seconds time the delegator entered
    # the active set.
    StakerStartTime u64              @8
}

# ----------------------------------------------------------------------
# ValidatorWeightDiff — per-height validator-set weight diff
# ----------------------------------------------------------------------
#
# Stored under ValidatorWeightDiffsPrefix keyed by (chainID + height +
# nodeID). Replays let ApplyValidatorWeightDiffs reconstruct a past
# validator set from a future one.
#
# Fixed section: 1 (Decrease) + 8 (Amount) + 32 (ValidationID) = 41
struct ValidatorWeightDiff
    shape_kind = 0x77       # ShapeKindPStateValidatorWeightDiff
{
    # Decrease=1 means the diff subtracts Amount from the validator's
    # weight; =0 means it adds.
    Decrease     u8                  @0

    # Amount is the absolute weight delta.
    Amount       u64                 @1

    # ValidationID preserves the originating tx/validation ID across
    # diff replays so the validator-set entry remains keyed correctly.
    ValidationID id32                @9
}

# ----------------------------------------------------------------------
# HeightRange — indexed-heights span singleton value
# ----------------------------------------------------------------------
#
# Stored at HeightsIndexedKey. Marks the (lower, upper) inclusive
# range of heights for which the validator-diff index is consistent
# and safe to use the native DB iterator over.
#
# Fixed section: 8 (LowerBound) + 8 (UpperBound) = 16
struct HeightRange
    shape_kind = 0x78       # ShapeKindPStateHeightRange
{
    LowerBound u64                   @0
    UpperBound u64                   @8
}

# ----------------------------------------------------------------------
# TxBytesAndStatus — stored tx value (bytes + acceptance status)
# ----------------------------------------------------------------------
#
# Stored under TxPrefix keyed by TxID. Carries the raw signed-tx
# bytes plus the platformvm/status.Status enum (Aborted=1, Committed=2,
# Processing=3, Unknown=4, Dropped=5 — see node/vms/platformvm/status).
#
# Fixed section: 8 (Tx) + 4 (Status) = 12
struct TxBytesAndStatus
    shape_kind = 0x79       # ShapeKindPStateTxBytesAndStatus
{
    # Tx is the full signed-tx bytes — these are the exact bytes the
    # producer wrote (byte-preserving). Decoding routes through the
    # codec version prefix on the bytes themselves.
    Tx     bytes                     @0

    # Status is the acceptance state of this tx (a uint32 enum from
    # node/vms/platformvm/status).
    Status u32                       @8
}

# ----------------------------------------------------------------------
# StateBlk — LEGACY pre-PR-#1719 block-storage envelope (read-only)
# ----------------------------------------------------------------------
#
# Stored under the legacy block KV prefix in pre-v1.14.x state DBs.
# RegisterStateBlockType registers this on block.GenesisCodec so
# parseStoredBlock can fall back to it for old DB upgrades.
#
# Modern (post-PR-#1719) block storage writes raw block bytes — this
# envelope is read-only legacy.
#
# Fixed section: 8 (Bytes) + 4 (Status) = 12
struct StateBlk
    shape_kind = 0x7A       # ShapeKindPStateBlk
{
    # Bytes is the legacy-wrapped block-bytes payload.
    Bytes  bytes                     @0

    # Status is the legacy choices.Status enum value.
    Status u32                       @8
}

# ----------------------------------------------------------------------
# NetToL1Conversion — chain-to-L1 conversion summary record
# ----------------------------------------------------------------------
#
# Stored under NetToL1ConversionPrefix keyed by chainID. Captures the
# canonical conversion ID + manager-chain handoff for a chain that
# went through ConvertNetworkToL1Tx.
#
# Fixed section: 32 (ConversionID) + 32 (ChainID) + 8 (Addr) = 72
struct NetToL1Conversion
    shape_kind = 0x7B       # ShapeKindPStateNetToL1Conversion
{
    # ConversionID is the hash-derived ID over the conversion's
    # canonical data (see warp.ChainToL1ConversionID).
    ConversionID id32                @0

    # ChainID is the manager-chain ID (where the validator-manager
    # contract lives post-conversion).
    ChainID      id32                @32

    # Addr is the validator-manager contract address on ChainID.
    Addr         bytes               @64
}
