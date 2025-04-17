package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
)

// Handler functions for SFC contract variables
// This file contains handlers for variable getters (as opposed to function methods)

func handleOwner(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(owner)))
	result, err := SfcAbi.Methods["owner"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

func handleCurrentSealedEpoch(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
	result, err := SfcAbi.Methods["currentSealedEpoch"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleLastValidatorID(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lastValidatorIDSlot)))
	result, err := SfcAbi.Methods["lastValidatorID"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleTotalStake(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
	result, err := SfcAbi.Methods["totalStake"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleTotalActiveStake(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
	result, err := SfcAbi.Methods["totalActiveStake"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleTotalSlashedStake(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSlashedStakeSlot)))
	result, err := SfcAbi.Methods["totalSlashedStake"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleTotalSupply(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)))
	result, err := SfcAbi.Methods["totalSupply"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleStakeTokenizerAddress(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeTokenizerAddressSlot)))
	result, err := SfcAbi.Methods["stakeTokenizerAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

func handleMinGasPrice(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	result, err := SfcAbi.Methods["minGasPrice"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleTreasuryAddress(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(treasuryAddressSlot)))
	result, err := SfcAbi.Methods["treasuryAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

func handleVoteBookAddress(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(voteBookAddressSlot)))
	result, err := SfcAbi.Methods["voteBookAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

func handleGetValidator(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID := args[0].(*big.Int)
	key := common.BigToHash(validatorID)
	slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(validatorSlot)).Bytes())
	val := evm.SfcStateDB.GetState(ContractAddress, slot)
	result, err := SfcAbi.Methods["getValidator"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleGetValidatorID(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	key := addr.Hash()
	slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(validatorIDSlot)).Bytes())
	val := evm.SfcStateDB.GetState(ContractAddress, slot)
	result, err := SfcAbi.Methods["getValidatorID"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleGetValidatorPubkey(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID := args[0].(*big.Int)
	key := common.BigToHash(validatorID)
	slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(validatorPubkeySlot)).Bytes())
	val := evm.SfcStateDB.GetState(ContractAddress, slot)
	result, err := SfcAbi.Methods["getValidatorPubkey"].Outputs.Pack(val.Bytes())
	return result, 0, err
}

func handleStashedRewardsUntilEpoch(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)
	key1 := addr.Hash()
	slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(stashedRewardsUntilEpochSlot)).Bytes())
	key2 := common.BigToHash(validatorID)
	slot := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())
	val := evm.SfcStateDB.GetState(ContractAddress, slot)
	result, err := SfcAbi.Methods["stashedRewardsUntilEpoch"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleGetWithdrawalRequest(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 3 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)
	wrID := args[2].(*big.Int)
	key1 := addr.Hash()
	slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(withdrawalRequestSlot)).Bytes())
	key2 := common.BigToHash(validatorID)
	slot2 := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())
	key3 := common.BigToHash(wrID)
	slot := crypto.Keccak256Hash(key3.Bytes(), slot2.Bytes())

	epoch := evm.SfcStateDB.GetState(ContractAddress, slot)
	time := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(1))))
	amount := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(2))))

	result, err := SfcAbi.Methods["getWithdrawalRequest"].Outputs.Pack(
		epoch.Big(),
		time.Big(),
		amount.Big(),
	)
	return result, 0, err
}

func handleGetStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)
	key1 := addr.Hash()
	slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(stakeSlot)).Bytes())
	key2 := common.BigToHash(validatorID)
	slot := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())
	val := evm.SfcStateDB.GetState(ContractAddress, slot)
	result, err := SfcAbi.Methods["getStake"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleGetLockupInfo(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)
	key1 := addr.Hash()
	slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(lockupInfoSlot)).Bytes())
	key2 := common.BigToHash(validatorID)
	slot := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())

	lockedStake := evm.SfcStateDB.GetState(ContractAddress, slot)
	fromEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(1))))
	endTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(2))))
	duration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(3))))

	result, err := SfcAbi.Methods["getLockupInfo"].Outputs.Pack(
		lockedStake.Big(),
		fromEpoch.Big(),
		endTime.Big(),
		duration.Big(),
	)
	return result, 0, err
}

func handleGetStashedLockupRewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)
	key1 := addr.Hash()
	slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(stashedLockupRewardsSlot)).Bytes())
	key2 := common.BigToHash(validatorID)
	slot := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())

	lockupBaseReward := evm.SfcStateDB.GetState(ContractAddress, slot)
	lockupExtraReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(1))))
	unlockedReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(2))))

	result, err := SfcAbi.Methods["getStashedLockupRewards"].Outputs.Pack(
		lockupBaseReward.Big(),
		lockupExtraReward.Big(),
		unlockedReward.Big(),
	)
	return result, 0, err
}

func handleSlashingRefundRatio(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID := args[0].(*big.Int)
	key := common.BigToHash(validatorID)
	slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(slashingRefundRatioSlot)).Bytes())
	val := evm.SfcStateDB.GetState(ContractAddress, slot)
	result, err := SfcAbi.Methods["slashingRefundRatio"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleGetEpochSnapshot(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch := args[0].(*big.Int)
	key := common.BigToHash(epoch)
	slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(epochSnapshotSlot)).Bytes())

	endTime := evm.SfcStateDB.GetState(ContractAddress, slot)
	epochFee := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(1))))
	totalBaseRewardWeight := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(2))))
	totalTxRewardWeight := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(3))))
	baseRewardPerSecond := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(4))))
	totalStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(5))))
	totalSupply := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(6))))

	result, err := SfcAbi.Methods["getEpochSnapshot"].Outputs.Pack(
		endTime.Big(),
		epochFee.Big(),
		totalBaseRewardWeight.Big(),
		totalTxRewardWeight.Big(),
		baseRewardPerSecond.Big(),
		totalStake.Big(),
		totalSupply.Big(),
	)
	return result, 0, err
}
