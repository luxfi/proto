// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package rpcdb is the transport-agnostic Layer B service spec for the
// remote `database.Database` service.
//
// Layered topology:
//
//	Layer A — wire framing                   (github.com/luxfi/api/zap)
//	Layer B — service spec (this package)    (github.com/luxfi/protocol/rpcdb)
//	Layer C — service impl + transport       (github.com/luxfi/node/db/rpcdb)
//
// This package contains ONLY the data carriers (HasRequest/HasResponse
// /…) and the wire-typed Error enum. Every transport adapter — gRPC,
// ZAP, future ones — depends on this package and stays orthogonal to
// every other transport.
//
// The actual storage backend on the server side is the
// `github.com/luxfi/database`.Database interface; the canonical Service
// + transport adapters live in github.com/luxfi/node/db/rpcdb.
package rpcdb

// Error is the on-the-wire error code returned by every database RPC.
// Every transport carries it as a single byte (or proto enum, depending
// on transport) and maps it back to a `database` sentinel on the
// caller's side.
type Error int32

const (
	// Error_ERROR_UNSPECIFIED is the success / no-error sentinel.
	Error_ERROR_UNSPECIFIED Error = 0
	// Error_ERROR_CLOSED maps to database.ErrClosed.
	Error_ERROR_CLOSED Error = 1
	// Error_ERROR_NOT_FOUND maps to database.ErrNotFound.
	Error_ERROR_NOT_FOUND Error = 2
)

func (e Error) String() string {
	switch e {
	case Error_ERROR_CLOSED:
		return "CLOSED"
	case Error_ERROR_NOT_FOUND:
		return "NOT_FOUND"
	default:
		return "UNSPECIFIED"
	}
}

// HasRequest is the wire payload for Database.Has.
type HasRequest struct {
	Key []byte
}

// HasResponse is the wire reply for Database.Has.
type HasResponse struct {
	Has bool
	Err Error
}

// GetRequest is the wire payload for Database.Get.
type GetRequest struct {
	Key []byte
}

// GetResponse is the wire reply for Database.Get.
type GetResponse struct {
	Value []byte
	Err   Error
}

// PutRequest is the wire payload for Database.Put (and the PUT entries
// in WriteBatch / IteratorNext).
type PutRequest struct {
	Key   []byte
	Value []byte
}

// PutResponse is the wire reply for Database.Put.
type PutResponse struct {
	Err Error
}

// DeleteRequest is the wire payload for Database.Delete.
type DeleteRequest struct {
	Key []byte
}

// DeleteResponse is the wire reply for Database.Delete.
type DeleteResponse struct {
	Err Error
}

// WriteBatchRequest is the wire payload for Database.WriteBatch.
type WriteBatchRequest struct {
	Puts    []*PutRequest
	Deletes []*DeleteRequest
}

// WriteBatchResponse is the wire reply for Database.WriteBatch.
type WriteBatchResponse struct {
	Err Error
}

// CompactRequest is the wire payload for Database.Compact.
type CompactRequest struct {
	Start []byte
	Limit []byte
}

// CompactResponse is the wire reply for Database.Compact.
type CompactResponse struct {
	Err Error
}

// CloseRequest is the wire payload for Database.Close.
type CloseRequest struct{}

// CloseResponse is the wire reply for Database.Close.
type CloseResponse struct {
	Err Error
}

// HealthCheckResponse is the wire reply for Database.HealthCheck.
type HealthCheckResponse struct {
	Details []byte
}

// NewIteratorWithStartAndPrefixRequest is the wire payload for
// Database.NewIteratorWithStartAndPrefix.
type NewIteratorWithStartAndPrefixRequest struct {
	Start  []byte
	Prefix []byte
}

// NewIteratorWithStartAndPrefixResponse is the wire reply for
// Database.NewIteratorWithStartAndPrefix.
type NewIteratorWithStartAndPrefixResponse struct {
	Id uint64
}

// IteratorNextRequest is the wire payload for Database.IteratorNext.
type IteratorNextRequest struct {
	Id uint64
}

// IteratorNextResponse is the wire reply for Database.IteratorNext.
// Data is the next batch of (key, value) pairs to read; an empty
// batch indicates the iterator has been exhausted.
type IteratorNextResponse struct {
	Data []*PutRequest
}

// IteratorErrorRequest is the wire payload for Database.IteratorError.
type IteratorErrorRequest struct {
	Id uint64
}

// IteratorErrorResponse is the wire reply for Database.IteratorError.
type IteratorErrorResponse struct {
	Err Error
}

// IteratorReleaseRequest is the wire payload for Database.IteratorRelease.
type IteratorReleaseRequest struct {
	Id uint64
}

// IteratorReleaseResponse is the wire reply for Database.IteratorRelease.
type IteratorReleaseResponse struct {
	Err Error
}
