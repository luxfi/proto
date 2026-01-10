// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package api

import (
	"github.com/luxfi/codec/jsonrpc"
	"github.com/luxfi/ids"
	"github.com/luxfi/protocol/p/signer"
)

// Staker is the representation of a staker sent via RPC.
type Staker struct {
	TxID      ids.ID      `json:"txID"`
	StartTime json.Uint64 `json:"startTime"`
	EndTime   json.Uint64 `json:"endTime"`
	Weight    json.Uint64 `json:"weight,omitempty"`
	StakeAmount *json.Uint64 `json:"stakeAmount,omitempty"`
	NodeID    ids.NodeID  `json:"nodeID"`
}

// Owner is the representation of an owner sent via RPC.
type Owner struct {
	Locktime  json.Uint64 `json:"locktime"`
	Threshold json.Uint32 `json:"threshold"`
	Addresses []string    `json:"addresses"`
}

// Delegator is the representation of a delegator sent via RPC.
type Delegator struct {
	Staker
	RewardOwner     *Owner       `json:"rewardOwner,omitempty"`
	PotentialReward *json.Uint64 `json:"potentialReward,omitempty"`
}

// PermissionlessValidator is the representation of a permissionless validator sent via RPC.
type PermissionlessValidator struct {
	Staker
	// L1 Validator fields
	ValidationID          *ids.ID      `json:"validationID,omitempty"`
	RemainingBalanceOwner *Owner       `json:"remainingBalanceOwner,omitempty"`
	DeactivationOwner     *Owner       `json:"deactivationOwner,omitempty"`
	MinNonce              *json.Uint64 `json:"minNonce,omitempty"`
	Balance               *json.Uint64 `json:"balance,omitempty"`

	// Standard validator fields
	ValidationRewardOwner  *Owner                      `json:"validationRewardOwner,omitempty"`
	DelegationRewardOwner  *Owner                      `json:"delegationRewardOwner,omitempty"`
	PotentialReward        *json.Uint64                `json:"potentialReward,omitempty"`
	AccruedDelegateeReward *json.Uint64                `json:"accruedDelegateeReward,omitempty"`
	DelegationFee          json.Float32                `json:"delegationFee"`
	Uptime                 *json.Float32               `json:"uptime,omitempty"`
	Connected              *bool                       `json:"connected,omitempty"`
	Signer                 *signer.ProofOfPossession   `json:"signer,omitempty"`
	DelegatorCount         *json.Uint64                `json:"delegatorCount,omitempty"`
	DelegatorWeight        *json.Uint64                `json:"delegatorWeight,omitempty"`
	Delegators             *[]Delegator                `json:"delegators,omitempty"`
}
