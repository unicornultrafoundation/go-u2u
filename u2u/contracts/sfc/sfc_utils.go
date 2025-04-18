package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// Gas costs for storage operations
const (
	SloadGasCost  uint64 = 2100  // Cost of SLOAD (GetState) operation (ColdSloadCostEIP2929)
	SstoreGasCost uint64 = 20000 // Cost of SSTORE (SetState) operation (SstoreSetGasEIP2200)
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
func getValidatorStatusSlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// The Validator struct has the following fields:
	// uint256 status;
	// uint256 deactivatedTime;
	// uint256 deactivatedEpoch;
	// uint256 receivedStake;
	// uint256 createdEpoch;
	// uint256 createdTime;
	// address auth;

	// For a mapping(uint256 => Validator), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorSlot))
	// Then, for the status field (which is the first field), we use that slot directly

	// Create the hash input: abi.encode(validatorID, validatorSlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                 // Left-pad to 32 bytes
	validatorSlotBytes := common.LeftPadBytes(big.NewInt(validatorSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The status field is at slot + 0
	return slot.Int64(), gasUsed
}

// getValidatorCreatedEpochSlot calculates the storage slot for a validator's created epoch
func getValidatorCreatedEpochSlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// The Validator struct has the following fields:
	// uint256 status;
	// uint256 deactivatedTime;
	// uint256 deactivatedEpoch;
	// uint256 receivedStake;
	// uint256 createdEpoch;
	// uint256 createdTime;
	// address auth;

	// For a mapping(uint256 => Validator), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorSlot))
	// Then, for the createdEpoch field (which is the fifth field), we add 4 to the base slot

	// Create the hash input: abi.encode(validatorID, validatorSlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                 // Left-pad to 32 bytes
	validatorSlotBytes := common.LeftPadBytes(big.NewInt(validatorSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The createdEpoch field is at slot + 4
	return new(big.Int).Add(slot, big.NewInt(4)).Int64(), gasUsed
}

// getValidatorCreatedTimeSlot calculates the storage slot for a validator's created time
func getValidatorCreatedTimeSlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// The Validator struct has the following fields:
	// uint256 status;
	// uint256 deactivatedTime;
	// uint256 deactivatedEpoch;
	// uint256 receivedStake;
	// uint256 createdEpoch;
	// uint256 createdTime;
	// address auth;

	// For a mapping(uint256 => Validator), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorSlot))
	// Then, for the createdTime field (which is the sixth field), we add 5 to the base slot

	// Create the hash input: abi.encode(validatorID, validatorSlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                 // Left-pad to 32 bytes
	validatorSlotBytes := common.LeftPadBytes(big.NewInt(validatorSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The createdTime field is at slot + 5
	return new(big.Int).Add(slot, big.NewInt(5)).Int64(), gasUsed
}

// getValidatorDeactivatedEpochSlot calculates the storage slot for a validator's deactivated epoch
func getValidatorDeactivatedEpochSlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// The Validator struct has the following fields:
	// uint256 status;
	// uint256 deactivatedTime;
	// uint256 deactivatedEpoch;
	// uint256 receivedStake;
	// uint256 createdEpoch;
	// uint256 createdTime;
	// address auth;

	// For a mapping(uint256 => Validator), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorSlot))
	// Then, for the deactivatedEpoch field (which is the third field), we add 2 to the base slot

	// Create the hash input: abi.encode(validatorID, validatorSlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                 // Left-pad to 32 bytes
	validatorSlotBytes := common.LeftPadBytes(big.NewInt(validatorSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The deactivatedEpoch field is at slot + 2
	return new(big.Int).Add(slot, big.NewInt(2)).Int64(), gasUsed
}

// getValidatorDeactivatedTimeSlot calculates the storage slot for a validator's deactivated time
func getValidatorDeactivatedTimeSlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// The Validator struct has the following fields:
	// uint256 status;
	// uint256 deactivatedTime;
	// uint256 deactivatedEpoch;
	// uint256 receivedStake;
	// uint256 createdEpoch;
	// uint256 createdTime;
	// address auth;

	// For a mapping(uint256 => Validator), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorSlot))
	// Then, for the deactivatedTime field (which is the second field), we add 1 to the base slot

	// Create the hash input: abi.encode(validatorID, validatorSlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                 // Left-pad to 32 bytes
	validatorSlotBytes := common.LeftPadBytes(big.NewInt(validatorSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The deactivatedTime field is at slot + 1
	return new(big.Int).Add(slot, big.NewInt(1)).Int64(), gasUsed
}

// getValidatorCommissionSlot calculates the storage slot for a validator's commission
func getValidatorCommissionSlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(uint256 => uint256), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorCommissionSlot))

	// Create the hash input: abi.encode(validatorID, validatorCommissionSlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                                     // Left-pad to 32 bytes
	validatorCommissionSlotBytes := common.LeftPadBytes(big.NewInt(validatorCommissionSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorCommissionSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot.Int64(), gasUsed
}

// getVoteBookAddressSlot returns the storage slot for the voteBookAddress
func getVoteBookAddressSlot() (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// This is a simple storage slot, not a mapping
	return voteBookAddressSlot, gasUsed
}

// getValidatorAuthSlot calculates the storage slot for a validator's auth address
func getValidatorAuthSlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// The Validator struct has the following fields:
	// uint256 status;
	// uint256 deactivatedTime;
	// uint256 deactivatedEpoch;
	// uint256 receivedStake;
	// uint256 createdEpoch;
	// uint256 createdTime;
	// address auth;

	// For a mapping(uint256 => Validator), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorSlot))
	// Then, for the auth field (which is the seventh field), we add 6 to the base slot

	// Create the hash input: abi.encode(validatorID, validatorSlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                 // Left-pad to 32 bytes
	validatorSlotBytes := common.LeftPadBytes(big.NewInt(validatorSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The auth field is at slot + 6
	return new(big.Int).Add(slot, big.NewInt(6)).Int64(), gasUsed
}

// getValidatorPubkeySlot calculates the storage slot for a validator's pubkey
func getValidatorPubkeySlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(uint256 => bytes), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorPubkeySlot))

	// Create the hash input: abi.encode(validatorID, validatorPubkeySlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                             // Left-pad to 32 bytes
	validatorPubkeySlotBytes := common.LeftPadBytes(big.NewInt(validatorPubkeySlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorPubkeySlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot.Int64(), gasUsed
}

// getStakeSlot calculates the storage slot for a delegator's stake
func getStakeSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, stakeSlot))))

	// Create the inner hash input: abi.encode(delegator, stakeSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)             // Left-pad to 32 bytes
	stakeSlotBytes := common.LeftPadBytes(big.NewInt(stakeSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, stakeSlotBytes...)

	// Calculate the inner hash
	innerHash := crypto.Keccak256(innerHashInput)

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// getValidatorReceivedStakeSlot calculates the storage slot for a validator's received stake
func getValidatorReceivedStakeSlot(validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// The Validator struct has the following fields:
	// uint256 status;
	// uint256 deactivatedTime;
	// uint256 deactivatedEpoch;
	// uint256 receivedStake;
	// uint256 createdEpoch;
	// uint256 createdTime;
	// address auth;

	// For a mapping(uint256 => Validator), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorSlot))
	// Then, for the receivedStake field (which is the fourth field), we add 3 to the base slot

	// Create the hash input: abi.encode(validatorID, validatorSlot)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)                 // Left-pad to 32 bytes
	validatorSlotBytes := common.LeftPadBytes(big.NewInt(validatorSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(validatorIDBytes, validatorSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The receivedStake field is at slot + 3
	return new(big.Int).Add(slot, big.NewInt(3)).Int64(), gasUsed
}

// getWithdrawalRequestSlot calculates the storage slot for a withdrawal request
func getWithdrawalRequestSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => mapping(uint256 => WithdrawalRequest))), we need to calculate the slot in multiple steps

	// Step 1: Calculate keccak256(abi.encode(delegator, withdrawalRequestSlot))
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                                     // Left-pad to 32 bytes
	withdrawalRequestSlotBytes := common.LeftPadBytes(big.NewInt(withdrawalRequestSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput1 := append(delegatorBytes, withdrawalRequestSlotBytes...)
	innerHash1 := crypto.Keccak256(innerHashInput1)

	// Step 2: Calculate keccak256(abi.encode(toValidatorID, innerHash1))
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput2 := append(toValidatorIDBytes, innerHash1...)
	innerHash2 := crypto.Keccak256(innerHashInput2)

	// Step 3: Calculate keccak256(abi.encode(wrID, innerHash2))
	wrIDBytes := common.LeftPadBytes(wrID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(wrIDBytes, innerHash2...)
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The WithdrawalRequest struct has the following fields:
	// uint256 amount;
	// uint256 epoch;
	// uint256 time;

	// We're returning the base slot for the struct
	return slot.Int64(), gasUsed
}

// getWithdrawalRequestAmountSlot calculates the storage slot for a withdrawal request amount
func getWithdrawalRequestAmountSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) int64 {
	// TODO: Implement proper slot calculation
	return withdrawalRequestSlot
}

// getWithdrawalRequestEpochSlot calculates the storage slot for a withdrawal request epoch
func getWithdrawalRequestEpochSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => mapping(uint256 => WithdrawalRequest))), we need to calculate the slot in multiple steps

	// Step 1: Calculate keccak256(abi.encode(delegator, withdrawalRequestSlot))
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                                     // Left-pad to 32 bytes
	withdrawalRequestSlotBytes := common.LeftPadBytes(big.NewInt(withdrawalRequestSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput1 := append(delegatorBytes, withdrawalRequestSlotBytes...)
	innerHash1 := crypto.Keccak256(innerHashInput1)

	// Step 2: Calculate keccak256(abi.encode(toValidatorID, innerHash1))
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput2 := append(toValidatorIDBytes, innerHash1...)
	innerHash2 := crypto.Keccak256(innerHashInput2)

	// Step 3: Calculate keccak256(abi.encode(wrID, innerHash2))
	wrIDBytes := common.LeftPadBytes(wrID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(wrIDBytes, innerHash2...)
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The WithdrawalRequest struct has the following fields:
	// uint256 amount;
	// uint256 epoch;
	// uint256 time;

	// The epoch field is at slot + 1
	return new(big.Int).Add(slot, big.NewInt(1)).Int64(), gasUsed
}

// getWithdrawalRequestTimeSlot calculates the storage slot for a withdrawal request time
func getWithdrawalRequestTimeSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => mapping(uint256 => WithdrawalRequest))), we need to calculate the slot in multiple steps

	// Step 1: Calculate keccak256(abi.encode(delegator, withdrawalRequestSlot))
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                                     // Left-pad to 32 bytes
	withdrawalRequestSlotBytes := common.LeftPadBytes(big.NewInt(withdrawalRequestSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput1 := append(delegatorBytes, withdrawalRequestSlotBytes...)
	innerHash1 := crypto.Keccak256(innerHashInput1)

	// Step 2: Calculate keccak256(abi.encode(toValidatorID, innerHash1))
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput2 := append(toValidatorIDBytes, innerHash1...)
	innerHash2 := crypto.Keccak256(innerHashInput2)

	// Step 3: Calculate keccak256(abi.encode(wrID, innerHash2))
	wrIDBytes := common.LeftPadBytes(wrID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(wrIDBytes, innerHash2...)
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The WithdrawalRequest struct has the following fields:
	// uint256 amount;
	// uint256 epoch;
	// uint256 time;

	// The time field is at slot + 2
	return new(big.Int).Add(slot, big.NewInt(2)).Int64(), gasUsed
}

// getValidatorIDSlot calculates the storage slot for a validator ID
func getValidatorIDSlot(addr common.Address) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => uint256), the slot is calculated as:
	// keccak256(abi.encode(addr, validatorIDSlot))

	// Create the hash input: abi.encode(addr, validatorIDSlot)
	addrBytes := common.LeftPadBytes(addr.Bytes(), 32)                                   // Left-pad to 32 bytes
	validatorIDSlotBytes := common.LeftPadBytes(big.NewInt(validatorIDSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(addrBytes, validatorIDSlotBytes...)

	// Calculate the hash
	hash := crypto.Keccak256(hashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot.Int64(), gasUsed
}

// getLockedStakeSlot calculates the storage slot for a delegation's locked stake
func getLockedStakeSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => LockedDelegation)), first we need to get the slot for the LockedDelegation struct
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, lockupInfoSlot))))

	// Create the inner hash input: abi.encode(delegator, lockupInfoSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                       // Left-pad to 32 bytes
	lockupInfoSlotBytes := common.LeftPadBytes(big.NewInt(lockupInfoSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, lockupInfoSlotBytes...)

	// Calculate the inner hash
	innerHash := crypto.Keccak256(innerHashInput)

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The lockedStake field is at slot + 0
	return slot.Int64(), gasUsed
}

// getLockupFromEpochSlot calculates the storage slot for a delegation's lockup from epoch
func getLockupFromEpochSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => LockedDelegation)), first we need to get the slot for the LockedDelegation struct
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, lockupInfoSlot))))

	// Create the inner hash input: abi.encode(delegator, lockupInfoSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                       // Left-pad to 32 bytes
	lockupInfoSlotBytes := common.LeftPadBytes(big.NewInt(lockupInfoSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, lockupInfoSlotBytes...)

	// Calculate the inner hash
	innerHash := crypto.Keccak256(innerHashInput)

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The fromEpoch field is at slot + 1
	return new(big.Int).Add(slot, big.NewInt(1)).Int64(), gasUsed
}

// getLockupEndTimeSlot calculates the storage slot for a delegation's lockup end time
func getLockupEndTimeSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => LockedDelegation)), first we need to get the slot for the LockedDelegation struct
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, lockupInfoSlot))))

	// Create the inner hash input: abi.encode(delegator, lockupInfoSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                       // Left-pad to 32 bytes
	lockupInfoSlotBytes := common.LeftPadBytes(big.NewInt(lockupInfoSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, lockupInfoSlotBytes...)

	// Calculate the inner hash
	innerHash := crypto.Keccak256(innerHashInput)

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The endTime field is at slot + 2
	return new(big.Int).Add(slot, big.NewInt(2)).Int64(), gasUsed
}

// getLockupDurationSlot calculates the storage slot for a delegation's lockup duration
func getLockupDurationSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => LockedDelegation)), first we need to get the slot for the LockedDelegation struct
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, lockupInfoSlot))))

	// Create the inner hash input: abi.encode(delegator, lockupInfoSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                       // Left-pad to 32 bytes
	lockupInfoSlotBytes := common.LeftPadBytes(big.NewInt(lockupInfoSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, lockupInfoSlotBytes...)

	// Calculate the inner hash
	innerHash := crypto.Keccak256(innerHashInput)

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The duration field is at slot + 3
	return new(big.Int).Add(slot, big.NewInt(3)).Int64(), gasUsed
}

// getEarlyWithdrawalPenaltySlot calculates the storage slot for a delegation's early withdrawal penalty
func getEarlyWithdrawalPenaltySlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// TODO: Implement proper slot calculation
	return stakeSlot, gasUsed
}

// getRewardsStashSlot calculates the storage slot for a delegation's rewards stash
func getRewardsStashSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => Rewards)), first we need to get the slot for the Rewards struct
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, rewardsStashSlot))))

	// Create the inner hash input: abi.encode(delegator, rewardsStashSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                           // Left-pad to 32 bytes
	rewardsStashSlotBytes := common.LeftPadBytes(big.NewInt(rewardsStashSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, rewardsStashSlotBytes...)

	// Calculate the inner hash
	innerHash := crypto.Keccak256(innerHashInput)

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// getStashedRewardsUntilEpochSlot calculates the storage slot for a delegation's stashed rewards until epoch
func getStashedRewardsUntilEpochSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, stashedRewardsUntilEpochSlot))))

	// Create the inner hash input: abi.encode(delegator, stashedRewardsUntilEpochSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                                                   // Left-pad to 32 bytes
	stashedRewardsUntilEpochSlotBytes := common.LeftPadBytes(big.NewInt(stashedRewardsUntilEpochSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, stashedRewardsUntilEpochSlotBytes...)

	// Calculate the inner hash
	innerHash := crypto.Keccak256(innerHashInput)

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash
	outerHash := crypto.Keccak256(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// callConstantManagerMethod calls a method on the ConstantManager contract and returns the result
// methodName: the name of the method to call
// args: the arguments to pass to the method
// Returns: the result of the method call, the gas used, or an error if the call failed
func callConstantManagerMethod(evm *vm.EVM, methodName string, args ...interface{}) ([]interface{}, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the ConstantsManager contract address (SLOAD operation)
	constantsManager := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(constantsManagerSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	constantsManagerAddr := common.BytesToAddress(constantsManager.Bytes())

	// Pack the function call data
	data, err := CMAbi.Pack(methodName, args...)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Make the call to the ConstantsManager contract
	result, leftOverGas, err := evm.Call(vm.AccountRef(ContractAddress), constantsManagerAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, gasUsed + (50000 - leftOverGas), err
	}

	// Add the gas used by the call
	gasUsed += (50000 - leftOverGas)

	// Unpack the result
	values, err := CMAbi.Methods[methodName].Outputs.Unpack(result)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return values, gasUsed, nil
}

// getCurrentEpoch returns the current epoch value (currentSealedEpoch + 1)
// This implements the logic from the currentEpoch() function in SFCBase.sol
func getCurrentEpoch(evm *vm.EVM) (*big.Int, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the current sealed epoch (SLOAD operation)
	currentSealedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	currentSealedEpochBigInt := new(big.Int).SetBytes(currentSealedEpoch.Bytes())

	// Calculate current epoch as currentSealedEpoch + 1
	currentEpochBigInt := new(big.Int).Add(currentSealedEpochBigInt, big.NewInt(1))

	return currentEpochBigInt, gasUsed, nil
}

// getMinSelfStake returns the minimum self-stake value from the ConstantManager contract
func getMinSelfStake(evm *vm.EVM) (*big.Int, uint64, error) {
	// Call the minSelfStake method on the ConstantManager contract
	values, gasUsed, err := callConstantManagerMethod(evm, "minSelfStake")
	if err != nil {
		return nil, gasUsed, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	minSelfStakeBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return minSelfStakeBigInt, gasUsed, nil
}

// getMaxDelegatedRatio returns the maximum delegated ratio value from the ConstantManager contract
func getMaxDelegatedRatio(evm *vm.EVM) (*big.Int, uint64, error) {
	// Call the maxDelegatedRatio method on the ConstantManager contract
	values, gasUsed, err := callConstantManagerMethod(evm, "maxDelegatedRatio")
	if err != nil {
		return nil, gasUsed, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	maxDelegatedRatioBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return maxDelegatedRatioBigInt, gasUsed, nil
}

// getWithdrawalPeriodEpochs returns the withdrawal period epochs value from the ConstantManager contract
func getWithdrawalPeriodEpochs(evm *vm.EVM) (*big.Int, uint64, error) {
	// Call the withdrawalPeriodEpochs method on the ConstantManager contract
	values, gasUsed, err := callConstantManagerMethod(evm, "withdrawalPeriodEpochs")
	if err != nil {
		return nil, gasUsed, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	withdrawalPeriodEpochsBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return withdrawalPeriodEpochsBigInt, gasUsed, nil
}

// getMinLockupDuration returns the minimum lockup duration value from the ConstantManager contract
func getMinLockupDuration(evm *vm.EVM) (*big.Int, uint64, error) {
	// Call the minLockupDuration method on the ConstantManager contract
	values, gasUsed, err := callConstantManagerMethod(evm, "minLockupDuration")
	if err != nil {
		return nil, gasUsed, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	minLockupDurationBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return minLockupDurationBigInt, gasUsed, nil
}

// getMaxLockupDuration returns the maximum lockup duration value from the ConstantManager contract
func getMaxLockupDuration(evm *vm.EVM) (*big.Int, uint64, error) {
	// Call the maxLockupDuration method on the ConstantManager contract
	values, gasUsed, err := callConstantManagerMethod(evm, "maxLockupDuration")
	if err != nil {
		return nil, gasUsed, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	maxLockupDurationBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return maxLockupDurationBigInt, gasUsed, nil
}

// getWithdrawalPeriodTime returns the withdrawal period time value from the ConstantManager contract
func getWithdrawalPeriodTime(evm *vm.EVM) (*big.Int, uint64, error) {
	// Call the withdrawalPeriodTime method on the ConstantManager contract
	values, gasUsed, err := callConstantManagerMethod(evm, "withdrawalPeriodTime")
	if err != nil {
		return nil, gasUsed, err
	}

	// The result should be a single *big.Int value
	if len(values) != 1 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	withdrawalPeriodTimeBigInt, ok := values[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return withdrawalPeriodTimeBigInt, gasUsed, nil
}
