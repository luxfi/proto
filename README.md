# luxfi/proto

Canonical schema definitions for the Lux platform. **Module path is
`github.com/luxfi/proto`** — the directory now matches (was `protocol/`).

## Layout: ZAP + gRPC side-by-side, 1:1

Same package paths under both `zap/` and `pb/` (or in files suffixed
`_zap.go` and `_pb.go` for thin packages). The Go build picks one
or the other based on the build tag:

| Default (no tag) | `-tags grpc` |
|---|---|
| `zap/<pkg>/*_zap.go` (`//go:build !grpc`) | `pb/<pkg>/*_pb.go` (`//go:build grpc`) |
| ZAP-native types + codecs | protobuf-generated types + codecs |
| Zero `google.golang.org/grpc` in dep graph | Includes `google.golang.org/grpc` |

ZAP is the canonical wire. The protobuf path exists only for external
integrations that mandate OTLP-over-gRPC. Internal Lux services use
ZAP exclusively.

## Consolidation policy

**New schemas land here.** This directory is the single source of
truth for any cross-package wire definition. The scattered per-repo
`proto/` subdirs (see below) are legacy — they should migrate into
this module when the natural rewrite happens.

Existing scattered locations (do not add to them):

| Path | Status |
|---|---|
| `~/work/lux/p2p/proto/{zap,pb,p2p}/` | legacy — to migrate |
| `~/work/lux/node/proto/zap/` | legacy — to migrate |
| `~/work/lux/vm/proto/pb/` | legacy — to migrate |
| `~/work/lux/api/schema/proto/` | legacy — to migrate |

Migration cost is one import-path bump per consumer. Not blocking — fold
in as those packages are touched for unrelated reasons.

## Hanzo mirror

`~/work/hanzo/proto` follows the same pattern for the Hanzo platform.
See [hanzoai/proto README](file:///Users/z/work/hanzo/proto/README.md).
