package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// checkOnlyOwner checks if the caller is the owner of the contract
// Returns nil if the caller is the owner, otherwise returns an ABI-encoded revert reason
func checkOnlyOwner(evm *vm.EVM, caller common.Address, methodName string) ([]byte, error) {
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	ownerAddr := common.BytesToAddress(owner.Bytes())
	if caller.Cmp(ownerAddr) != 0 {
		// Return ABI-encoded revert reason: "Ownable: caller is not the owner"
		revertReason := "Ownable: caller is not the owner"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkOnlyDriver checks if the caller is the NodeDriverAuth contract
// Returns nil if the caller is the NodeDriverAuth, otherwise returns an ABI-encoded revert reason
func checkOnlyDriver(evm *vm.EVM, caller common.Address, methodName string) ([]byte, error) {
	node := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(nodeDriverAuthSlot)))
	nodeAddr := common.BytesToAddress(node.Bytes())
	if caller.Cmp(nodeAddr) != 0 {
		// Return ABI-encoded revert reason: "caller is not the NodeDriverAuth contract"
		revertReason := "caller is not the NodeDriverAuth contract"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkValidatorExists checks if a validator with the given ID exists
// Returns nil if the validator exists, otherwise returns an ABI-encoded revert reason
func checkValidatorExists(evm *vm.EVM, validatorID *big.Int, methodName string) ([]byte, error) {
	// Check if validator exists
	// Use validatorSlot directly from sfc_variable_layout.go
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorSlot)))
	emptyHash := common.Hash{}
	if validatorStatus.Cmp(emptyHash) == 0 {
		// Return ABI-encoded revert reason: "validator doesn't exist"
		revertReason := "validator doesn't exist"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkValidatorNotExists checks if a validator with the given ID does not exist
// Returns nil if the validator does not exist, otherwise returns an ABI-encoded revert reason
func checkValidatorNotExists(evm *vm.EVM, validatorID *big.Int, methodName string) ([]byte, error) {
	// Check if validator doesn't exist
	// Use validatorSlot directly from sfc_variable_layout.go
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorSlot)))
	emptyHash := common.Hash{}
	if validatorStatus.Cmp(emptyHash) != 0 {
		// Return ABI-encoded revert reason: "validator already exists"
		revertReason := "validator already exists"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkValidatorActive checks if a validator is active
// Returns nil if the validator is active, otherwise returns an ABI-encoded revert reason
func checkValidatorActive(evm *vm.EVM, validatorID *big.Int, methodName string) ([]byte, error) {
	// First check if validator exists
	revertData, err := checkValidatorExists(evm, validatorID, methodName)
	if err != nil {
		return revertData, err
	}

	// Check if validator is active
	// Use validatorSlot directly from sfc_variable_layout.go
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorSlot)))
	statusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if validator is not deactivated
	if statusBigInt.Bit(0) == 1 { // WITHDRAWN_BIT
		// Return ABI-encoded revert reason: "validator is deactivated"
		revertReason := "validator is deactivated"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}

	// Check if validator is not offline
	if statusBigInt.Bit(3) == 1 { // OFFLINE_BIT
		// Return ABI-encoded revert reason: "validator is offline"
		revertReason := "validator is offline"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}

	// Check if validator is not a cheater
	if statusBigInt.Bit(7) == 1 { // DOUBLESIGN_BIT
		// Return ABI-encoded revert reason: "validator is a cheater"
		revertReason := "validator is a cheater"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}

	return nil, nil
}

// checkDelegationExists checks if a delegation exists
// Returns nil if the delegation exists, otherwise returns an ABI-encoded revert reason
func checkDelegationExists(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, methodName string) ([]byte, error) {
	// Check if delegation exists
	// Use stakeSlot directly from sfc_variable_layout.go
	delegation := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	emptyHash := common.Hash{}
	if delegation.Cmp(emptyHash) == 0 {
		// Return ABI-encoded revert reason: "delegation doesn't exist"
		revertReason := "delegation doesn't exist"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkDelegationNotExists checks if a delegation does not exist
// Returns nil if the delegation does not exist, otherwise returns an ABI-encoded revert reason
func checkDelegationNotExists(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, methodName string) ([]byte, error) {
	// Check if delegation doesn't exist
	// Use stakeSlot directly from sfc_variable_layout.go
	delegation := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	emptyHash := common.Hash{}
	if delegation.Cmp(emptyHash) != 0 {
		// Return ABI-encoded revert reason: "delegation already exists"
		revertReason := "delegation already exists"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkAlreadyInitialized checks if the contract is already initialized
// Returns nil if the contract is not initialized, otherwise returns an ABI-encoded revert reason
func checkAlreadyInitialized(evm *vm.EVM, methodName string) ([]byte, error) {
	// Check if contract is already initialized
	initializedFlag := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(isInitialized)))
	emptyHash := common.Hash{}
	if initializedFlag.Cmp(emptyHash) != 0 {
		// Return ABI-encoded revert reason: "already initialized"
		revertReason := "already initialized"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkZeroAddress checks if an address is the zero address
// Returns nil if the address is not zero, otherwise returns an ABI-encoded revert reason
func checkZeroAddress(addr common.Address, methodName string, message string) ([]byte, error) {
	emptyAddr := common.Address{}
	if addr.Cmp(emptyAddr) == 0 {
		// Return ABI-encoded revert reason with the provided message
		revertData, err := encodeRevertReason(methodName, message)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// encodeRevertReason encodes a revert reason as an ABI-encoded error
func encodeRevertReason(methodName string, reason string) ([]byte, error) {
	// Prepend the error signature: bytes4(keccak256("Error(string)"))
	errorSig := []byte{0x08, 0xc3, 0x79, 0xa0}
	// Pack the revert reason
	packedReason, err := abi.Arguments{{Type: abi.Type{T: abi.StringTy}}}.Pack(reason)
	if err != nil {
		return nil, err
	}
	// Combine the error signature and packed reason
	revertData := append(errorSig, packedReason...)
	return revertData, nil
}

// Helper functions for calculating validator storage slots

// getValidatorStatusSlot calculates the storage slot for a validator's status
func getValidatorStatusSlot(validatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return validatorSlot
}

// getValidatorCreatedEpochSlot calculates the storage slot for a validator's created epoch
func getValidatorCreatedEpochSlot(validatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return validatorSlot
}

// getValidatorCreatedTimeSlot calculates the storage slot for a validator's created time
func getValidatorCreatedTimeSlot(validatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return validatorSlot
}

// getValidatorDeactivatedEpochSlot calculates the storage slot for a validator's deactivated epoch
func getValidatorDeactivatedEpochSlot(validatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return validatorSlot
}

// getValidatorDeactivatedTimeSlot calculates the storage slot for a validator's deactivated time
func getValidatorDeactivatedTimeSlot(validatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return validatorSlot
}

// getValidatorAuthSlot calculates the storage slot for a validator's auth address
func getValidatorAuthSlot(validatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return validatorSlot
}

// getValidatorPubkeySlot calculates the storage slot for a validator's pubkey
func getValidatorPubkeySlot(validatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return validatorSlot
}

// getStakeSlot calculates the storage slot for a delegator's stake
func getStakeSlot(delegator common.Address, toValidatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return stakeSlot
}

// getValidatorReceivedStakeSlot calculates the storage slot for a validator's received stake
func getValidatorReceivedStakeSlot(validatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return validatorSlot
}

// getWithdrawalRequestSlot calculates the storage slot for a withdrawal request
func getWithdrawalRequestSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return withdrawalRequestSlot
}

// getWithdrawalRequestAmountSlot calculates the storage slot for a withdrawal request amount
func getWithdrawalRequestAmountSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return withdrawalRequestSlot
}

// getWithdrawalRequestEpochSlot calculates the storage slot for a withdrawal request epoch
func getWithdrawalRequestEpochSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return withdrawalRequestSlot
}

// getWithdrawalRequestTimeSlot calculates the storage slot for a withdrawal request time
func getWithdrawalRequestTimeSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return withdrawalRequestSlot
}

// getValidatorIDSlot calculates the storage slot for a validator ID
func getValidatorIDSlot(addr common.Address) int64 {
	// TODO: Implement proper slot calculation
	return validatorIDSlot
}

// getLockedStakeSlot calculates the storage slot for a delegation's locked stake
func getLockedStakeSlot(delegator common.Address, toValidatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return stakeSlot
}

// getLockupFromEpochSlot calculates the storage slot for a delegation's lockup from epoch
func getLockupFromEpochSlot(delegator common.Address, toValidatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return stakeSlot
}

// getLockupEndTimeSlot calculates the storage slot for a delegation's lockup end time
func getLockupEndTimeSlot(delegator common.Address, toValidatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return stakeSlot
}

// getLockupDurationSlot calculates the storage slot for a delegation's lockup duration
func getLockupDurationSlot(delegator common.Address, toValidatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return stakeSlot
}

// getEarlyWithdrawalPenaltySlot calculates the storage slot for a delegation's early withdrawal penalty
func getEarlyWithdrawalPenaltySlot(delegator common.Address, toValidatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return stakeSlot
}

// getRewardsStashSlot calculates the storage slot for a delegation's rewards stash
func getRewardsStashSlot(delegator common.Address, toValidatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return rewardsStashSlot
}

// getStashedRewardsUntilEpochSlot calculates the storage slot for a delegation's stashed rewards until epoch
func getStashedRewardsUntilEpochSlot(delegator common.Address, toValidatorID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return rewardsStashSlot
}

// callConstantManagerMethod calls a method on the ConstantManager contract and returns the result
// methodName: the name of the method to call
// args: the arguments to pass to the method
// Returns: the result of the method call, or an error if the call failed
func callConstantManagerMethod(evm *vm.EVM, methodName string, args ...interface{}) ([]interface{}, error) {
	// Get the ConstantsManager contract address
	constantsManager := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(constantsManagerSlot)))
	constantsManagerAddr := common.BytesToAddress(constantsManager.Bytes())

	// Pack the function call data
	data, err := CMAbi.Pack(methodName, args...)
	if err != nil {
		return nil, vm.ErrExecutionReverted
	}

	// Make the call to the ConstantsManager contract
	result, _, err := evm.Call(vm.AccountRef(ContractAddress), constantsManagerAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, err
	}

	// Unpack the result
	values, err := CMAbi.Methods[methodName].Outputs.Unpack(result)
	if err != nil {
		return nil, vm.ErrExecutionReverted
	}

	return values, nil
}

// getCurrentEpoch returns the current epoch value (currentSealedEpoch + 1)
// This implements the logic from the currentEpoch() function in SFCBase.sol
func getCurrentEpoch(evm *vm.EVM) (*big.Int, error) {
	// Get the current sealed epoch
	currentSealedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
	currentSealedEpochBigInt := new(big.Int).SetBytes(currentSealedEpoch.Bytes())

	// Calculate current epoch as currentSealedEpoch + 1
	currentEpochBigInt := new(big.Int).Add(currentSealedEpochBigInt, big.NewInt(1))

	return currentEpochBigInt, nil
}

// getMinSelfStake returns the minimum self-stake value from the ConstantManager contract
func getMinSelfStake(evm *vm.EVM) (*big.Int, error) {
	// Call the minSelfStake method on the ConstantManager contract
	values, err := callConstantManagerMethod(evm, "minSelfStake")
	if err != nil {
		return nil, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, vm.ErrExecutionReverted
	}

	minSelfStakeBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, vm.ErrExecutionReverted
	}

	return minSelfStakeBigInt, nil
}

// getMaxDelegatedRatio returns the maximum delegated ratio value from the ConstantManager contract
func getMaxDelegatedRatio(evm *vm.EVM) (*big.Int, error) {
	// Call the maxDelegatedRatio method on the ConstantManager contract
	values, err := callConstantManagerMethod(evm, "maxDelegatedRatio")
	if err != nil {
		return nil, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, vm.ErrExecutionReverted
	}

	maxDelegatedRatioBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, vm.ErrExecutionReverted
	}

	return maxDelegatedRatioBigInt, nil
}

// getWithdrawalPeriodEpochs returns the withdrawal period epochs value from the ConstantManager contract
func getWithdrawalPeriodEpochs(evm *vm.EVM) (*big.Int, error) {
	// Call the withdrawalPeriodEpochs method on the ConstantManager contract
	values, err := callConstantManagerMethod(evm, "withdrawalPeriodEpochs")
	if err != nil {
		return nil, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, vm.ErrExecutionReverted
	}

	withdrawalPeriodEpochsBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, vm.ErrExecutionReverted
	}

	return withdrawalPeriodEpochsBigInt, nil
}

// getMinLockupDuration returns the minimum lockup duration value from the ConstantManager contract
func getMinLockupDuration(evm *vm.EVM) (*big.Int, error) {
	// Call the minLockupDuration method on the ConstantManager contract
	values, err := callConstantManagerMethod(evm, "minLockupDuration")
	if err != nil {
		return nil, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, vm.ErrExecutionReverted
	}

	minLockupDurationBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, vm.ErrExecutionReverted
	}

	return minLockupDurationBigInt, nil
}

// getMaxLockupDuration returns the maximum lockup duration value from the ConstantManager contract
func getMaxLockupDuration(evm *vm.EVM) (*big.Int, error) {
	// Call the maxLockupDuration method on the ConstantManager contract
	values, err := callConstantManagerMethod(evm, "maxLockupDuration")
	if err != nil {
		return nil, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, vm.ErrExecutionReverted
	}

	maxLockupDurationBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, vm.ErrExecutionReverted
	}

	return maxLockupDurationBigInt, nil
}

// getWithdrawalPeriodTime returns the withdrawal period time value from the ConstantManager contract
func getWithdrawalPeriodTime(evm *vm.EVM) (*big.Int, error) {
	// Call the withdrawalPeriodTime method on the ConstantManager contract
	values, err := callConstantManagerMethod(evm, "withdrawalPeriodTime")
	if err != nil {
		return nil, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, vm.ErrExecutionReverted
	}

	withdrawalPeriodTimeBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, vm.ErrExecutionReverted
	}

	return withdrawalPeriodTimeBigInt, nil
}
