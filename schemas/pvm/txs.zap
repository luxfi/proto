# Copyright (C) 2026, Lux Industries Inc. All rights reserved.
# See the file LICENSE for licensing terms.
#
# P-chain (platformvm) transaction schemas — ZAP v3 native layout.
#
# Consumed by: github.com/luxfi/proto/p/txs and downstream node
# package github.com/luxfi/node/vms/platformvm/txs. Codegen emits
# *_zap.go siblings next to the generated Go consumers.
#
# Source of truth: github.com/luxfi/proto/zap_native (hand-rolled
# zero-copy accessors at canonical offsets). The legacy
# github.com/luxfi/proto/p/txs (linearcodec accessors) is being
# phased out and decodes the same wire bytes via the codec slot map.
#
# Wire envelope convention: every fixed section begins with a 1-byte
# TxKind discriminator at offset @0 (see proto/zap_native/kind.go).
# Subsequent fields shift by +1 vs the pre-v3 schemas. TxKind=0 is
# reserved (rejected by Wrap*Tx). Every field uses
# explicit ZAP slot widths:
#
#   - uint8 / uint32 / uint64 — inlined at fixed offsets
#   - ids.ShortID (20 bytes) — inlined as bytes_fixed[20]
#   - ids.ID (32 bytes) — inlined as bytes_fixed[32]
#   - bls.PublicKey (48 bytes), bls.Signature (96 bytes) — inlined
#   - list pointer = 4-byte relOffset + 4-byte count = 8 bytes
#   - bytes pointer = 4-byte relOffset + 4-byte length = 8 bytes
#   - embedded Owner header = 20 bytes (Threshold u32 + Locktime u64 +
#     AddressList 8B)
#   - OwnerStub (single-address fast path) = 32 bytes (Threshold u32 +
#     Locktime u64 + Address ShortID 20B)
#
# Standard tx prologue (every BaseTxFull-embedded tx — all txs except
# the single-field proposals AdvanceTimeTx / RewardValidatorTx /
# RegisterL1ValidatorTx / SetL1ValidatorWeightTx /
# IncreaseL1ValidatorBalanceTx / DisableL1ValidatorTx /
# SlashValidatorTx / RemoveChainValidatorTx / TransferChainOwnershipTx):
#
#   TxKind         u8     @0
#   NetworkID      u32    @1
#   BlockchainID   id32   @5
#   OutsList       8B     @37
#   InsList        8B     @45
#   CredsList      8B     @53
#   SigIndicesArr  8B     @61
#   SigArr         8B     @69
#   Memo           8B     @77
#   (tx-specific fields start @85)
#
# TxKind values (proto/zap_native/kind.go):
#
#   0  Reserved
#   1  AdvanceTime
#   2  RewardValidator
#   3  SetL1ValidatorWeight
#   4  IncreaseL1ValidatorBalance
#   5  DisableL1Validator
#   6  Base                        (metadata-only BaseTx)
#   7  RegisterL1Validator
#   8  SlashValidator
#   9  TransferChainOwnership
#   10 RemoveChainValidator
#   11 BaseFull                    (BaseTx + Outs + Ins + Credentials)
#   12 AddPermissionlessValidator
#   13 Import
#   14 Export
#   15 CreateChain
#   16 AddValidator                (pre-Etna legacy)
#   17 AddDelegator                (pre-Etna legacy)
#   18 AddPermissionlessDelegator
#   19 AddChainValidator           (was AddSubnetValidator)
#   20 CreateNetwork               (was CreateSubnet)
#   21 TransformChain              (was TransformSubnet)
#   22 ConvertNetworkToL1          (was ConvertSubnetToL1)
#   23 CreateSovereignL1
#
# Naming note (per the "no subnet language" durable rule): every
# upstream `Subnet*` name is translated to its Lux-internal name
# (Network / Chain / L1) in struct names below. Upstream-wire-byte
# compatibility is preserved — only the symbol changes.
#
# Source mapping (proto/zap_native/*.go for canonical offsets):
#   base_tx.go                                -> BaseTx (TxKind=6)
#   base_tx_full.go                           -> BaseTxFull (TxKind=11)
#   advance_time_tx.go                        -> AdvanceTimeTx (1)
#   reward_validator_tx.go                    -> RewardValidatorTx (2)
#   set_l1_validator_weight_tx.go             -> SetL1ValidatorWeightTx (3)
#   increase_l1_validator_balance_tx.go       -> IncreaseL1ValidatorBalanceTx (4)
#   disable_l1_validator_tx.go                -> DisableL1ValidatorTx (5)
#   register_l1_validator_tx.go               -> RegisterL1ValidatorTx (7)
#   slash_validator_tx.go                     -> SlashValidatorTx (8)
#   transfer_chain_ownership_tx.go            -> TransferChainOwnershipTx (9)
#   remove_chain_validator_tx.go              -> RemoveChainValidatorTx (10)
#   add_permissionless_validator_tx.go        -> AddPermissionlessValidatorTx (12)
#   import_tx.go                              -> ImportTx (13)
#   export_tx.go                              -> ExportTx (14)
#   create_chain_tx.go                        -> CreateChainTx (15)
#   add_validator_tx.go                       -> AddValidatorTx (16)
#   add_delegator_tx.go                       -> AddDelegatorTx (17)
#   add_permissionless_delegator_tx.go        -> AddPermissionlessDelegatorTx (18)
#   add_chain_validator_tx.go                 -> AddChainValidatorTx (19)
#   create_network_tx.go                      -> CreateNetworkTx (20)
#   transform_chain_tx.go                     -> TransformChainTx (21)
#   convert_network_to_l1_tx.go               -> ConvertNetworkToL1Tx (22)
#   create_sovereign_l1_tx.go                 -> CreateSovereignL1Tx (23)

package txs

# ----------------------------------------------------------------------
# Primitive aliases
# ----------------------------------------------------------------------
type id32 = bytes_fixed[32]      # ids.ID, BlockchainID, ValidationID, AssetID, ...
type id20 = bytes_fixed[20]      # ids.NodeID, ids.ShortID
type bls48 = bytes_fixed[48]     # bls.PublicKey compressed (G1)
type sig96 = bytes_fixed[96]     # bls.Signature / signer.ProofOfPossession (G2)

# ======================================================================
# Metadata-only BaseTx (TxKind=6, SizeBaseTx=45)
# ======================================================================

# ----------------------------------------------------------------------
# BaseTx — minimal {NetworkID, BlockchainID, Memo} envelope
# ----------------------------------------------------------------------
#
# The metadata-only spend-state-less form. Used where a tx carries only
# routing info — see proto/zap_native/base_tx.go. Higher-level txs
# embed BaseTxFull (below) when they also carry Outs/Ins/Credentials.
#
# Fixed section: 1 (TxKind) + 4 (NetworkID) + 32 (BlockchainID) +
# 8 (Memo) = 45
struct BaseTx
    shape_kind = 6           # TxKindBase
{
    # TxKind discriminator. Always equals 6 for this shape.
    TxKind       u8                  @0

    # NetworkID is the Lux primary network ID (1=mainnet, 2=testnet,
    # 3=devnet, 1337=local).
    NetworkID    u32                 @1

    # BlockchainID is the 32-byte chain ID this tx belongs to.
    # Replay-protects across chains.
    BlockchainID id32                @5

    # Memo is opaque bytes up to MaxMemoSize=256. Post-Durango activation
    # this field MUST be empty.
    Memo         bytes               @37
}

# ======================================================================
# Full BaseTxFull (TxKind=11, SizeBaseTxFull=85)
# ======================================================================

# ----------------------------------------------------------------------
# BaseTxFull — spend-state envelope (Outs + Ins + Credentials + Memo)
# ----------------------------------------------------------------------
#
# Embedded inline by every higher-level tx. Carries the spent UTXOs
# (Ins), produced UTXOs (Outs), and per-input credentials.
# SigIndicesArr / SigArr are shared cross-input arrays that the
# InputList and CredentialList entries slice into.
#
# Fixed section: 1 (TxKind) + 4 (NetworkID) + 32 (BlockchainID) +
# 8 (OutsList) + 8 (InsList) + 8 (CredsList) + 8 (SigIndicesArr) +
# 8 (SigArr) + 8 (Memo) = 85
struct BaseTxFull
    shape_kind = 11          # TxKindBaseFull
{
    TxKind        u8                 @0
    NetworkID     u32                @1
    BlockchainID  id32               @5

    # OutsList: list of TransferableOutput entries (see
    # utxo/utxo.zap). Each entry is fixed-stride.
    OutsList      bytes              @37

    # InsList: list of TransferableInput entries (see utxo/utxo.zap).
    InsList       bytes              @45

    # CredsList: list of Credential entries — one per Input.
    CredsList     bytes              @53

    # SigIndicesArr: shared uint32 array of sig indices that
    # InputList entries slice into.
    SigIndicesArr bytes              @61

    # SigArr: shared signature-blob array that CredentialList entries
    # slice into.
    SigArr        bytes              @69

    Memo          bytes              @77
}

# ======================================================================
# Single-field proposals (carry only TxKind + payload)
# ======================================================================

# ----------------------------------------------------------------------
# AdvanceTimeTx (TxKind=1, Size=9)
# ----------------------------------------------------------------------
#
# Pre-Banff proposal-block tx that advances the chain timestamp. Banff+
# blocks carry the timestamp in the block header, but pre-Banff blocks
# on disk still reference this tx.
#
# Fixed section: 1 (TxKind) + 8 (Time) = 9
struct AdvanceTimeTx
    shape_kind = 1           # TxKindAdvanceTime
{
    TxKind u8                        @0

    # Time is the unix-seconds timestamp the proposal proposes to
    # advance the chain to.
    Time   u64                       @1
}

# ----------------------------------------------------------------------
# RewardValidatorTx (TxKind=2, Size=33)
# ----------------------------------------------------------------------
#
# Issued as the proposal Tx of a ProposalBlock at validator exit.
# The next-block Commit/Abort decides whether the validator receives
# the reward (Commit) or only stake back (Abort).
#
# Fixed section: 1 (TxKind) + 32 (TxID) = 33
struct RewardValidatorTx
    shape_kind = 2           # TxKindRewardValidator
{
    TxKind u8                        @0

    # TxID identifies the validator-add tx whose staker is now
    # exiting the set.
    TxID   id32                      @1
}

# ======================================================================
# L1 validator weight/balance ops (carry a Warp message)
# ======================================================================

# ----------------------------------------------------------------------
# SetL1ValidatorWeightTx (TxKind=3, Size=49)
# ----------------------------------------------------------------------
#
# Update an L1 validator's weight. Carries the ValidationID + Nonce +
# new Weight. The actual authorisation comes from an accompanying
# signed Warp message; here we record the (ValidationID, Nonce, Weight)
# triple inline for fast field access.
#
# Reserved: Nonce == MaxUint64 is removal-only (Weight MUST be 0).
#
# Fixed section: 1 (TxKind) + 32 (ValidationID) + 8 (Nonce) +
# 8 (Weight) = 49
struct SetL1ValidatorWeightTx
    shape_kind = 3           # TxKindSetL1ValidatorWeight
{
    TxKind       u8                  @0
    ValidationID id32                @1
    Nonce        u64                 @33
    Weight       u64                 @41
}

# ----------------------------------------------------------------------
# IncreaseL1ValidatorBalanceTx (TxKind=4, Size=41)
# ----------------------------------------------------------------------
#
# Top up an L1 validator's pay-as-you-go balance. Balance MUST be > 0
# (ErrZeroBalance otherwise).
#
# Fixed section: 1 (TxKind) + 32 (ValidationID) + 8 (Balance) = 41
struct IncreaseL1ValidatorBalanceTx
    shape_kind = 4           # TxKindIncreaseL1ValidatorBalance
{
    TxKind       u8                  @0
    ValidationID id32                @1
    Balance      u64                 @33
}

# ----------------------------------------------------------------------
# DisableL1ValidatorTx (TxKind=5, Size=33)
# ----------------------------------------------------------------------
#
# Graceful L1 validator exit. Auth comes from the validator's
# DeactivationOwner via the credential at the BaseTxFull level — this
# fixed section carries only ValidationID.
#
# Fixed section: 1 (TxKind) + 32 (ValidationID) = 33
struct DisableL1ValidatorTx
    shape_kind = 5           # TxKindDisableL1Validator
{
    TxKind       u8                  @0
    ValidationID id32                @1
}

# ----------------------------------------------------------------------
# RegisterL1ValidatorTx (TxKind=7, RegisterL1ValidatorTx layout)
# ----------------------------------------------------------------------
#
# Register an L1 validator on the P-chain. Carries the validator's
# BLS key + PoP + remaining-balance-owner ID and the registration
# expiry. The actual remote-issued Warp message lives outside this tx
# — only its hash anchors the on-chain ValidationID.
#
# Fixed section: 1 (TxKind) + 32 (ValidationID) + 48 (BLSPublicKey) +
# 96 (ProofOfPossession) + 8 (Expiry) +
# 32 (RemainingBalanceOwnerID) = 217
struct RegisterL1ValidatorTx
    shape_kind = 7           # TxKindRegisterL1Validator
{
    TxKind                  u8       @0

    # ValidationID = hash(RegisterL1Validator warp-message bytes).
    ValidationID            id32     @1

    BLSPublicKey            bls48    @33

    # ProofOfPossession is the 96-byte BLS PoP proving the issuer
    # holds the matching secret key.
    ProofOfPossession       sig96    @81

    # Expiry is the unix-seconds deadline after which this
    # registration request is invalid (replay-protected via the
    # state ExpiryEntry record).
    Expiry                  u64      @177

    # RemainingBalanceOwnerID is the 32-byte ID hashing the
    # PChainOwner that receives leftover balance at exit time.
    RemainingBalanceOwnerID id32     @185
}

# ----------------------------------------------------------------------
# SlashValidatorTx (TxKind=8, Size=57)
# ----------------------------------------------------------------------
#
# Slash a validator for provable equivocation. Carries the NodeID +
# Network + slash percentage. The Evidence (two conflicting signed
# messages at the same height) is carried separately via the variable
# section / accompanying credentials.
#
# Fixed section: 1 (TxKind) + 20 (NodeID) + 32 (Network) +
# 4 (SlashPercentage) = 57
struct SlashValidatorTx
    shape_kind = 8           # TxKindSlashValidator
{
    TxKind          u8               @0

    # NodeID of the validator being slashed.
    NodeID          id20             @1

    # Network is the 32-byte network ID the validator was active on
    # (the L1's network ID, or PrimaryNetworkID for primary).
    Network         id32             @21

    # SlashPercentage is the percent of stake to burn, in
    # reward.PercentDenominator units (100_000 = 10%). Capped at
    # 1_000_000 (100%).
    SlashPercentage u32              @53
}

# ----------------------------------------------------------------------
# TransferChainOwnershipTx (TxKind=9, Size=65)
# ----------------------------------------------------------------------
#
# Rotate a chain's owner via an OwnerStub. Wire op upstream-named
# TransferSubnetOwnershipTx.
#
# Fixed section: 1 (TxKind) + 32 (Chain) + 4 (OwnerThreshold) +
# 8 (OwnerLocktime) + 20 (OwnerAddress) = 65
struct TransferChainOwnershipTx
    shape_kind = 9           # TxKindTransferChainOwnership
{
    TxKind         u8                @0

    # Chain is the chain whose owner is being rotated. MUST NOT be a
    # permissionless chain.
    Chain          id32              @1

    # OwnerStub: single-address Owner fast path (Threshold + Locktime
    # + Address).
    OwnerThreshold u32               @33
    OwnerLocktime  u64               @37
    OwnerAddress   id20              @45
}

# ----------------------------------------------------------------------
# RemoveChainValidatorTx (TxKind=10, Size=53)
# ----------------------------------------------------------------------
#
# Remove a validator from a permissioned chain. Wire op upstream-named
# RemoveSubnetValidatorTx.
#
# Fixed section: 1 (TxKind) + 20 (NodeID) + 32 (Network) = 53
struct RemoveChainValidatorTx
    shape_kind = 10          # TxKindRemoveChainValidator
{
    TxKind  u8                       @0

    # NodeID of the validator to remove.
    NodeID  id20                     @1

    # Network is the 32-byte chain ID this validator is being
    # removed from. MUST NOT be PrimaryNetworkID.
    Network id32                     @21
}

# ======================================================================
# BaseTxFull-embedded txs (every field below @85 is tx-specific)
# ======================================================================

# ----------------------------------------------------------------------
# AddPermissionlessValidatorTx (TxKind=12, Size=381)
# ----------------------------------------------------------------------
#
# Banff+ permissionless validator add. Used for both primary-network
# validators (Chain==PrimaryNetworkID, BLS PoP required) and L1
# validators (Chain!=Primary, empty signer).
#
# Carries two OwnerStub-style rewards owners — single-address fast
# path. Multi-address callers flow through the legacy linearcodec
# accessor in proto/p/txs/.
#
# Fixed section continues from BaseTxFull (offsets 0..84):
#   NodeID                       @85   id20
#   StakeStart                   @105  u64
#   StakeEnd                     @113  u64
#   StakeWeight                  @121  u64
#   StakeAssetID                 @129  id32
#   StakeAmount                  @161  u64
#   BLSPublicKey                 @169  bls48
#   BLSProofOfPossession         @217  sig96
#   ValidationRewardsOwnerThresh @313  u32
#   ValidationRewardsOwnerLT     @317  u64
#   ValidationRewardsOwnerAddr   @325  id20
#   DelegationRewardsOwnerThresh @345  u32
#   DelegationRewardsOwnerLT     @349  u64
#   DelegationRewardsOwnerAddr   @357  id20
#   DelegationShares             @377  u32
#   Size = 381
struct AddPermissionlessValidatorTx
    shape_kind = 12          # TxKindAddPermissionlessValidator
{
    TxKind                       u8   @0
    NetworkID                    u32  @1
    BlockchainID                 id32 @5
    OutsList                     bytes @37
    InsList                      bytes @45
    CredsList                    bytes @53
    SigIndicesArr                bytes @61
    SigArr                       bytes @69
    Memo                         bytes @77

    # NodeID is the 20-byte short-id of the validator's node.
    NodeID                       id20 @85

    # StakeStart / StakeEnd: unix-seconds active-set window.
    StakeStart                   u64  @105
    StakeEnd                     u64  @113

    # StakeWeight: validator sampling weight. MUST equal the sum of
    # StakeOuts amounts at the BaseTxFull layer.
    StakeWeight                  u64  @121

    # StakeAssetID + StakeAmount: the staked asset's ID and total
    # staked quantity. AssetID equals the runtime UTXOAssetID for
    # primary-network stake; can differ on transformed L1s.
    StakeAssetID                 id32 @129
    StakeAmount                  u64  @161

    # BLSPublicKey is the 48-byte compressed BLS public key. Required
    # for primary-network validators; empty (zeroed) for L1 validators.
    BLSPublicKey                 bls48 @169

    # BLSProofOfPossession is the 96-byte BLS PoP. Required for
    # primary-network validators.
    BLSProofOfPossession         sig96 @217

    # ValidationRewardsOwner — OwnerStub for the validator's own
    # validation rewards.
    ValidationRewardsOwnerThresh u32  @313
    ValidationRewardsOwnerLT     u64  @317
    ValidationRewardsOwnerAddr   id20 @325

    # DelegationRewardsOwner — OwnerStub for the validator's share of
    # delegation rewards.
    DelegationRewardsOwnerThresh u32  @345
    DelegationRewardsOwnerLT     u64  @349
    DelegationRewardsOwnerAddr   id20 @357

    # DelegationShares: delegation fee in PercentDenominator units
    # (max 1_000_000 = 100%).
    DelegationShares             u32  @377
}

# ----------------------------------------------------------------------
# ImportTx (TxKind=13, Size=125)
# ----------------------------------------------------------------------
#
# Atomic UTXO import from SourceNetwork. The combined Ins list carries
# both local fee-paying inputs and remote (imported) inputs; the
# ImportedInsStart / ImportedInsCount header fields identify the
# imported slice within.
#
# Fixed section continues from BaseTxFull:
#   SourceNetwork      @85    id32
#   ImportedInsStart   @117   u32 (index into Ins list)
#   ImportedInsCount   @121   u32
#   Size = 125
struct ImportTx
    shape_kind = 13          # TxKindImport
{
    TxKind            u8             @0
    NetworkID         u32            @1
    BlockchainID      id32           @5
    OutsList          bytes          @37
    InsList           bytes          @45
    CredsList         bytes          @53
    SigIndicesArr     bytes          @61
    SigArr            bytes          @69
    Memo              bytes          @77

    # SourceNetwork is the 32-byte chain ID we are importing UTXOs
    # from (typically the X-chain in the X->P bootstrap funding flow).
    SourceNetwork     id32           @85

    # ImportedInsStart: index into InsList where the imported inputs
    # begin. Local fee-paying inputs occupy [0, ImportedInsStart);
    # imported inputs occupy [ImportedInsStart,
    # ImportedInsStart+ImportedInsCount).
    ImportedInsStart  u32            @117

    ImportedInsCount  u32            @121
}

# ----------------------------------------------------------------------
# ExportTx (TxKind=14, Size=125)
# ----------------------------------------------------------------------
#
# Atomic UTXO export to DestinationNetwork. Symmetric to ImportTx —
# the OutsList carries both local (change) outputs and exported
# outputs, identified by ExportedOutsStart/Count.
#
# Fixed section continues from BaseTxFull:
#   DestinationNetwork @85    id32
#   ExportedOutsStart  @117   u32
#   ExportedOutsCount  @121   u32
#   Size = 125
struct ExportTx
    shape_kind = 14          # TxKindExport
{
    TxKind             u8            @0
    NetworkID          u32           @1
    BlockchainID       id32          @5
    OutsList           bytes         @37
    InsList            bytes         @45
    CredsList          bytes         @53
    SigIndicesArr      bytes         @61
    SigArr             bytes         @69
    Memo               bytes         @77

    DestinationNetwork id32          @85
    ExportedOutsStart  u32           @117
    ExportedOutsCount  u32           @121
}

# ----------------------------------------------------------------------
# CreateChainTx (TxKind=15, Size=221)
# ----------------------------------------------------------------------
#
# Register a new chain. Carries the parent network ID + VM ID +
# OwnerStub + optional Warp message hash anchoring this CreateChain
# to a cross-network commit + variable-length genesis-data blob.
#
# Note: the on-disk layout in proto/zap_native/create_chain_tx.go
# uses ParentNetwork (the chain ID that owns this chain) instead of
# the upstream "subnetID" parameter — the rename is part of the
# "no subnet language" rule.
#
# Fixed section continues from BaseTxFull:
#   ParentNetwork   @85    id32   (was subnetID upstream)
#   VMID            @117   id32
#   OwnerThreshold  @149   u32
#   OwnerLocktime   @153   u64
#   OwnerAddress    @161   id20
#   WarpMessageHash @181   id32
#   GenesisData     @213   bytes (variable)
#   Size = 221
struct CreateChainTx
    shape_kind = 15          # TxKindCreateChain
{
    TxKind          u8               @0
    NetworkID       u32              @1
    BlockchainID    id32             @5
    OutsList        bytes            @37
    InsList         bytes            @45
    CredsList       bytes            @53
    SigIndicesArr   bytes            @61
    SigArr          bytes            @69
    Memo            bytes            @77

    # ParentNetwork is the network ID this chain is being created
    # under. MUST NOT be PrimaryNetworkID — that path is closed,
    # primary-network chains are baked in at genesis.
    ParentNetwork   id32             @85

    # VMID is the 32-byte VM ID (EVM, DEX, FHE, ...).
    VMID            id32             @117

    # OwnerStub: single-address fast path.
    OwnerThreshold  u32              @149
    OwnerLocktime   u64              @153
    OwnerAddress    id20             @161

    # WarpMessageHash is the SHA-256 hash of the originating Warp
    # message when this CreateChain was authorised via a cross-network
    # commit. Zero ids.ID when locally authorised.
    WarpMessageHash id32             @181

    # GenesisData is the new chain's genesis-state blob.
    GenesisData     bytes            @213
}

# ----------------------------------------------------------------------
# AddValidatorTx (TxKind=16, Size=173) — LEGACY pre-Etna
# ----------------------------------------------------------------------
#
# Pre-Etna primary-network validator add. Mostly superseded by
# AddPermissionlessValidatorTx but still parsed for historical blocks.
# Multi-address rewards owners go through the legacy linearcodec
# accessor; this fixed-section layout assumes the single-address
# OwnerStub fast path.
#
# Fixed section continues from BaseTxFull:
#   NodeID                @85   id20
#   StakeStart            @105  u64
#   StakeEnd              @113  u64
#   StakeWeight           @121  u64
#   StakeOutsList         @129  8B (sibling Outs for staked tokens)
#   RewardsOwnerThreshold @137  u32
#   RewardsOwnerLocktime  @141  u64
#   RewardsOwnerAddress   @149  id20
#   DelegationShares      @169  u32
#   Size = 173
struct AddValidatorTx
    shape_kind = 16          # TxKindAddValidator
{
    TxKind                u8         @0
    NetworkID             u32        @1
    BlockchainID          id32       @5
    OutsList              bytes      @37
    InsList               bytes      @45
    CredsList             bytes      @53
    SigIndicesArr         bytes      @61
    SigArr                bytes      @69
    Memo                  bytes      @77

    NodeID                id20       @85
    StakeStart            u64        @105
    StakeEnd              u64        @113
    StakeWeight           u64        @121

    # StakeOutsList: sibling Outs list specifically carrying the
    # staked outputs (distinct from the BaseTxFull OutsList which is
    # for fee-change).
    StakeOutsList         bytes      @129

    RewardsOwnerThreshold u32        @137
    RewardsOwnerLocktime  u64        @141
    RewardsOwnerAddress   id20       @149
    DelegationShares      u32        @169
}

# ----------------------------------------------------------------------
# AddDelegatorTx (TxKind=17, Size=169) — LEGACY pre-Etna
# ----------------------------------------------------------------------
#
# Pre-Etna delegator add. Superseded by AddPermissionlessDelegatorTx
# but still parsed for historical blocks.
#
# Fixed section continues from BaseTxFull:
#   NodeID                          @85   id20
#   StakeStart                      @105  u64
#   StakeEnd                        @113  u64
#   StakeWeight                     @121  u64
#   StakeOutsList                   @129  8B
#   DelegationRewardsThreshold      @137  u32
#   DelegationRewardsLocktime       @141  u64
#   DelegationRewardsAddress        @149  id20
#   Size = 169
struct AddDelegatorTx
    shape_kind = 17          # TxKindAddDelegator
{
    TxKind                     u8    @0
    NetworkID                  u32   @1
    BlockchainID               id32  @5
    OutsList                   bytes @37
    InsList                    bytes @45
    CredsList                  bytes @53
    SigIndicesArr              bytes @61
    SigArr                     bytes @69
    Memo                       bytes @77

    NodeID                     id20  @85
    StakeStart                 u64   @105
    StakeEnd                   u64   @113
    StakeWeight                u64   @121
    StakeOutsList              bytes @129
    DelegationRewardsThreshold u32   @137
    DelegationRewardsLocktime  u64   @141
    DelegationRewardsAddress   id20  @149
}

# ----------------------------------------------------------------------
# AddPermissionlessDelegatorTx (TxKind=18, Size=233)
# ----------------------------------------------------------------------
#
# Etna+ permissionless delegator. Stakes against an
# AddPermissionlessValidator-registered validator. Chain identifies
# the validator's chain (PrimaryNetworkID for primary).
#
# Fixed section continues from BaseTxFull:
#   NodeID                          @85   id20
#   StakeStart                      @105  u64
#   StakeEnd                        @113  u64
#   StakeWeight                     @121  u64
#   Chain                           @129  id32
#   StakeAssetID                    @161  id32
#   StakeOutsList                   @193  8B
#   DelegationRewardsThreshold      @201  u32
#   DelegationRewardsLocktime       @205  u64
#   DelegationRewardsAddress        @213  id20
#   Size = 233
struct AddPermissionlessDelegatorTx
    shape_kind = 18          # TxKindAddPermissionlessDelegator
{
    TxKind                     u8    @0
    NetworkID                  u32   @1
    BlockchainID               id32  @5
    OutsList                   bytes @37
    InsList                    bytes @45
    CredsList                  bytes @53
    SigIndicesArr              bytes @61
    SigArr                     bytes @69
    Memo                       bytes @77

    NodeID                     id20  @85
    StakeStart                 u64   @105
    StakeEnd                   u64   @113
    StakeWeight                u64   @121

    # Chain is the chain ID being delegated on. PrimaryNetworkID for
    # primary-network delegators; L1 chain ID otherwise.
    Chain                      id32  @129

    # StakeAssetID is the asset ID being staked.
    StakeAssetID               id32  @161

    StakeOutsList              bytes @193

    DelegationRewardsThreshold u32   @201
    DelegationRewardsLocktime  u64   @205
    DelegationRewardsAddress   id20  @213
}

# ----------------------------------------------------------------------
# AddChainValidatorTx (TxKind=19, Size=161)
# ----------------------------------------------------------------------
#
# Add a permissioned per-chain validator. Wire op upstream-named
# AddSubnetValidatorTx. Auth flows through the CredsList at the
# BaseTxFull layer (proves issuer-owner threshold).
#
# Fixed section continues from BaseTxFull:
#   NodeID       @85   id20
#   StakeStart   @105  u64
#   StakeEnd     @113  u64
#   StakeWeight  @121  u64
#   Chain        @129  id32
#   Size = 161
struct AddChainValidatorTx
    shape_kind = 19          # TxKindAddChainValidator
{
    TxKind        u8               @0
    NetworkID     u32              @1
    BlockchainID  id32             @5
    OutsList      bytes            @37
    InsList       bytes            @45
    CredsList     bytes            @53
    SigIndicesArr bytes            @61
    SigArr        bytes            @69
    Memo          bytes            @77

    NodeID        id20             @85
    StakeStart    u64              @105
    StakeEnd      u64              @113
    StakeWeight   u64              @121

    # Chain is the chain ID being validated. MUST NOT be
    # PrimaryNetworkID.
    Chain         id32             @129
}

# ----------------------------------------------------------------------
# CreateNetworkTx (TxKind=20, Size=117)
# ----------------------------------------------------------------------
#
# Register a new network. Wire op upstream-named CreateSubnetTx.
# Allocates a new network ID (derived from tx hash) and records the
# OwnerStub authorising future admin txs against it.
#
# Fixed section continues from BaseTxFull:
#   OwnerThreshold @85   u32
#   OwnerLocktime  @89   u64
#   OwnerAddress   @97   id20
#   Size = 117
struct CreateNetworkTx
    shape_kind = 20          # TxKindCreateNetwork
{
    TxKind         u8              @0
    NetworkID      u32             @1
    BlockchainID   id32            @5
    OutsList       bytes           @37
    InsList        bytes           @45
    CredsList      bytes           @53
    SigIndicesArr  bytes           @61
    SigArr         bytes           @69
    Memo           bytes           @77

    OwnerThreshold u32             @85
    OwnerLocktime  u64             @89
    OwnerAddress   id20            @97
}

# ----------------------------------------------------------------------
# TransformChainTx (TxKind=21, Size=222)
# ----------------------------------------------------------------------
#
# Convert a permissioned chain to permissionless. Replaces the chain's
# staking parameters with an Avalanche-style permissionless config.
# Wire op upstream-named TransformSubnetTx.
#
# Fixed section continues from BaseTxFull:
#   Chain                    @85   id32
#   AssetID                  @117  id32
#   InitialSupply            @149  u64
#   MaximumSupply            @157  u64
#   MinConsumptionRate       @165  u64
#   MaxConsumptionRate       @173  u64
#   MinValidatorStake        @181  u64
#   MaxValidatorStake        @189  u64
#   MinStakeDuration         @197  u32
#   MaxStakeDuration         @201  u32
#   MinDelegationFee         @205  u32
#   MinDelegatorStake        @209  u64
#   MaxValidatorWeightFactor @217  u8
#   UptimeRequirement        @218  u32
#   Size = 222
struct TransformChainTx
    shape_kind = 21          # TxKindTransformChain
{
    TxKind                   u8      @0
    NetworkID                u32     @1
    BlockchainID             id32    @5
    OutsList                 bytes   @37
    InsList                  bytes   @45
    CredsList                bytes   @53
    SigIndicesArr            bytes   @61
    SigArr                   bytes   @69
    Memo                     bytes   @77

    # Chain is the chain being transformed. MUST NOT be
    # PrimaryNetworkID.
    Chain                    id32    @85

    # AssetID is the per-chain native asset that becomes the
    # staking + reward asset. MUST NOT be PrimaryNetwork's UTXO
    # asset.
    AssetID                  id32    @117

    InitialSupply            u64     @149
    MaximumSupply            u64     @157
    MinConsumptionRate       u64     @165
    MaxConsumptionRate       u64     @173
    MinValidatorStake        u64     @181
    MaxValidatorStake        u64     @189

    MinStakeDuration         u32     @197
    MaxStakeDuration         u32     @201
    MinDelegationFee         u32     @205

    MinDelegatorStake        u64     @209

    # MaxValidatorWeightFactor: single-byte multiplier on max
    # delegation per validator. Value 1 effectively disables
    # delegation. MUST be > 0.
    MaxValidatorWeightFactor u8      @217

    UptimeRequirement        u32     @218
}

# ----------------------------------------------------------------------
# ConvertNetworkToL1Tx (TxKind=22, Size=165)
# ----------------------------------------------------------------------
#
# Convert a managed chain to a sovereign L1. Wire op upstream-named
# ConvertSubnetToL1Tx. Validators sub-list is a fixed-stride record per
# validator (180 bytes each) carrying NodeID + Weight + BLSPubKey +
# BLSPoP + RegistrationExpiry.
#
# Fixed section continues from BaseTxFull:
#   Chain          @85   id32
#   ManagerChainID @117  id32
#   Address        @149  bytes (variable, chain-manager contract addr)
#   ValidatorsList @157  bytes (180-byte stride records)
#   Size = 165
struct ConvertNetworkToL1Tx
    shape_kind = 22          # TxKindConvertNetworkToL1
{
    TxKind         u8                @0
    NetworkID      u32               @1
    BlockchainID   id32              @5
    OutsList       bytes             @37
    InsList        bytes             @45
    CredsList      bytes             @53
    SigIndicesArr  bytes             @61
    SigArr         bytes             @69
    Memo           bytes             @77

    # Chain is the chain being converted. MUST NOT be
    # PrimaryNetworkID; MUST NOT already be permissionless.
    Chain          id32              @85

    # ManagerChainID is the chain ID (on the now-sovereign L1)
    # hosting the validator-manager contract.
    ManagerChainID id32              @117

    # Address is the validator-manager contract address.
    # Bounded by MaxChainAddressLength=4096.
    Address        bytes             @149

    # ValidatorsList: list of L1 validator records (180-byte stride
    # each) — see proto/zap_native/validators_list.go for the per-
    # validator schema (NodeID + Weight + BLSPubKey + BLSPoP +
    # RegistrationExpiry + RemainingBalanceOwner +
    # DeactivationOwner).
    ValidatorsList bytes             @157
}

# ----------------------------------------------------------------------
# CreateSovereignL1Tx (TxKind=23, Size=193)
# ----------------------------------------------------------------------
#
# Atomic create-network + validators + chains + convert. Replaces the
# four-step CreateNetwork + AddChainValidator×N + CreateChain×K +
# ConvertNetworkToL1 flow. See proto/zap_native/create_sovereign_l1_tx.go.
#
# The variable section carries three sibling blobs holding per-chain
# Name / FxIDs / GenesisData lists — keyed by ManagerChainIdx to
# select which chain entry hosts the validator manager.
#
# Fixed section continues from BaseTxFull:
#   OwnerThreshold     @85   u32
#   OwnerLocktime      @89   u64
#   OwnerAddress       @97   id20
#   ManagerChainIdx    @117  u32
#   ManagerAddress     @121  bytes (variable)
#   ValidatorsList     @129  bytes (180-byte stride records)
#   ChainsList         @137  bytes (per-chain entries)
#   NameBlobs          @145  bytes (concatenated per-chain BlockchainName)
#   FxIDsBlobs         @153  bytes (concatenated per-chain FxIDs)
#   GenesisDataBlobs   @161  bytes (concatenated per-chain GenesisData)
#   Size = 193
struct CreateSovereignL1Tx
    shape_kind = 23          # TxKindCreateSovereignL1
{
    TxKind           u8              @0
    NetworkID        u32             @1
    BlockchainID     id32            @5
    OutsList         bytes           @37
    InsList          bytes           @45
    CredsList        bytes           @53
    SigIndicesArr    bytes           @61
    SigArr           bytes           @69
    Memo             bytes           @77

    # OwnerStub: authorises future admin txs against the L1's
    # primary-network registry record.
    OwnerThreshold   u32             @85
    OwnerLocktime    u64             @89
    OwnerAddress     id20            @97

    # ManagerChainIdx selects which entry in ChainsList[] hosts the
    # validator-manager contract.
    ManagerChainIdx  u32             @117

    # ManagerAddress is the validator-manager contract address on
    # ChainsList[ManagerChainIdx].
    ManagerAddress   bytes           @121

    # ValidatorsList: list of L1 validator records (same 180-byte
    # stride as ConvertNetworkToL1Tx).
    ValidatorsList   bytes           @129

    # ChainsList: list of per-chain fixed-stride records. Each entry
    # references its name / FxIDs / GenesisData via offsets into the
    # three sibling blob arrays.
    ChainsList       bytes           @137

    # NameBlobs concatenates every chain's BlockchainName bytes;
    # ChainsList entries slice into it.
    NameBlobs        bytes           @145

    # FxIDsBlobs concatenates every chain's FxIDs list bytes.
    FxIDsBlobs       bytes           @153

    # GenesisDataBlobs concatenates every chain's GenesisData blob.
    GenesisDataBlobs bytes           @161
}
