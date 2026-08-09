// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

// Marshal is the one encoder for a P-chain unsigned tx: a plain struct in, the
// chain's self-delimiting zap buffer out. Build order is load-bearing — every
// child list and tail is written first, the object last, because a zap object
// can only point backwards into what is already laid down.

import (
	"fmt"

	"github.com/luxfi/ids"
	"github.com/luxfi/proto/p/security"
	"github.com/luxfi/zap"
)

// Marshal encodes an unsigned tx as the chain's wire bytes.
func Marshal(u UnsignedTx) ([]byte, error) {
	m := &marshaler{}
	if err := u.Visit(m); err != nil {
		return nil, err
	}
	return m.buf, nil
}

type marshaler struct {
	buf []byte
}

// ---- security (4 fixed fields, shared by CreateNetwork and ConvertNetwork) ----

func setSecurity(ob zap.ObjectBuilder, offRestake, offAdmission, offManager, offThreshold int, m security.Mode) {
	var restake uint8
	if m.RestakeParent {
		restake = 1
	}
	ob.SetUint8(offRestake, restake)
	ob.SetUint8(offAdmission, uint8(m.Admission))
	ob.SetUint8(offManager, uint8(m.Manager))
	ob.SetUint64(offThreshold, m.Threshold)
}

func readSecurity(o zap.Object, offRestake, offAdmission, offManager, offThreshold int) security.Mode {
	return security.Mode{
		RestakeParent: o.Uint8(offRestake) != 0,
		Admission:     security.Admission(o.Uint8(offAdmission)),
		Manager:       security.Manager(o.Uint8(offManager)),
		Threshold:     o.Uint64(offThreshold),
	}
}

// ---- network validator list (fixed stride 192) ----

const (
	nvWeight        = 0
	nvBalance       = 8
	nvSignerPub     = 16
	nvSignerPoP     = 64
	nvNodeIDStart   = 160
	nvNodeIDLen     = 164
	nvRemThreshold  = 168
	nvRemAddrStart  = 172
	nvRemAddrCount  = 176
	nvDeacThreshold = 180
	nvDeacAddrStart = 184
	nvDeacAddrCount = 188
	nvStride        = 192
)

func writeNetworkValidators(b *zap.Builder, vdrs []*NetworkValidator) (listOff, listCount int, nodeIDs []byte, addrs []ids.ShortID) {
	if len(vdrs) == 0 {
		return 0, 0, nil, nil
	}
	lb := b.StartList(nvStride)
	for _, v := range vdrs {
		var e [nvStride]byte
		putU64(e[nvWeight:], v.Weight)
		putU64(e[nvBalance:], v.Balance)
		copy(e[nvSignerPub:], v.Signer.PublicKey[:])
		copy(e[nvSignerPoP:], v.Signer.ProofOfPossession[:])
		putU32(e[nvNodeIDStart:], uint32(len(nodeIDs)))
		putU32(e[nvNodeIDLen:], uint32(len(v.NodeID)))
		nodeIDs = append(nodeIDs, v.NodeID...)
		putU32(e[nvRemThreshold:], v.RemainingBalanceOwner.Threshold)
		putU32(e[nvRemAddrStart:], uint32(len(addrs)))
		putU32(e[nvRemAddrCount:], uint32(len(v.RemainingBalanceOwner.Addresses)))
		addrs = append(addrs, v.RemainingBalanceOwner.Addresses...)
		putU32(e[nvDeacThreshold:], v.DeactivationOwner.Threshold)
		putU32(e[nvDeacAddrStart:], uint32(len(addrs)))
		putU32(e[nvDeacAddrCount:], uint32(len(v.DeactivationOwner.Addresses)))
		addrs = append(addrs, v.DeactivationOwner.Addresses...)
		lb.AddBytes(e[:])
	}
	listOff, _ = lb.Finish()
	listCount = len(vdrs) // AddBytes counts bytes, not elements
	return listOff, listCount, nodeIDs, addrs
}

func writeAddrPool(b *zap.Builder, addrs []ids.ShortID) (off, count int) {
	if len(addrs) == 0 {
		return 0, 0
	}
	lb := b.StartList(addrStride)
	for _, a := range addrs {
		lb.AddBytes(a[:])
	}
	off, _ = lb.Finish()
	return off, len(addrs)
}

// ---- proposal txs (no spending envelope) ----

const offAdvanceTimeTime = 1 // u64, after kind@0

func (m *marshaler) AdvanceTimeTx(tx *AdvanceTimeTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 16 + 9)
	ob := b.StartObject(9)
	ob.SetUint8(offKind, uint8(kindAdvanceTime))
	ob.SetUint64(offAdvanceTimeTime, tx.Time)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const offRewardTxID = 1 // 32B, after kind@0

func (m *marshaler) RewardValidatorTx(tx *RewardValidatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 16 + 33)
	ob := b.StartObject(33)
	ob.SetUint8(offKind, uint8(kindRewardValidator))
	setID(ob, offRewardTxID, tx.TxID)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

// ---- spending txs ----

func (m *marshaler) BaseTx(tx *BaseTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 256 + spendSize)
	p, err := writeSpending(b, &tx.BaseTx)
	if err != nil {
		return err
	}
	ob := b.StartObject(spendSize)
	setEnvelope(ob, kindBase, &tx.BaseTx, p)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offImportSourceChain = spendSize // 77
	offImportInputs      = 109
	offImportSigIndices  = 117
	sizeImport           = 125
)

func (m *marshaler) ImportTx(tx *ImportTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 512 + sizeImport)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	listOff, listCount, sigOff, sigCount, err := writeExtraIns(b, tx.ImportedInputs)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeImport)
	setEnvelope(ob, kindImport, &tx.BaseTx.BaseTx, p)
	setID(ob, offImportSourceChain, tx.SourceChain)
	ob.SetList(offImportInputs, listOff, listCount)
	ob.SetList(offImportSigIndices, sigOff, sigCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offExportDestChain = spendSize // 77
	offExportOutputs   = 109
	offExportAddrs     = 117
	sizeExport         = 125
)

func (m *marshaler) ExportTx(tx *ExportTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 512 + sizeExport)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	listOff, listCount, addrOff, addrCount, err := writeExtraOuts(b, tx.ExportedOutputs)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeExport)
	setEnvelope(ob, kindExport, &tx.BaseTx.BaseTx, p)
	setID(ob, offExportDestChain, tx.DestinationChain)
	ob.SetList(offExportOutputs, listOff, listCount)
	ob.SetList(offExportAddrs, addrOff, addrCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offCNParent         = spendSize       // 32B
	offCNOwnerThreshold = spendSize + 32  // u32
	offCNOwnerLocktime  = spendSize + 36  // u64
	offCNOwnerAddrPtr   = spendSize + 44  // 8B
	offCNRestakeParent  = spendSize + 52  // u8
	offCNAdmission      = spendSize + 53  // u8
	offCNManager        = spendSize + 54  // u8
	offCNThreshold      = spendSize + 55  // u64
	offCNValidators     = spendSize + 63  // 8B list
	offCNValNodeIDPool  = spendSize + 71  // 8B bytes
	offCNValAddrPool    = spendSize + 79  // 8B list
	offCNManagerChainID = spendSize + 87  // 32B
	offCNManagerAddress = spendSize + 119 // 8B bytes
	sizeCreateNetwork   = spendSize + 127
)

func (m *marshaler) CreateNetworkTx(tx *CreateNetworkTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 1024 + sizeCreateNetwork + len(tx.Validators)*nvStride)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	oThreshold, oLocktime, oAddrOff, oAddrCount, err := writeOwner(b, tx.Owner)
	if err != nil {
		return err
	}
	vdrOff, vdrCount, nodeIDPool, addrPool := writeNetworkValidators(b, tx.Validators)
	valAddrOff, valAddrCount := writeAddrPool(b, addrPool)

	ob := b.StartObject(sizeCreateNetwork)
	setEnvelope(ob, kindCreateNetwork, &tx.BaseTx.BaseTx, p)
	setID(ob, offCNParent, tx.Parent)
	setOwner(ob, offCNOwnerThreshold, offCNOwnerLocktime, offCNOwnerAddrPtr, oThreshold, oLocktime, oAddrOff, oAddrCount)
	setSecurity(ob, offCNRestakeParent, offCNAdmission, offCNManager, offCNThreshold, tx.Security)
	ob.SetList(offCNValidators, vdrOff, vdrCount)
	ob.SetBytes(offCNValNodeIDPool, nodeIDPool)
	ob.SetList(offCNValAddrPool, valAddrOff, valAddrCount)
	setID(ob, offCNManagerChainID, tx.ManagerChainID)
	ob.SetBytes(offCNManagerAddress, tx.ManagerAddress)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offCreateChainID      = spendSize // 77
	offCreateChainVMID    = 109
	offCreateChainName    = 141
	offCreateChainFxIDs   = 149
	offCreateChainGenesis = 157
	offCreateChainAuth    = 165
	sizeCreateChain       = 173
)

func (m *marshaler) CreateChainTx(tx *CreateChainTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 512 + sizeCreateChain + len(tx.GenesisData))
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	fxOff, fxCount := writeIDList(b, tx.FxIDs)
	authOff, authCount, err := writeAuth(b, tx.ChainAuth)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeCreateChain)
	setEnvelope(ob, kindCreateChain, &tx.BaseTx.BaseTx, p)
	setID(ob, offCreateChainID, tx.ValidateNetworkID)
	setID(ob, offCreateChainVMID, tx.VMID)
	ob.SetText(offCreateChainName, tx.BlockchainName)
	ob.SetList(offCreateChainFxIDs, fxOff, fxCount)
	ob.SetBytes(offCreateChainGenesis, tx.GenesisData)
	ob.SetList(offCreateChainAuth, authOff, authCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offTransferChain           = spendSize // 77
	offTransferChainAuth       = 109
	offTransferOwnerThreshold  = 117
	offTransferOwnerLocktime   = 121
	offTransferOwnerAddrs      = 129
	sizeTransferChainOwnership = 137
)

func (m *marshaler) TransferChainOwnershipTx(tx *TransferChainOwnershipTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 512 + sizeTransferChainOwnership)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	authOff, authCount, err := writeAuth(b, tx.ChainAuth)
	if err != nil {
		return err
	}
	threshold, locktime, addrOff, addrCount, err := writeOwner(b, tx.Owner)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeTransferChainOwnership)
	setEnvelope(ob, kindTransferChainOwnership, &tx.BaseTx.BaseTx, p)
	setID(ob, offTransferChain, tx.Chain)
	ob.SetList(offTransferChainAuth, authOff, authCount)
	setOwner(ob, offTransferOwnerThreshold, offTransferOwnerLocktime, offTransferOwnerAddrs, threshold, locktime, addrOff, addrCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offRemoveNodeID    = spendSize // 20B
	offRemoveChain     = 97        // 32B
	offRemoveChainAuth = 129       // list ptr (8B)
	sizeRemove         = 137
)

func (m *marshaler) RemoveChainValidatorTx(tx *RemoveChainValidatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 256 + sizeRemove)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	authOff, authCount, err := writeAuth(b, tx.ChainAuth)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeRemove)
	setEnvelope(ob, kindRemoveChainValidator, &tx.BaseTx.BaseTx, p)
	setNodeID(ob, offRemoveNodeID, tx.NodeID)
	setID(ob, offRemoveChain, tx.Chain)
	ob.SetList(offRemoveChainAuth, authOff, authCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offTransformChain                    = spendSize // 77
	offTransformAssetID                  = 109
	offTransformInitialSupply            = 141
	offTransformMaximumSupply            = 149
	offTransformMinConsumptionRate       = 157
	offTransformMaxConsumptionRate       = 165
	offTransformMinValidatorStake        = 173
	offTransformMaxValidatorStake        = 181
	offTransformMinStakeDuration         = 189
	offTransformMaxStakeDuration         = 193
	offTransformMinDelegationFee         = 197
	offTransformMinDelegatorStake        = 201
	offTransformMaxValidatorWeightFactor = 209
	offTransformUptimeRequirement        = 210
	offTransformChainAuth                = 214
	sizeTransformChain                   = 222
)

func (m *marshaler) TransformChainTx(tx *TransformChainTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 512 + sizeTransformChain)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	authOff, authCount, err := writeAuth(b, tx.ChainAuth)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeTransformChain)
	setEnvelope(ob, kindTransformChain, &tx.BaseTx.BaseTx, p)
	setID(ob, offTransformChain, tx.Chain)
	setID(ob, offTransformAssetID, tx.AssetID)
	ob.SetUint64(offTransformInitialSupply, tx.InitialSupply)
	ob.SetUint64(offTransformMaximumSupply, tx.MaximumSupply)
	ob.SetUint64(offTransformMinConsumptionRate, tx.MinConsumptionRate)
	ob.SetUint64(offTransformMaxConsumptionRate, tx.MaxConsumptionRate)
	ob.SetUint64(offTransformMinValidatorStake, tx.MinValidatorStake)
	ob.SetUint64(offTransformMaxValidatorStake, tx.MaxValidatorStake)
	ob.SetUint32(offTransformMinStakeDuration, tx.MinStakeDuration)
	ob.SetUint32(offTransformMaxStakeDuration, tx.MaxStakeDuration)
	ob.SetUint32(offTransformMinDelegationFee, tx.MinDelegationFee)
	ob.SetUint64(offTransformMinDelegatorStake, tx.MinDelegatorStake)
	ob.SetUint8(offTransformMaxValidatorWeightFactor, tx.MaxValidatorWeightFactor)
	ob.SetUint32(offTransformUptimeRequirement, tx.UptimeRequirement)
	ob.SetList(offTransformChainAuth, authOff, authCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offAVValidator        = spendSize                      // 77
	offAVStakeOuts        = offAVValidator + validatorSize // 121
	offAVStakeAddrs       = offAVStakeOuts + 8             // 129
	offAVRewardsThreshold = offAVStakeAddrs + 8            // 137
	offAVRewardsLocktime  = offAVRewardsThreshold + 4      // 141
	offAVRewardsAddrs     = offAVRewardsLocktime + 8       // 149
	offAVDelegationShares = offAVRewardsAddrs + 8          // 157
	addValidatorSize      = offAVDelegationShares + 4      // 161
)

func (m *marshaler) AddValidatorTx(tx *AddValidatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 1024 + addValidatorSize)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	stakeListOff, stakeListCount, stakeAddrOff, stakeAddrCount, err := writeExtraOuts(b, tx.StakeOuts)
	if err != nil {
		return err
	}
	ownerThreshold, ownerLocktime, ownerAddrOff, ownerAddrCount, err := writeOwner(b, tx.RewardsOwner)
	if err != nil {
		return err
	}
	ob := b.StartObject(addValidatorSize)
	setEnvelope(ob, kindAddValidator, &tx.BaseTx.BaseTx, p)
	setValidator(ob, offAVValidator, tx.Validator)
	ob.SetList(offAVStakeOuts, stakeListOff, stakeListCount)
	ob.SetList(offAVStakeAddrs, stakeAddrOff, stakeAddrCount)
	setOwner(ob, offAVRewardsThreshold, offAVRewardsLocktime, offAVRewardsAddrs, ownerThreshold, ownerLocktime, ownerAddrOff, ownerAddrCount)
	ob.SetUint32(offAVDelegationShares, tx.DelegationShares)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offACVValidator       = spendSize                       // 77
	offACVChain           = offACVValidator + validatorSize // 121
	offACVChainAuth       = offACVChain + 32                // 153
	addChainValidatorSize = offACVChainAuth + 8             // 161
)

func (m *marshaler) AddChainValidatorTx(tx *AddChainValidatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 1024 + addChainValidatorSize)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	authOff, authCount, err := writeAuth(b, tx.ChainAuth)
	if err != nil {
		return err
	}
	ob := b.StartObject(addChainValidatorSize)
	setEnvelope(ob, kindAddChainValidator, &tx.BaseTx.BaseTx, p)
	setValidator(ob, offACVValidator, tx.ChainValidator.Validator)
	setID(ob, offACVChain, tx.ChainValidator.Chain)
	ob.SetList(offACVChainAuth, authOff, authCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offADValidator        = spendSize                      // 77
	offADStakeOuts        = offADValidator + validatorSize // 121
	offADStakeAddrs       = offADStakeOuts + 8             // 129
	offADRewardsThreshold = offADStakeAddrs + 8            // 137
	offADRewardsLocktime  = offADRewardsThreshold + 4      // 141
	offADRewardsAddrs     = offADRewardsLocktime + 8       // 149
	addDelegatorSize      = offADRewardsAddrs + 8          // 157
)

func (m *marshaler) AddDelegatorTx(tx *AddDelegatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 1024 + addDelegatorSize)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	stakeListOff, stakeListCount, stakeAddrOff, stakeAddrCount, err := writeExtraOuts(b, tx.StakeOuts)
	if err != nil {
		return err
	}
	ownerThreshold, ownerLocktime, ownerAddrOff, ownerAddrCount, err := writeOwner(b, tx.DelegationRewardsOwner)
	if err != nil {
		return err
	}
	ob := b.StartObject(addDelegatorSize)
	setEnvelope(ob, kindAddDelegator, &tx.BaseTx.BaseTx, p)
	setValidator(ob, offADValidator, tx.Validator)
	ob.SetList(offADStakeOuts, stakeListOff, stakeListCount)
	ob.SetList(offADStakeAddrs, stakeAddrOff, stakeAddrCount)
	setOwner(ob, offADRewardsThreshold, offADRewardsLocktime, offADRewardsAddrs, ownerThreshold, ownerLocktime, ownerAddrOff, ownerAddrCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offAPVValidator                = spendSize                       // 77
	offAPVChain                    = offAPVValidator + validatorSize // 121
	offAPVSigner                   = offAPVChain + 32                // 153
	offAPVStakeOuts                = offAPVSigner + signerSize       // 298
	offAPVStakeAddrs               = offAPVStakeOuts + 8             // 306
	offAPVValRewardsThreshold      = offAPVStakeAddrs + 8            // 314
	offAPVValRewardsLocktime       = offAPVValRewardsThreshold + 4   // 318
	offAPVValRewardsAddrs          = offAPVValRewardsLocktime + 8    // 326
	offAPVDelRewardsThreshold      = offAPVValRewardsAddrs + 8       // 334
	offAPVDelRewardsLocktime       = offAPVDelRewardsThreshold + 4   // 338
	offAPVDelRewardsAddrs          = offAPVDelRewardsLocktime + 8    // 346
	offAPVDelegationShares         = offAPVDelRewardsAddrs + 8       // 354
	addPermissionlessValidatorSize = offAPVDelegationShares + 4      // 358
)

func (m *marshaler) AddPermissionlessValidatorTx(tx *AddPermissionlessValidatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 1024 + addPermissionlessValidatorSize)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	stakeListOff, stakeListCount, stakeAddrOff, stakeAddrCount, err := writeExtraOuts(b, tx.StakeOuts)
	if err != nil {
		return err
	}
	valThreshold, valLocktime, valAddrOff, valAddrCount, err := writeOwner(b, tx.ValidatorRewardsOwner)
	if err != nil {
		return err
	}
	delThreshold, delLocktime, delAddrOff, delAddrCount, err := writeOwner(b, tx.DelegatorRewardsOwner)
	if err != nil {
		return err
	}
	ob := b.StartObject(addPermissionlessValidatorSize)
	setEnvelope(ob, kindAddPermissionlessValidator, &tx.BaseTx.BaseTx, p)
	setValidator(ob, offAPVValidator, tx.Validator)
	setID(ob, offAPVChain, tx.Chain)
	if err := setSigner(ob, offAPVSigner, tx.Signer); err != nil {
		return err
	}
	ob.SetList(offAPVStakeOuts, stakeListOff, stakeListCount)
	ob.SetList(offAPVStakeAddrs, stakeAddrOff, stakeAddrCount)
	setOwner(ob, offAPVValRewardsThreshold, offAPVValRewardsLocktime, offAPVValRewardsAddrs, valThreshold, valLocktime, valAddrOff, valAddrCount)
	setOwner(ob, offAPVDelRewardsThreshold, offAPVDelRewardsLocktime, offAPVDelRewardsAddrs, delThreshold, delLocktime, delAddrOff, delAddrCount)
	ob.SetUint32(offAPVDelegationShares, tx.DelegationShares)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offAPDValidator                = spendSize                       // 77
	offAPDChain                    = offAPDValidator + validatorSize // 121
	offAPDStakeOuts                = offAPDChain + 32                // 153
	offAPDStakeAddrs               = offAPDStakeOuts + 8             // 161
	offAPDRewardsThreshold         = offAPDStakeAddrs + 8            // 169
	offAPDRewardsLocktime          = offAPDRewardsThreshold + 4      // 173
	offAPDRewardsAddrs             = offAPDRewardsLocktime + 8       // 181
	addPermissionlessDelegatorSize = offAPDRewardsAddrs + 8          // 189
)

func (m *marshaler) AddPermissionlessDelegatorTx(tx *AddPermissionlessDelegatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 1024 + addPermissionlessDelegatorSize)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	stakeListOff, stakeListCount, stakeAddrOff, stakeAddrCount, err := writeExtraOuts(b, tx.StakeOuts)
	if err != nil {
		return err
	}
	ownerThreshold, ownerLocktime, ownerAddrOff, ownerAddrCount, err := writeOwner(b, tx.DelegationRewardsOwner)
	if err != nil {
		return err
	}
	ob := b.StartObject(addPermissionlessDelegatorSize)
	setEnvelope(ob, kindAddPermissionlessDelegator, &tx.BaseTx.BaseTx, p)
	setValidator(ob, offAPDValidator, tx.Validator)
	setID(ob, offAPDChain, tx.Chain)
	ob.SetList(offAPDStakeOuts, stakeListOff, stakeListCount)
	ob.SetList(offAPDStakeAddrs, stakeAddrOff, stakeAddrCount)
	setOwner(ob, offAPDRewardsThreshold, offAPDRewardsLocktime, offAPDRewardsAddrs, ownerThreshold, ownerLocktime, ownerAddrOff, ownerAddrCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offRegisterBalance      = spendSize // 77
	offRegisterPoP          = 85
	offRegisterMessage      = 181
	sizeRegisterL1Validator = 189
)

func (m *marshaler) RegisterL1ValidatorTx(tx *RegisterL1ValidatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 512 + sizeRegisterL1Validator + len(tx.Message))
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeRegisterL1Validator)
	setEnvelope(ob, kindRegisterL1Validator, &tx.BaseTx.BaseTx, p)
	ob.SetUint64(offRegisterBalance, tx.Balance)
	ob.SetBytesFixed(offRegisterPoP, tx.ProofOfPossession[:])
	ob.SetBytes(offRegisterMessage, tx.Message)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offSetWeightMessage = spendSize // bytes ptr (8B)
	sizeSetWeight       = 85
)

func (m *marshaler) SetL1ValidatorWeightTx(tx *SetL1ValidatorWeightTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 256 + sizeSetWeight + len(tx.Message))
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeSetWeight)
	setEnvelope(ob, kindSetL1ValidatorWeight, &tx.BaseTx.BaseTx, p)
	ob.SetBytes(offSetWeightMessage, tx.Message)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offIncreaseValidationID = spendSize // 32B
	offIncreaseBalance      = 109       // u64
	sizeIncrease            = 117
)

func (m *marshaler) IncreaseL1ValidatorBalanceTx(tx *IncreaseL1ValidatorBalanceTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 256 + sizeIncrease)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeIncrease)
	setEnvelope(ob, kindIncreaseL1ValidatorBalance, &tx.BaseTx.BaseTx, p)
	setID(ob, offIncreaseValidationID, tx.ValidationID)
	ob.SetUint64(offIncreaseBalance, tx.Balance)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offDisableValidationID = spendSize // 32B
	offDisableAuth         = 109       // list ptr (8B)
	sizeDisable            = 117
)

func (m *marshaler) DisableL1ValidatorTx(tx *DisableL1ValidatorTx) error {
	b := zap.NewBuilder(zap.HeaderSize + 256 + sizeDisable)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	authOff, authCount, err := writeAuth(b, tx.DisableAuth)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeDisable)
	setEnvelope(ob, kindDisableL1Validator, &tx.BaseTx.BaseTx, p)
	setID(ob, offDisableValidationID, tx.ValidationID)
	ob.SetList(offDisableAuth, authOff, authCount)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

const (
	offCVNetwork        = spendSize       // 32B
	offCVParent         = spendSize + 32  // 32B
	offCVManagerChainID = spendSize + 64  // 32B
	offCVManagerAddress = spendSize + 96  // bytes ptr (8B)
	offCVValidators     = spendSize + 104 // list ptr (8B)
	offCVValNodeIDPool  = spendSize + 112 // bytes ptr (8B)
	offCVValAddrPool    = spendSize + 120 // list ptr (8B)
	offCVAuthPtr        = spendSize + 128 // sig-idx list ptr (8B)
	offCVRestakeParent  = spendSize + 136 // u8
	offCVAdmission      = spendSize + 137 // u8
	offCVManager        = spendSize + 138 // u8
	offCVThreshold      = spendSize + 139 // u64
	sizeConvertNetwork  = spendSize + 147
)

func (m *marshaler) ConvertNetworkToL1Tx(tx *ConvertNetworkToL1Tx) error {
	b := zap.NewBuilder(zap.HeaderSize + 512 + sizeConvertNetwork + len(tx.Validators)*nvStride)
	p, err := writeSpending(b, &tx.BaseTx.BaseTx)
	if err != nil {
		return err
	}
	vdrOff, vdrCount, nodeIDPool, addrPool := writeNetworkValidators(b, tx.Validators)
	valAddrOff, valAddrCount := writeAddrPool(b, addrPool)
	authOff, authCount, err := writeAuth(b, tx.ChainAuth)
	if err != nil {
		return err
	}
	ob := b.StartObject(sizeConvertNetwork)
	setEnvelope(ob, kindConvertNetwork, &tx.BaseTx.BaseTx, p)
	setID(ob, offCVNetwork, tx.Chain)
	setID(ob, offCVParent, tx.Parent)
	setID(ob, offCVManagerChainID, tx.ManagerChainID)
	ob.SetBytes(offCVManagerAddress, tx.Address)
	ob.SetList(offCVValidators, vdrOff, vdrCount)
	ob.SetBytes(offCVValNodeIDPool, nodeIDPool)
	ob.SetList(offCVValAddrPool, valAddrOff, valAddrCount)
	ob.SetList(offCVAuthPtr, authOff, authCount)
	setSecurity(ob, offCVRestakeParent, offCVAdmission, offCVManager, offCVThreshold, tx.Security)
	ob.FinishAsRoot()
	m.buf = b.Finish()
	return nil
}

// errUnknownKind names a discriminator the wire does not define.
func errUnknownKind(k kind) error { return fmt.Errorf("zap: unknown tx kind %d", k) }
