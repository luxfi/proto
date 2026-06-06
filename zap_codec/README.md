# proto/zap_codec — ZAP-native wire codec for proto/{p,x,block,warp}

This package is the **single canonical construction site** for the wire
codecs consumed by `proto/p`, `proto/x`, `proto/p/warp`,
`proto/p/warp/payload`, `proto/p/warp/message`, `proto/p/block`, and
`proto/x/block`.

It exists so SDK consumers (`sdk/wallet/chain/{p,x}`) can wire up the
parser dependencies of the proto packages **without importing
`github.com/luxfi/codec` at the wallet layer**. The codec backend
choice is bound EXACTLY ONCE — here.

## Wire-format verdict — BREAK from legacy linearcodec

This codec emits **little-endian** wire bytes via
`github.com/luxfi/codec/zapcodec`. The previous linearcodec-backed
construction emitted **big-endian** wire bytes.

For every multi-byte field — `uint16`/`uint32`/`uint64`, length
prefixes, interface type-id discriminators, codec version bytes, string
length prefixes — the byte order is **inverted** vs the pre-Wave-2G
wire layout.

This is a **forward-only** wire-format change aligned with LP-023
ZAP-native activation (`proto/zap_native/codec_select.go`:
`ZAPActivationUnix=0`). There is no dual-mode emission, no fallback,
no rollback path. Network validators expecting LE wire bytes will
accept this codec's output; validators still on BE linearcodec will not.

### Canonical hex dump comparison

Test struct:

```go
type dummyTx struct {
    NetworkID uint32 `serialize:"true"`  // 0x01020304
    Nonce     uint64 `serialize:"true"`  // 0x0102030405060708
    Memo      []byte `serialize:"true"`  // "hi"
}
```

Marshalled at codec version 0:

```
ZAP-native (this codec):
  00 00                                          ← codec version 0 (LE uint16)
  04 03 02 01                                    ← NetworkID 0x01020304 (LE uint32)
  08 07 06 05 04 03 02 01                        ← Nonce 0x0102030405060708 (LE uint64)
  02 00 00 00                                    ← Memo length 2 (LE uint32)
  68 69                                          ← Memo "hi"

linearcodec (legacy, gone after Wave 2G):
  00 00                                          ← codec version 0 (BE uint16, same byte pattern at v=0)
  01 02 03 04                                    ← NetworkID 0x01020304 (BE uint32)
  01 02 03 04 05 06 07 08                        ← Nonce 0x0102030405060708 (BE uint64)
  00 00 00 02                                    ← Memo length 2 (BE uint32)
  68 69                                          ← Memo "hi" (byte body identical)
```

Both encodings carry the same fields in the same order — the delta is
**byte order alone**.

The above ZAP-native dump is captured live by
`TestManager_WireFormat_ZAPNative_HexDump`. To reproduce:

```
$ go test -v ./zap_codec/... -run TestManager_WireFormat_ZAPNative_HexDump
ZAP-native canonical dummyTx hex dump:
  0000040302010807060504030201020000006869
```

The legacy linearcodec bytes shown above are anchored against the
linearcodec spec (`luxfi/codec/linearcodec`); they are historical and
no longer produced by any in-tree code path.

## Architecture

Following Rich Hickey decomplection:

- **Value, not place** — the wire codec choice is a *value* qualified
  by namespace (`zap_codec.NewVersionedManager`). It is bound in ONE
  place, not braided into every consumer's package init.
- **Composition over inheritance** — `Manager` wraps `zapcodec.Codec`
  with a thin version-prefix outer layer. The outer carries the codec
  version (uint16 LE, matching the inner byte order). The two layers
  are independently complete primitives.
- **Orthogonal** — this package neither imports `proto/{p,x}` nor knows
  about block/tx/warp types. Per-chain type registration happens at the
  caller site (`sdk/wallet/chain/{p/pcodecs.go,x/constants.go,
  x/builder/constants.go}`) which `RegisterTypes` onto the registries
  returned here.
- **One canonical way** — there is exactly ONE construction expression
  per chain primitive (`NewPVMRuntime`, `NewPVMGenesis`, `NewXVMParser`,
  `NewWarp`, `NewPayload`, `NewMessage`). Wallet, builder, signer,
  genesis builder all reach for the same constructor.

## Surface

| Constructor                                | Returns                            | Budget         |
|--------------------------------------------|------------------------------------|----------------|
| `NewVersionedManager(v, maxSize)`          | `*Manager`                         | caller-supplied|
| `NewPVMRuntime(v)`                         | `*Manager`                         | `MaxSize` (1 MiB) |
| `NewPVMGenesis(v)`                         | `*Manager`                         | `math.MaxInt32`|
| `NewXVMParser(v)`                          | `(*Manager runtime, *Manager genesis)` | `MaxSize` / `math.MaxInt32` |
| `NewWarp(v)`                               | `*Manager`                         | `math.MaxInt`  |
| `NewPayload(v, payloadMax)`                | `*Manager`                         | caller-supplied|
| `NewMessage(v)`                            | `*Manager`                         | `math.MaxInt`  |

Each `*Manager` satisfies the local `Codec` interface from every
proto/{p,x,block,warp,payload,message} package by shape
(`Marshal(version, source)`, `Unmarshal(bytes, dest)`,
`Size(version, value)`) and also acts as a `Registry` (`RegisterType`,
`SkipRegistrations`).
