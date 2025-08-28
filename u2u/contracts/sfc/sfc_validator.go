package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleCreateValidator creates a new validator
func handleCreateValidator(evm *vm.EVM, caller common.Address, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	pubkey, ok := args[0].([]byte)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the pubkey is not empty
	if len(pubkey) == 0 {
		revertData, err := encodeRevertReason("createValidator", "empty pubkey")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Call the minSelfStake method on the ConstantsManager contract
	minSelfStake := getConstantsManagerVariable(evm.SfcStateDB, "minSelfStake")
	// Check that the value is at least the minimum self-stake
	if value.Cmp(minSelfStake) < 0 {
		revertData, err := encodeRevertReason("createValidator", "insufficient self-stake")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Call the internal _createValidator function
	validatorIDBytes, createValidatorGasUsed, err := handleInternalCreateValidator(evm, caller, pubkey)
	gasUsed += createValidatorGasUsed
	if err != nil {
		return validatorIDBytes, gasUsed, err
	}

	// Convert validator ID bytes back to *big.Int for delegate call
	newValidatorID := new(big.Int).SetBytes(validatorIDBytes)

	// Delegate the value to the validator
	// This is equivalent to _delegate(msg.sender, lastValidatorID, msg.value)
	result, delegateGasUsed, err := handleInternalDelegate(evm, caller, newValidatorID, value)
	if err != nil {
		return result, gasUsed + delegateGasUsed, err
	}

	// Add the gas used by handleInternalDelegate
	gasUsed += delegateGasUsed

	return nil, gasUsed, nil
}

// handleDeactivateValidator deactivates a validator
func handleDeactivateValidator(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Check if caller is the NodeDriverAuth contract (onlyDriver modifier)
	revertData, checkGasUsed, err := checkOnlyDriver(evm, caller, "deactivateValidator")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Parse arguments: validatorID, status
	if len(args) != 2 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse validatorID (uint256)
	validatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse status (uint256)
	status, ok := args[1].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// require(status != OK_STATUS, "wrong status")
	if status.Cmp(big.NewInt(int64(OK_STATUS))) == 0 {
		revertData, err := encodeRevertReason("deactivateValidator", "wrong status")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Call _setValidatorDeactivated(validatorID, status)
	setDeactivatedGasUsed, err := handleInternalSetValidatorDeactivated(evm, validatorID, status.Uint64())
	gasUsed += setDeactivatedGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Call _syncValidator(validatorID, false)
	syncValidatorGasUsed, err := handleInternalSyncValidator(evm, validatorID, false)
	gasUsed += syncValidatorGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Get validatorAddr = getValidator[validatorID].auth
	validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(validatorID)
	gasUsed += slotGasUsed

	validatorAuthHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
	gasUsed += SloadGasCost

	validatorAuth := common.BytesToAddress(validatorAuthHash.Bytes())

	// Call _recountVotes(validatorAddr, validatorAddr, false)
	_, recountVotesGasUsed, err := handleRecountVotes(evm, validatorAuth, validatorAuth, false)
	gasUsed += recountVotesGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	return nil, gasUsed, nil
}

// handleSetGenesisValidator sets a genesis validator
func handleSetGenesisValidator(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Check if caller is the NodeDriverAuth contract (onlyDriver modifier)
	revertData, checkGasUsed, err := checkOnlyDriver(evm, caller, "setGenesisValidator")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Parse arguments: auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime
	if len(args) != 8 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse auth (address)
	auth, ok := args[0].(common.Address)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse validatorID (uint256)
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse pubkey (bytes)
	pubkey, ok := args[2].([]byte)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse status (uint256)
	status, ok := args[3].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse createdEpoch (uint256)
	createdEpoch, ok := args[4].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse createdTime (uint256)
	createdTime, ok := args[5].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse deactivatedEpoch (uint256)
	deactivatedEpoch, ok := args[6].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Parse deactivatedTime (uint256)
	deactivatedTime, ok := args[7].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Call _rawCreateValidator(auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
	revertData, rawCreateGasUsed, err := handleInternalRawCreateValidator(
		evm,
		auth,
		validatorID,
		pubkey,
		status,
		createdEpoch,
		createdTime,
		deactivatedEpoch,
		deactivatedTime,
	)
	gasUsed += rawCreateGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Update lastValidatorID if validatorID > lastValidatorID
	// Get current lastValidatorID
	lastValidatorIDSlotHash := common.BigToHash(big.NewInt(lastValidatorIDSlot))
	currentLastValidatorID := evm.SfcStateDB.GetState(ContractAddress, lastValidatorIDSlotHash)
	gasUsed += SloadGasCost

	currentLastValidatorIDBigInt := new(big.Int).SetBytes(currentLastValidatorID.Bytes())

	// Check if validatorID > lastValidatorID
	if validatorID.Cmp(currentLastValidatorIDBigInt) > 0 {
		// Update lastValidatorID = validatorID
		evm.SfcStateDB.SetState(ContractAddress, lastValidatorIDSlotHash, common.BigToHash(validatorID))
		gasUsed += SstoreGasCost
	}

	return nil, gasUsed, nil
}
