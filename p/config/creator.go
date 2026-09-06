// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package config

import "github.com/luxfi/ids"

// ChainParameters is what the P-Chain knows about a chain it has accepted a
// CreateChainTx for: enough to ask for the chain, and nothing about how one
// gets built.
type ChainParameters struct {
	// ID is the chain being created, which is the transaction that created it.
	ID ids.ID
	// ChainID is the network the chain belongs to.
	ChainID ids.ID
	// GenesisData is passed to the VM at initialization, uninterpreted here.
	GenesisData []byte
	// VMID names the VM to run it.
	VMID ids.ID
	// FxIDs name the feature extensions the VM is given.
	FxIDs []ids.ID
	// Name is for logs and for people.
	Name string
}

// ChainCreator is asked to bring a chain into being. The P-Chain accepts the
// transaction that says a chain exists; something above it does the creating,
// and this is the whole of what the P-Chain needs to say so.
//
// It is declared here, where it is called, rather than imported from the node
// that implements it. The P-Chain is a chain the node runs, so a P-Chain that
// imports the node to describe one method makes the two modules require each
// other.
type ChainCreator interface {
	QueueChainCreation(ChainParameters)
}
