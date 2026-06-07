# Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
# See the file LICENSE for licensing terms.
#
# Internal codec for chainadapter/messaging conversations.
# Replaces the legacy json/v2 Serialize/DeserializeConversation pair,
# which was UTF-8 unsafe (binary tags stuffed into Go map keys).
#
# Wire shape carries the entire Conversation envelope plus its two
# nested CRDTs (membership OR-Set + read-marker LWW register set).
# Each CRDT element is serialized as a typed nested object — the
# map-key tag is now an explicit bytes_fixed[16] field rather than a
# UTF-8-encoded string.

package messaging

type id32 = bytes_fixed[32]
type tag16 = bytes_fixed[16]

# A single membership entry. The unique-tag is moved out of the map
# key into a typed field, so binary safety is no longer an issue.
struct MemberEntryWire {
    Tag               tag16  @0
    AccountID         id32   @16
    Role              text   @48      # "admin" | "member" | "viewer"
    AddedAtUnixNs     i64    @56
    AddedBy           id32   @64
    EncryptedNickname bytes  @96
    Removed           u8     @104     # 0 = active, 1 = removed
    RemovedAtUnixNs   i64    @105
}

# A single read marker (LWW register keyed by account).
struct ReadMarkerWire {
    AccountID        id32  @0
    LastReadID       id32  @32
    LastReadUnixNs   i64   @64
    UpdatedAtUnixNs  i64   @72
}

# The whole Conversation envelope.
struct ConversationWire {
    ID                    id32                     @0
    Type                  u8                       @32
    CreatedAtUnixNs       i64                      @33
    UpdatedAtUnixNs       i64                      @41
    EncryptedName         bytes                    @49
    EncryptedDescription  bytes                    @57
    EncryptedAvatar       bytes                    @65
    KeyRotationEpoch      u64                      @73
    DisappearingMessages  u8                       @81
    MessageTTL            i64                      @82
    Version               id32                     @90
    Members               list<MemberEntryWire>    @122
    ReadMarkers           list<ReadMarkerWire>     @130
    EncryptedKeys         list<MemberEntryWire>    @138   # placeholder — see notes
}
