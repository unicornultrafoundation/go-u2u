package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// Handler functions for SFC contract internal and private functions

// _delegate is an internal function to delegate calls to an implementation address
// This is a Go implementation of the Solidity function:
//
//	function _delegate(address implementation) internal {
//	    assembly {
//	        calldatacopy(0, 0, calldatasize())
//	        let result := delegatecall(gas(), implementation, 0, calldatasize(), 0, 0)
//	        returndatacopy(0, 0, returndatasize())
//	        switch result
//	        case 0 { revert(0, returndatasize()) }
//	        default { return(0, returndatasize()) }
//	    }
//	}
func handle_delegate(evm *vm.EVM, args []interface{}, input []byte) ([]byte, uint64, error) {
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the implementation address from args
	implementation, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// The original calldata is available in the input parameter
	// For the _delegate function, we need to skip the method selector (first 4 bytes)
	// This simulates the Solidity assembly code: calldatacopy(0, 0, calldatasize())
	originalInput := input

	// If the input starts with the _delegate method selector, skip it
	if len(originalInput) >= 4 {
		// Check if the first 4 bytes match the _delegate method selector
		if method, err := SfcAbi.MethodById(originalInput[:4]); err == nil && method.Name == "_delegate" {
			// Skip the method selector and the ABI-encoded implementation address
			// The implementation address is already extracted from args, so we don't need it in the input
			originalInput = []byte{}
		}
	}

	// Create a contract reference for the caller
	callerRef := vm.AccountRef(evm.TxContext.Origin)

	// Make the delegate call
	// This simulates the Solidity assembly code: let result := delegatecall(gas, implementation, 0, calldatasize, 0, 0)
	// Use a fixed gas amount for now
	gas := uint64(1000000)
	ret, leftOverGas, err := evm.DelegateCall(callerRef, implementation, originalInput, gas)

	// Calculate gas used
	gasUsed := gas - leftOverGas

	// Handle errors similar to the Solidity assembly code:
	// switch result
	// case 0 { revert(0, returndatasize) }
	// default { return (0, returndatasize) }
	if err != nil {
		return nil, gasUsed, err
	}

	return ret, gasUsed, nil
}

// _sealEpoch_offline is an internal function to seal offline validators in an epoch
func handle_sealEpoch_offline(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _sealEpoch_offline handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _sealEpoch_rewards is an internal function to seal rewards in an epoch
func handle_sealEpoch_rewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _sealEpoch_rewards handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _sealEpoch_minGasPrice is an internal function to seal minimum gas price in an epoch
func handle_sealEpoch_minGasPrice(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _sealEpoch_minGasPrice handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _calcRawValidatorEpochBaseReward is an internal function to calculate raw validator epoch base reward
func handle_calcRawValidatorEpochBaseReward(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _calcRawValidatorEpochBaseReward handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _calcRawValidatorEpochTxReward is an internal function to calculate raw validator epoch transaction reward
func handle_calcRawValidatorEpochTxReward(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _calcRawValidatorEpochTxReward handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _calcValidatorCommission is an internal function to calculate validator commission
func handle_calcValidatorCommission(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _calcValidatorCommission handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _mintNativeToken is an internal function to mint native tokens
func handle_mintNativeToken(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _mintNativeToken handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _scaleLockupReward is an internal function to scale lockup reward
func handle_scaleLockupReward(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _scaleLockupReward handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _setValidatorDeactivated is an internal function to set a validator as deactivated
func handle_setValidatorDeactivated(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _setValidatorDeactivated handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _syncValidator is an internal function to sync validator data
func handle_syncValidator(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _syncValidator handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _validatorExists is an internal function to check if a validator exists
func handle_validatorExists(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _validatorExists handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _now is an internal function to get the current time
func handle_now(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _now handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// getSlashingPenalty is an internal function to get the slashing penalty
func handleGetSlashingPenalty(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getSlashingPenalty handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleIsNode checks if the caller is the node (address(0))
func handleIsNode(evm *vm.EVM, caller common.Address) (bool, error) {
	// Check if caller is address(0)
	emptyAddr := common.Address{}
	return caller.Cmp(emptyAddr) == 0, nil
}

// handleGetUnlockedStake returns the unlocked stake of a delegator
func handleGetUnlockedStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Get the arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	delegator, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the delegation stake
	stakeSlot := getStakeSlot(delegator, toValidatorID)
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())

	// Get the delegation locked stake
	lockedStakeSlot := getLockedStakeSlot(delegator, toValidatorID)
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Calculate the unlocked stake
	unlockedStake := new(big.Int).Sub(stakeBigInt, lockedStakeBigInt)
	if unlockedStake.Cmp(big.NewInt(0)) < 0 {
		unlockedStake = big.NewInt(0)
	}

	// Pack the result
	result, err := SfcAbi.Methods["getUnlockedStake"].Outputs.Pack(unlockedStake)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, 0, nil
}

// handleCheckAllowedToWithdraw checks if a delegator is allowed to withdraw
func handleCheckAllowedToWithdraw(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int) (bool, error) {
	// Get the validator status
	validatorStatusSlot := getValidatorStatusSlot(toValidatorID)
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if the validator is deactivated
	isDeactivated := (validatorStatusBigInt.Bit(0) == 1) // WITHDRAWN_BIT

	// Get the validator auth
	validatorAuthSlot := getValidatorAuthSlot(toValidatorID)
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Check if the delegator is the validator auth
	isAuth := (delegator.Cmp(validatorAuthAddr) == 0)

	// A delegator is allowed to withdraw if the validator is deactivated or if the delegator is the validator auth
	return isDeactivated || isAuth, nil
}

// handleCheckDelegatedStakeLimit checks if a validator's delegated stake is within the limit
func handleCheckDelegatedStakeLimit(evm *vm.EVM, validatorID *big.Int) (bool, error) {
	// Get the self-stake
	// Create arguments for handleGetSelfStake
	args := []interface{}{validatorID}
	// Call handleGetSelfStake
	result, _, err := handleGetSelfStake(evm, args)
	if err != nil {
		return false, err
	}

	// Unpack the result
	selfStakeValues, err := SfcAbi.Methods["getSelfStake"].Outputs.Unpack(result)
	if err != nil {
		return false, vm.ErrExecutionReverted
	}

	// The result should be a single *big.Int value
	if len(selfStakeValues) != 1 {
		return false, vm.ErrExecutionReverted
	}

	selfStake, ok := selfStakeValues[0].(*big.Int)
	if !ok {
		return false, vm.ErrExecutionReverted
	}

	// Get the validator's received stake
	validatorReceivedStakeSlot := getValidatorReceivedStakeSlot(validatorID)
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

	// Get the max delegated ratio
	maxDelegatedRatioBigInt, err := getMaxDelegatedRatio(evm)
	if err != nil {
		return false, err
	}

	// Calculate the delegated stake
	delegatedStake := new(big.Int).Sub(receivedStakeBigInt, selfStake)
	if delegatedStake.Cmp(big.NewInt(0)) < 0 {
		delegatedStake = big.NewInt(0)
	}

	// Calculate the maximum allowed delegated stake
	maxDelegatedStake := new(big.Int).Mul(selfStake, maxDelegatedRatioBigInt)
	maxDelegatedStake = new(big.Int).Div(maxDelegatedStake, big.NewInt(1e18)) // Divide by Decimal.unit()

	// Check if the delegated stake is within the limit
	return delegatedStake.Cmp(maxDelegatedStake) <= 0, nil
}

// handleGetSelfStake returns the self-stake of a validator
func handleGetSelfStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the validator auth
	validatorAuthSlot := getValidatorAuthSlot(validatorID)
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Get the self-stake
	stakeSlot := getStakeSlot(validatorAuthAddr, validatorID)
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getSelfStake"].Outputs.Pack(stakeBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, 0, nil
}

// handleStashRewards stashes the rewards for a delegator
func handleStashRewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Get the arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	delegator, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	// Get the current epoch
	currentEpochBigInt, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}

	// Get the stashed rewards until epoch
	stashedRewardsUntilEpochSlot := getStashedRewardsUntilEpochSlot(delegator, toValidatorID)
	stashedRewardsUntilEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stashedRewardsUntilEpochSlot)))
	stashedRewardsUntilEpochBigInt := new(big.Int).SetBytes(stashedRewardsUntilEpoch.Bytes())

	// Check if rewards are already stashed for the current epoch
	if stashedRewardsUntilEpochBigInt.Cmp(currentEpochBigInt) >= 0 {
		return nil, 0, nil
	}

	// Calculate the rewards
	// TODO: Implement reward calculation
	rewards := big.NewInt(0) // Placeholder

	// Get the current stashed rewards
	rewardsStashSlot := getRewardsStashSlot(delegator, toValidatorID)
	rewardsStash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)))
	rewardsStashBigInt := new(big.Int).SetBytes(rewardsStash.Bytes())

	// Add the rewards to the stash
	newRewardsStash := new(big.Int).Add(rewardsStashBigInt, rewards)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)), common.BigToHash(newRewardsStash))

	// Update the stashed rewards until epoch
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(stashedRewardsUntilEpochSlot)), common.BigToHash(currentEpochBigInt))

	return nil, 0, nil
}

// handleSyncValidator synchronizes a validator's state
func handleSyncValidator(evm *vm.EVM, validatorID *big.Int) ([]byte, uint64, error) {
	// Get the validator status
	validatorStatusSlot := getValidatorStatusSlot(validatorID)
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if the validator is active
	isActive := (validatorStatusBigInt.Cmp(big.NewInt(0)) == 0) // OK_STATUS

	// Get the self-stake
	// Create arguments for handleGetSelfStake
	args := []interface{}{validatorID}
	// Call handleGetSelfStake
	result, _, err := handleGetSelfStake(evm, args)
	if err != nil {
		return nil, 0, err
	}

	// Unpack the result
	selfStakeValues, err := SfcAbi.Methods["getSelfStake"].Outputs.Unpack(result)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// The result should be a single *big.Int value
	if len(selfStakeValues) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}

	selfStake, ok := selfStakeValues[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the minimum self-stake
	minSelfStakeBigInt, err := getMinSelfStake(evm)
	if err != nil {
		return nil, 0, err
	}

	// Check if the self-stake is at least the minimum self-stake
	hasSelfStake := selfStake.Cmp(big.NewInt(0)) > 0
	hasEnoughSelfStake := selfStake.Cmp(minSelfStakeBigInt) >= 0

	// Check if the delegated stake is within the limit
	withinDelegatedLimit, err := handleCheckDelegatedStakeLimit(evm, validatorID)
	if err != nil {
		return nil, 0, err
	}

	// Update the validator status if necessary
	if isActive && (!hasSelfStake || !hasEnoughSelfStake || !withinDelegatedLimit) {
		// Set the validator as deactivated
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT

		// Set the validator deactivated epoch
		validatorDeactivatedEpochSlot := getValidatorDeactivatedEpochSlot(validatorID)
		currentEpochBigInt, err := getCurrentEpoch(evm)
		if err != nil {
			return nil, 0, err
		}
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedEpochSlot)), common.BigToHash(currentEpochBigInt))

		// Set the validator deactivated time
		validatorDeactivatedTimeSlot := getValidatorDeactivatedTimeSlot(validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedTimeSlot)), common.BigToHash(evm.Context.Time))

		// Update the total active stake
		validatorReceivedStakeSlot := getValidatorReceivedStakeSlot(validatorID)
		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Sub(totalActiveStakeBigInt, receivedStakeBigInt)
		if newTotalActiveStake.Cmp(big.NewInt(0)) < 0 {
			newTotalActiveStake = big.NewInt(0)
		}
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	} else if !isActive && hasSelfStake && hasEnoughSelfStake && withinDelegatedLimit {
		// Set the validator as active
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)), common.BigToHash(big.NewInt(0))) // OK_STATUS

		// Clear the validator deactivated epoch
		validatorDeactivatedEpochSlot := getValidatorDeactivatedEpochSlot(validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedEpochSlot)), common.BigToHash(big.NewInt(0)))

		// Clear the validator deactivated time
		validatorDeactivatedTimeSlot := getValidatorDeactivatedTimeSlot(validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedTimeSlot)), common.BigToHash(big.NewInt(0)))

		// Update the total active stake
		validatorReceivedStakeSlot := getValidatorReceivedStakeSlot(validatorID)
		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Add(totalActiveStakeBigInt, receivedStakeBigInt)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	}

	return nil, 0, nil
}
