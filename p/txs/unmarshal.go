// Copyright (C) 2019-2026, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

// Unmarshal is the exact inverse of Marshal: the chain's wire bytes in, a
// plain struct out. Dispatch is the single kind byte at object offset 0.

import (
	"github.com/luxfi/proto/p/warp/message"
	"github.com/luxfi/zap"
)

// Unmarshal decodes an unsigned tx from the chain's wire bytes.
func Unmarshal(buf []byte) (UnsignedTx, error) {
	msg, err := zap.Parse(buf)
	if err != nil {
		return nil, err
	}
	u, err := decode(msg)
	if err != nil {
		return nil, err
	}
	u.SetBytes(buf)
	return u, nil
}

func decode(msg *zap.Message) (UnsignedTx, error) {
	o := msg.Root()
	switch k := kindOf(msg); k {
	case kindAdvanceTime:
		return &AdvanceTimeTx{Time: o.Uint64(offAdvanceTimeTime)}, nil

	case kindRewardValidator:
		return &RewardValidatorTx{TxID: readID(o, offRewardTxID)}, nil

	case kindBase:
		return &BaseTx{BaseTx: readEnvelope(o)}, nil

	case kindImport:
		return &ImportTx{
			BaseTx:         BaseTx{BaseTx: readEnvelope(o)},
			SourceChain:    readID(o, offImportSourceChain),
			ImportedInputs: readExtraIns(o, offImportInputs, offImportSigIndices),
		}, nil

	case kindExport:
		return &ExportTx{
			BaseTx:           BaseTx{BaseTx: readEnvelope(o)},
			DestinationChain: readID(o, offExportDestChain),
			ExportedOutputs:  readExtraOuts(o, offExportOutputs, offExportAddrs),
		}, nil

	case kindCreateNetwork:
		return &CreateNetworkTx{
			BaseTx:         BaseTx{BaseTx: readEnvelope(o)},
			Parent:         readID(o, offCNParent),
			Owner:          readOwner(o, offCNOwnerThreshold, offCNOwnerLocktime, offCNOwnerAddrPtr),
			Security:       readSecurity(o, offCNRestakeParent, offCNAdmission, offCNManager, offCNThreshold),
			Validators:     readNetworkValidators(o, offCNValidators, offCNValNodeIDPool, offCNValAddrPool),
			ManagerChainID: readID(o, offCNManagerChainID),
			ManagerAddress: readBytes(o, offCNManagerAddress),
		}, nil

	case kindCreateChain:
		return &CreateChainTx{
			BaseTx:            BaseTx{BaseTx: readEnvelope(o)},
			ValidateNetworkID: readID(o, offCreateChainID),
			VMID:              readID(o, offCreateChainVMID),
			BlockchainName:    o.Text(offCreateChainName),
			FxIDs:             readIDList(o, offCreateChainFxIDs),
			GenesisData:       readBytes(o, offCreateChainGenesis),
			ChainAuth:         readAuth(o, offCreateChainAuth),
		}, nil

	case kindTransferChainOwnership:
		return &TransferChainOwnershipTx{
			BaseTx:    BaseTx{BaseTx: readEnvelope(o)},
			Chain:     readID(o, offTransferChain),
			ChainAuth: readAuth(o, offTransferChainAuth),
			Owner:     readOwner(o, offTransferOwnerThreshold, offTransferOwnerLocktime, offTransferOwnerAddrs),
		}, nil

	case kindRemoveChainValidator:
		return &RemoveChainValidatorTx{
			BaseTx:    BaseTx{BaseTx: readEnvelope(o)},
			NodeID:    readNodeID(o, offRemoveNodeID),
			Chain:     readID(o, offRemoveChain),
			ChainAuth: readAuth(o, offRemoveChainAuth),
		}, nil

	case kindTransformChain:
		return &TransformChainTx{
			BaseTx:                   BaseTx{BaseTx: readEnvelope(o)},
			Chain:                    readID(o, offTransformChain),
			AssetID:                  readID(o, offTransformAssetID),
			InitialSupply:            o.Uint64(offTransformInitialSupply),
			MaximumSupply:            o.Uint64(offTransformMaximumSupply),
			MinConsumptionRate:       o.Uint64(offTransformMinConsumptionRate),
			MaxConsumptionRate:       o.Uint64(offTransformMaxConsumptionRate),
			MinValidatorStake:        o.Uint64(offTransformMinValidatorStake),
			MaxValidatorStake:        o.Uint64(offTransformMaxValidatorStake),
			MinStakeDuration:         o.Uint32(offTransformMinStakeDuration),
			MaxStakeDuration:         o.Uint32(offTransformMaxStakeDuration),
			MinDelegationFee:         o.Uint32(offTransformMinDelegationFee),
			MinDelegatorStake:        o.Uint64(offTransformMinDelegatorStake),
			MaxValidatorWeightFactor: o.Uint8(offTransformMaxValidatorWeightFactor),
			UptimeRequirement:        o.Uint32(offTransformUptimeRequirement),
			ChainAuth:                readAuth(o, offTransformChainAuth),
		}, nil

	case kindAddValidator:
		return &AddValidatorTx{
			BaseTx:           BaseTx{BaseTx: readEnvelope(o)},
			Validator:        readValidator(o, offAVValidator),
			StakeOuts:        readExtraOuts(o, offAVStakeOuts, offAVStakeAddrs),
			RewardsOwner:     readOwner(o, offAVRewardsThreshold, offAVRewardsLocktime, offAVRewardsAddrs),
			DelegationShares: o.Uint32(offAVDelegationShares),
		}, nil

	case kindAddChainValidator:
		return &AddChainValidatorTx{
			BaseTx: BaseTx{BaseTx: readEnvelope(o)},
			ChainValidator: ChainValidator{
				Validator: readValidator(o, offACVValidator),
				Chain:     readID(o, offACVChain),
			},
			ChainAuth: readAuth(o, offACVChainAuth),
		}, nil

	case kindAddDelegator:
		return &AddDelegatorTx{
			BaseTx:                 BaseTx{BaseTx: readEnvelope(o)},
			Validator:              readValidator(o, offADValidator),
			StakeOuts:              readExtraOuts(o, offADStakeOuts, offADStakeAddrs),
			DelegationRewardsOwner: readOwner(o, offADRewardsThreshold, offADRewardsLocktime, offADRewardsAddrs),
		}, nil

	case kindAddPermissionlessValidator:
		return &AddPermissionlessValidatorTx{
			BaseTx:                BaseTx{BaseTx: readEnvelope(o)},
			Validator:             readValidator(o, offAPVValidator),
			Chain:                 readID(o, offAPVChain),
			Signer:                readSigner(o, offAPVSigner),
			StakeOuts:             readExtraOuts(o, offAPVStakeOuts, offAPVStakeAddrs),
			ValidatorRewardsOwner: readOwner(o, offAPVValRewardsThreshold, offAPVValRewardsLocktime, offAPVValRewardsAddrs),
			DelegatorRewardsOwner: readOwner(o, offAPVDelRewardsThreshold, offAPVDelRewardsLocktime, offAPVDelRewardsAddrs),
			DelegationShares:      o.Uint32(offAPVDelegationShares),
		}, nil

	case kindAddPermissionlessDelegator:
		return &AddPermissionlessDelegatorTx{
			BaseTx:                 BaseTx{BaseTx: readEnvelope(o)},
			Validator:              readValidator(o, offAPDValidator),
			Chain:                  readID(o, offAPDChain),
			StakeOuts:              readExtraOuts(o, offAPDStakeOuts, offAPDStakeAddrs),
			DelegationRewardsOwner: readOwner(o, offAPDRewardsThreshold, offAPDRewardsLocktime, offAPDRewardsAddrs),
		}, nil

	case kindRegisterL1Validator:
		tx := &RegisterL1ValidatorTx{
			BaseTx:  BaseTx{BaseTx: readEnvelope(o)},
			Balance: o.Uint64(offRegisterBalance),
			Message: readBytes(o, offRegisterMessage),
		}
		copy(tx.ProofOfPossession[:], o.BytesFixedSlice(offRegisterPoP, blsSigLen))
		return tx, nil

	case kindSetL1ValidatorWeight:
		return &SetL1ValidatorWeightTx{
			BaseTx:  BaseTx{BaseTx: readEnvelope(o)},
			Message: readBytes(o, offSetWeightMessage),
		}, nil

	case kindIncreaseL1ValidatorBalance:
		return &IncreaseL1ValidatorBalanceTx{
			BaseTx:       BaseTx{BaseTx: readEnvelope(o)},
			ValidationID: readID(o, offIncreaseValidationID),
			Balance:      o.Uint64(offIncreaseBalance),
		}, nil

	case kindDisableL1Validator:
		return &DisableL1ValidatorTx{
			BaseTx:       BaseTx{BaseTx: readEnvelope(o)},
			ValidationID: readID(o, offDisableValidationID),
			DisableAuth:  readAuth(o, offDisableAuth),
		}, nil

	case kindConvertNetwork:
		return &ConvertNetworkToL1Tx{
			BaseTx:         BaseTx{BaseTx: readEnvelope(o)},
			Chain:          readID(o, offCVNetwork),
			Parent:         readID(o, offCVParent),
			ManagerChainID: readID(o, offCVManagerChainID),
			Address:        readBytes(o, offCVManagerAddress),
			Validators:     readNetworkValidators(o, offCVValidators, offCVValNodeIDPool, offCVValAddrPool),
			ChainAuth:      readAuth(o, offCVAuthPtr),
			Security:       readSecurity(o, offCVRestakeParent, offCVAdmission, offCVManager, offCVThreshold),
		}, nil

	default:
		return nil, errUnknownKind(k)
	}
}

// readBytes copies a variable-length bytes field; nil when empty.
func readBytes(o zap.Object, off int) []byte {
	if v := o.Bytes(off); len(v) > 0 {
		return append([]byte(nil), v...)
	}
	return nil
}

func readNetworkValidators(obj zap.Object, listOff, nodeIDPoolOff, addrPoolOff int) []*NetworkValidator {
	list := obj.ListStride(listOff, nvStride)
	n := list.Len()
	if n == 0 {
		return nil
	}
	nodeIDBlob := obj.Bytes(nodeIDPoolOff)
	addrs := obj.ListStride(addrPoolOff, addrStride)
	out := make([]*NetworkValidator, n)
	for i := 0; i < n; i++ {
		e := list.Object(i, nvStride)
		v := &NetworkValidator{Weight: e.Uint64(nvWeight), Balance: e.Uint64(nvBalance)}
		copy(v.Signer.PublicKey[:], e.BytesFixedSlice(nvSignerPub, blsPubLen))
		copy(v.Signer.ProofOfPossession[:], e.BytesFixedSlice(nvSignerPoP, blsSigLen))
		if ns, nl := e.Uint32(nvNodeIDStart), e.Uint32(nvNodeIDLen); nl > 0 && int(ns)+int(nl) <= len(nodeIDBlob) {
			v.NodeID = append([]byte(nil), nodeIDBlob[ns:ns+nl]...)
		}
		v.RemainingBalanceOwner = message.PChainOwner{
			Threshold: e.Uint32(nvRemThreshold),
			Addresses: sliceAddrs(addrs, e.Uint32(nvRemAddrStart), e.Uint32(nvRemAddrCount)),
		}
		v.DeactivationOwner = message.PChainOwner{
			Threshold: e.Uint32(nvDeacThreshold),
			Addresses: sliceAddrs(addrs, e.Uint32(nvDeacAddrStart), e.Uint32(nvDeacAddrCount)),
		}
		out[i] = v
	}
	return out
}
