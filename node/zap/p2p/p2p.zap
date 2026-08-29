# The Lux consensus and peer-discovery wire.
#
# Every message a node sends another node is one of the structs below. This
# file is the whole of what a second implementation needs; codec.go in this
# package is one implementation of it, and codec_schema_test.go holds the two
# to each other field by field.
#
#
# FRAMING
#
# A frame is one discriminator byte followed by the named struct. Two
# framings encode that struct, and they carry the same values by different
# bytes:
#
#   the stream framing, which codec.go speaks today — fields in the order
#   declared here, big-endian, u32 and u64 written in place, bytes and text
#   written as a u32 count followed by the bytes, a list written as a u32
#   count followed by its elements, and an optional struct written as one
#   presence byte followed by its fields. Offsets are not addressable: a
#   reader walks the frame from the front.
#
#   the object framing, which is ZAP proper — the @N below is a byte offset
#   into the object's fixed section, a variable field's slot holds
#   {offset u32, length u32}, and a nested struct's slot holds one offset.
#   Nothing is walked; every field is a read at a constant.
#
# The stream framing has no ZAP header and reads back-to-front is impossible,
# so a frame produced today is not a ZAP object and zap.Parse rejects it.
#
#
# DISCRIMINATORS
#
# The IDL has no variant, so the byte that selects the struct is stated here
# and enforced in codec.go:
#
#    1 CompressedZstd (a zstd frame holding one of the below)
#    2 Ping                        14 Accepted
#    3 Pong                        15 GetAncestors
#    4 Handshake                   16 Ancestors
#    5 GetPeerList                 17 Get
#    6 PeerList                    18 Put
#    7 GetStateSummaryFrontier     19 PushQuery
#    8 StateSummaryFrontier        20 PullQuery
#    9 GetAcceptedStateSummary     21 Chits
#   10 AcceptedStateSummary        22 Request
#   11 GetAcceptedFrontier         23 Response
#   12 AcceptedFrontier            24 Gossip
#   13 GetAccepted                 25 Error
#                                  26 BFT
#
# A BFT frame is ChainId followed by a second discriminator and its payload:
#
#    1 BlockProposal      4 FinalizeVote (Vote)     7 Finalization (QuorumCertificate)
#    2 Vote               5 Notarization (QC)       8 ReplicationRequest
#    3 EmptyVote          6 EmptyNotarization       9 ReplicationResponse
#
#
# EngineType is 0 unspecified, 1 chain, 2 dag.

package p2p

type id32 = bytes_fixed[32]

# --- peer discovery -------------------------------------------------------

struct ChainUptime {
    ChainId id32 @0
    Uptime  u32  @32
}

struct Ping {
    Uptime       u32               @0
    ChainUptimes list<ChainUptime> @4
}

struct Pong {
    Uptime       u32               @0
    ChainUptimes list<ChainUptime> @4
}

struct Client {
    Name  text @0
    Major u32  @8
    Minor u32  @12
    Patch u32  @16
}

struct BloomFilter {
    Filter bytes @0
    Salt   bytes @8
}

# What a node must agree with a peer about before the two run a chain
# together. RulesId is variable because a chain that declares no rule
# generation sends nothing, which is a different claim from thirty-two zeros.
struct ChainIdentity {
    NetworkId     u32   @0
    ChainId       id32  @4
    VmId          id32  @36
    GenesisDigest id32  @68
    RulesId       bytes @100
}

# IpAddr is 4 bytes for IPv4 and 16 for IPv6, so it is not fixed.
# IpMldsaSig is empty from a peer that signs classically only.
struct Handshake {
    NetworkId     u32                 @0
    MyTime        u64                 @8
    IpAddr        bytes               @16
    IpPort        u32                 @24
    IpSigningTime u64                 @32
    IpNodeIdSig   bytes               @40
    TrackedNets   list<bytes>         @48
    Client        Client              @56
    SupportedLps  list<u32>           @60
    ObjectedLps   list<u32>           @68
    KnownPeers    BloomFilter         @76
    IpBlsSig      bytes               @80
    AllChains     bool                @88
    IpMldsaSig    bytes               @92
    Chains        list<ChainIdentity> @100
}

struct GetPeerList {
    KnownPeers BloomFilter @0
    AllChains  bool        @4
}

struct ClaimedIpPort {
    X509Certificate bytes @0
    IpAddr          bytes @8
    IpPort          u32   @16
    Timestamp       u64   @24
    Signature       bytes @32
    TxId            id32  @40
}

struct PeerList {
    ClaimedIpPorts list<ClaimedIpPort> @0
}

# --- state sync -----------------------------------------------------------

struct GetStateSummaryFrontier {
    ChainId   id32 @0
    RequestId u32  @32
    Deadline  u64  @40
}

struct StateSummaryFrontier {
    ChainId   id32  @0
    RequestId u32   @32
    Summary   bytes @40
}

struct GetAcceptedStateSummary {
    ChainId   id32      @0
    RequestId u32       @32
    Deadline  u64       @40
    Heights   list<u64> @48
}

struct AcceptedStateSummary {
    ChainId    id32       @0
    RequestId  u32        @32
    SummaryIds list<id32> @40
}

# --- bootstrapping --------------------------------------------------------

struct GetAcceptedFrontier {
    ChainId    id32 @0
    RequestId  u32  @32
    Deadline   u64  @40
    EngineType u32  @48
}

struct AcceptedFrontier {
    ChainId     id32 @0
    RequestId   u32  @32
    ContainerId id32 @36
}

struct GetAccepted {
    ChainId      id32       @0
    RequestId    u32        @32
    Deadline     u64        @40
    ContainerIds list<id32> @48
    EngineType   u32        @56
}

struct Accepted {
    ChainId      id32       @0
    RequestId    u32        @32
    ContainerIds list<id32> @40
}

struct GetAncestors {
    ChainId     id32 @0
    RequestId   u32  @32
    Deadline    u64  @40
    ContainerId id32 @48
    EngineType  u32  @80
}

struct Ancestors {
    ChainId    id32        @0
    RequestId  u32         @32
    Containers list<bytes> @40
}

# --- consensus ------------------------------------------------------------

struct Get {
    ChainId     id32 @0
    RequestId   u32  @32
    Deadline    u64  @40
    ContainerId id32 @48
    EngineType  u32  @80
}

struct Put {
    ChainId    id32  @0
    RequestId  u32   @32
    Container  bytes @40
    EngineType u32   @48
}

struct PushQuery {
    ChainId         id32  @0
    RequestId       u32   @32
    Deadline        u64   @40
    Container       bytes @48
    EngineType      u32   @56
    RequestedHeight u64   @64
}

struct PullQuery {
    ChainId         id32 @0
    RequestId       u32  @32
    Deadline        u64  @40
    ContainerId     id32 @48
    EngineType      u32  @80
    RequestedHeight u64  @88
}

struct Chits {
    ChainId             id32 @0
    RequestId           u32  @32
    PreferredId         id32 @36
    PreferredIdAtHeight id32 @68
    AcceptedId          id32 @100
    AcceptedHeight      u64  @136
}

# --- application ----------------------------------------------------------

struct Request {
    ChainId   id32  @0
    RequestId u32   @32
    Deadline  u64   @40
    AppBytes  bytes @48
}

struct Response {
    ChainId   id32  @0
    RequestId u32   @32
    AppBytes  bytes @40
}

struct Gossip {
    ChainId  id32  @0
    AppBytes bytes @32
}

struct Error {
    ChainId      id32 @0
    RequestId    u32  @32
    ErrorCode    i32  @36
    ErrorMessage text @40
}

# --- BFT payloads ---------------------------------------------------------
#
# Reached through discriminator 26 and the second byte listed above. The
# hashes here are whatever the signing scheme produces, so none is fixed.

struct BlockProposal {
    Block bytes @0
}

struct Vote {
    BlockHash bytes @0
    Signature bytes @8
}

struct EmptyVote {
    View      u64   @0
    Seq       u64   @8
    Signature bytes @16
}

struct QuorumCertificate {
    BlockHash           bytes @0
    View                u64   @8
    Seq                 u64   @16
    AggregatedSignature bytes @24
    Signers             bytes @32
}

struct EmptyNotarization {
    View                u64   @0
    Seq                 u64   @8
    AggregatedSignature bytes @16
    Signers             bytes @24
}

struct ReplicationRequest {
    Seqs        list<u64> @0
    LatestRound u64       @8
}

struct ReplicationResponse {
    Messages list<bytes> @0
}
