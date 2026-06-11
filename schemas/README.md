# proto/schemas — canonical home for Lux-specific ZAP schemas

This directory is the **single canonical home** for every Lux-specific
ZAP wire schema. The language-agnostic generic ZAP capabilities + envelope
spec lives upstream at `github.com/zap-proto/zap-spec` and is
referenced (never duplicated) from here. Lux-specific shapes — Tx types,
Block types, UTXO wire envelopes, agentic-network RPC messages, cross-chain
warp messages — consolidate here.

**Out of scope for this directory:**
- **IAM / identity schemas** live with the IAM service at
  `github.com/hanzoai/iam` (the Hanzo identity stack), not in the
  blockchain proto module. There is no `lux/iam`.
- **Hanzo / Zoo** application schemas live in their respective
  orgs' proto repos.

## Layout

One subdirectory per domain. Each subdirectory holds one or more `.zap`
schema files. The package directive at the top of each `.zap` file names
the Go package that the generated `*_zap.go` siblings will live in.

```
proto/schemas/
  api/agentic.zap                  — agentic-network RPC envelope (PQ-aware)
  messaging/conversation.zap       — chainadapter messaging CRDT envelope
  pvm/block.zap                    — P-chain block envelopes (skeleton)
  pvm/state.zap                    — P-chain on-disk state-value records (skeleton)
  pvm/txs.zap                      — P-chain transaction envelopes (skeleton)
  utxo/utxo.zap                    — cross-VM UTXO + TransferableIn/Out envelopes
  warp/message.zap                 — Lux Interchain Messaging (Warp) envelopes (skeleton)
  xvm/txs.zap                      — X-chain transaction envelopes (skeleton)
```

The five legacy stub directories (`airdrop/`, `bimap/`, `chainadapter/`,
`da/`, `ips/`) are pre-existing placeholders pending authored schemas.

## Codegen convention

Each consuming Go package declares a `//go:generate zapgen <schema>` line
in its `doc.go`. The path is the absolute repo path to the schema under
this directory. Running `go generate ./...` from the consuming module
regenerates the `*_zap.go` siblings in place.

```go
// utxo/wire/doc.go
//go:generate zapgen ../../../proto/schemas/utxo/utxo.zap
```

The output `*_zap.go` lands in the same Go package as the `doc.go` that
issued the directive — never inside `proto/schemas/`. Schemas are
inputs; generated Go is colocated with the consumer.

## Wire-stability rule

Every schema is **FROZEN at v1.0** once it ships in a tagged release.
Changes are **additive only**: new fields append at higher offsets, new
shapes get fresh `ShapeKind` discriminators, new top-level structs get
fresh package-relative type names. Field removal, offset reshuffle, type
narrowing, and wire-incompatible renames are forbidden — they require a
new schema file and a new package name.

This mirrors the rule already enforced in `zap-proto/zap-spec/SPEC.md`
for the generic Capability envelope.

## Cross-reference

The generic ZAP capability + envelope spec is the language-agnostic
upstream. This directory is the Lux-specific downstream. The two are
versioned independently:

- `github.com/zap-proto/zap-spec` — generic capabilities, envelope
  framing (`[TypeKind:1][ShapeKind:1][message: N]`), signature scope
  rules. Frozen at v1.0.
- `github.com/luxfi/proto/schemas/...` — Lux-specific shapes that ride
  on top of the generic envelope. Each schema declares its own
  `ShapeKind` byte assignment in a top-of-file comment.
