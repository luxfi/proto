// A separate module so proto's production dependency graph never gains the
// node. It exists to compare, byte for byte, what proto encodes against what
// the chain's own package encodes.
module github.com/luxfi/proto/p/txs/conformance

go 1.26.4

require (
	github.com/luxfi/crypto v1.20.4
	github.com/luxfi/ids v1.3.2
	github.com/luxfi/node v1.36.67
	github.com/luxfi/proto v1.4.4
	github.com/luxfi/utxo v0.5.8
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.5.0 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.6 // indirect
	github.com/btcsuite/btcd/chainhash/v2 v2.0.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.1.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1 // indirect
	github.com/go-json-experiment/json v0.0.0-20260601182631-00ed12fed2a6 // indirect
	github.com/gorilla/rpc v1.2.1 // indirect
	github.com/grandcat/zeroconf v1.0.0 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/luxfi/accel v1.2.4 // indirect
	github.com/luxfi/address v1.1.1 // indirect
	github.com/luxfi/api v1.1.9 // indirect
	github.com/luxfi/atomic v1.0.0 // indirect
	github.com/luxfi/cache v1.3.1 // indirect
	github.com/luxfi/compress v0.1.1 // indirect
	github.com/luxfi/concurrent v0.1.1 // indirect
	github.com/luxfi/consensus v1.36.81 // indirect
	github.com/luxfi/constants v1.6.2 // indirect
	github.com/luxfi/container v0.2.2 // indirect
	github.com/luxfi/database v1.21.5 // indirect
	github.com/luxfi/formatting v1.1.1 // indirect
	github.com/luxfi/geth v1.20.2 // indirect
	github.com/luxfi/keychain v1.1.1 // indirect
	github.com/luxfi/log v1.4.3 // indirect
	github.com/luxfi/math v1.5.1 // indirect
	github.com/luxfi/math/big v0.1.0 // indirect
	github.com/luxfi/mdns v0.1.1 // indirect
	github.com/luxfi/metric v1.8.1 // indirect
	github.com/luxfi/mock v0.1.1 // indirect
	github.com/luxfi/p2p v1.22.1 // indirect
	github.com/luxfi/pq v1.1.0 // indirect
	github.com/luxfi/runtime v1.3.1 // indirect
	github.com/luxfi/sampler v1.1.0 // indirect
	github.com/luxfi/threshold v1.12.6 // indirect
	github.com/luxfi/timer v1.1.1 // indirect
	github.com/luxfi/upgrade v1.0.3 // indirect
	github.com/luxfi/utils v1.3.1 // indirect
	github.com/luxfi/validators v1.3.1 // indirect
	github.com/luxfi/version v1.0.1 // indirect
	github.com/luxfi/vm v1.3.16 // indirect
	github.com/luxfi/warp v1.24.1 // indirect
	github.com/luxfi/zap v1.2.6 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	github.com/mr-tron/base58 v1.3.0 // indirect
	github.com/supranational/blst v0.3.16 // indirect
	go.uber.org/mock v0.6.0 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/exp v0.0.0-20260529124908-c761662dc8c9 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
	gonum.org/v1/gonum v0.17.0 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

// The module under test is this repo, not a published tag: CI must catch a
// drift in the working tree, not in a version that already shipped.
replace github.com/luxfi/proto => ../../..
