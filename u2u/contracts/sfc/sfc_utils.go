package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// Gas costs and limits
const (
	defaultGasLimit uint64 = 3000000 // Default gas limit for contract calls

	SloadGasCost  uint64 = 2100  // Cost of SLOAD (GetState) operation (ColdSloadCostEIP2929)
	SstoreGasCost uint64 = 20000 // Cost of SSTORE (SetState) operation (SstoreSetGasEIP2200)
	HashGasCost   uint64 = 30    // Cost of hash operation (Keccak256)
)

// EpochSnapshot struct offsets
// These offsets are based on the storage layout of the EpochSnapshot struct in SFCState.sol
// In Solidity, struct members are allocated storage slots sequentially in the order they appear
const (
	// Mappings are stored at their assigned slot (but actual data is at hash(key . slot))
	receiveStakeOffset                int64 = 0 // mapping(uint256 => uint256) receivedStake
	accumulatedRewardPerTokenOffset   int64 = 1 // mapping(uint256 => uint256) accumulatedRewardPerToken
	accumulatedUptimeOffset           int64 = 2 // mapping(uint256 => uint256) accumulatedUptime
	accumulatedOriginatedTxsFeeOffset int64 = 3 // mapping(uint256 => uint256) accumulatedOriginatedTxsFee
	offlineTimeOffset                 int64 = 4 // mapping(uint256 => uint256) offlineTime
	offlineBlocksOffset               int64 = 5 // mapping(uint256 => uint256) offlineBlocks

	// Dynamic array length is stored at the slot
	validatorIDsOffset int64 = 6 // uint256[] validatorIDs

	// Fixed-size fields are stored sequentially
	endTimeOffset             int64 = 7  // uint256 endTime
	epochFeeOffset            int64 = 8  // uint256 epochFee
	totalBaseRewardOffset     int64 = 9  // uint256 totalBaseRewardWeight
	totalTxRewardOffset       int64 = 10 // uint256 totalTxRewardWeight
	baseRewardPerSecondOffset int64 = 11 // uint256 baseRewardPerSecond
	totalStakeOffset          int64 = 12 // uint256 totalStake
	totalSupplyOffset         int64 = 13 // uint256 totalSupply
)

// Validator status bits
const (
	OK_STATUS      uint64 = 0
	WITHDRAWN_BIT  uint64 = 1
	OFFLINE_BIT    uint64 = 1 << 3
	DOUBLESIGN_BIT uint64 = 1 << 7
	CHEATER_MASK   uint64 = DOUBLESIGN_BIT
)

// checkOnlyOwner checks if the caller is the owner of the contract
// Returns nil if the caller is the owner, otherwise returns an ABI-encoded revert reason
func checkOnlyOwner(evm *vm.EVM, caller common.Address, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Get owner address (SLOAD operation)
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	gasUsed += SloadGasCost

	ownerAddr := common.BytesToAddress(owner.Bytes())
	if caller.Cmp(ownerAddr) != 0 {
		// Return ABI-encoded revert reason: "Ownable: caller is not the owner"
		revertReason := "Ownable: caller is not the owner"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// checkOnlyDriver checks if the caller is the NodeDriverAuth contract
// Returns nil if the caller is the NodeDriverAuth, otherwise returns an ABI-encoded revert reason
func checkOnlyDriver(evm *vm.EVM, caller common.Address, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Get NodeDriverAuth address (SLOAD operation)
	node := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(nodeDriverAuthSlot)))
	gasUsed += SloadGasCost

	nodeAddr := common.BytesToAddress(node.Bytes())
	if caller.Cmp(nodeAddr) != 0 {
		// Return ABI-encoded revert reason: "caller is not the NodeDriverAuth contract"
		revertReason := "caller is not the NodeDriverAuth contract"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		log.Error("SFC: Caller is not the NodeDriverAuth contract", "caller", caller, "nodeDriverAuth", nodeAddr)
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// checkValidatorExists checks if a validator with the given ID exists
// Returns nil if the validator exists, otherwise returns an ABI-encoded revert reason
func checkValidatorExists(evm *vm.EVM, validatorID *big.Int, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Calculate validator status slot
	statusSlot, slotGasUsed := getValidatorStatusSlot(validatorID)
	gasUsed += slotGasUsed

	// Check if validator exists (SLOAD operation)
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(statusSlot))
	gasUsed += SloadGasCost

	emptyHash := common.Hash{}
	if validatorStatus.Cmp(emptyHash) == 0 {
		// Return ABI-encoded revert reason: "validator doesn't exist"
		revertReason := "validator doesn't exist"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// checkValidatorNotExists checks if a validator with the given ID does not exist
// Returns nil if the validator does not exist, otherwise returns an ABI-encoded revert reason
func checkValidatorNotExists(evm *vm.EVM, validatorID *big.Int, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Calculate validator status slot
	statusSlot, slotGasUsed := getValidatorStatusSlot(validatorID)
	gasUsed += slotGasUsed

	// Check if validator doesn't exist (SLOAD operation)
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(statusSlot))
	gasUsed += SloadGasCost

	emptyHash := common.Hash{}
	if validatorStatus.Cmp(emptyHash) != 0 {
		// Return ABI-encoded revert reason: "validator already exists"
		revertReason := "validator already exists"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// checkValidatorActive checks if a validator is active
// Returns nil if the validator is active, otherwise returns an ABI-encoded revert reason
func checkValidatorActive(evm *vm.EVM, validatorID *big.Int, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// First check if validator exists
	revertData, existsGasUsed, err := checkValidatorExists(evm, validatorID, methodName)
	gasUsed += existsGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Calculate validator status slot
	statusSlot, slotGasUsed := getValidatorStatusSlot(validatorID)
	gasUsed += slotGasUsed

	// Check if validator is active (SLOAD operation)
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(statusSlot))
	gasUsed += SloadGasCost

	statusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if validator is not deactivated
	if statusBigInt.Bit(0) == 1 { // WITHDRAWN_BIT
		// Return ABI-encoded revert reason: "validator is deactivated"
		revertReason := "validator is deactivated"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Check if validator is not offline
	if statusBigInt.Bit(3) == 1 { // OFFLINE_BIT
		// Return ABI-encoded revert reason: "validator is offline"
		revertReason := "validator is offline"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Check if validator is not a cheater
	if statusBigInt.Bit(7) == 1 { // DOUBLESIGN_BIT
		// Return ABI-encoded revert reason: "validator is a cheater"
		revertReason := "validator is a cheater"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	return nil, gasUsed, nil
}

// checkDelegationExists checks if a delegation exists
// Returns nil if the delegation exists, otherwise returns an ABI-encoded revert reason
func checkDelegationExists(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Calculate stake slot
	stakeSlotNum, slotGasUsed := getStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed

	// Check if delegation exists (SLOAD operation)
	delegation := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stakeSlotNum))
	gasUsed += SloadGasCost

	emptyHash := common.Hash{}
	if delegation.Cmp(emptyHash) == 0 {
		// Return ABI-encoded revert reason: "delegation doesn't exist"
		revertReason := "delegation doesn't exist"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// checkDelegationNotExists checks if a delegation does not exist
// Returns nil if the delegation does not exist, otherwise returns an ABI-encoded revert reason
func checkDelegationNotExists(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Calculate stake slot
	stakeSlotNum, slotGasUsed := getStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed

	// Check if delegation doesn't exist (SLOAD operation)
	delegation := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stakeSlotNum))
	gasUsed += SloadGasCost

	emptyHash := common.Hash{}
	if delegation.Cmp(emptyHash) != 0 {
		// Return ABI-encoded revert reason: "delegation already exists"
		revertReason := "delegation already exists"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// checkAlreadyInitialized checks if the contract is already initialized
// Returns nil if the contract is not initialized, otherwise returns an ABI-encoded revert reason
func checkAlreadyInitialized(evm *vm.EVM, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Check if contract is already initialized (SLOAD operation)
	initializedFlag := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(isInitialized)))
	gasUsed += SloadGasCost

	emptyHash := common.Hash{}
	if initializedFlag.Cmp(emptyHash) != 0 {
		// Return ABI-encoded revert reason: "already initialized"
		revertReason := "already initialized"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// checkZeroAddress checks if an address is the zero address
// Returns nil if the address is not zero, otherwise returns an ABI-encoded revert reason
func checkZeroAddress(addr common.Address, methodName string, message string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	emptyAddr := common.Address{}
	if addr.Cmp(emptyAddr) == 0 {
		// Return ABI-encoded revert reason with the provided message
		revertData, err := encodeRevertReason(methodName, message)
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// encodeRevertReason encodes a revert reason as an ABI-encoded error
func encodeRevertReason(methodName string, reason string) ([]byte, error) {
	// Create a cache key for this error message
	errorMessage := methodName + ": " + reason
	cacheKey := "Error:" + errorMessage

	// Check if we have this error message cached
	if cachedData, ok := sfcCache.AbiPackCache[cacheKey]; ok {
		return cachedData, nil
	}

	// Prepend the error signature: bytes4(keccak256("Error(string)"))
	errorSig := []byte{0x08, 0xc3, 0x79, 0xa0}

	// Pack the revert reason
	packedReason, err := abi.Arguments{{Type: abi.Type{T: abi.StringTy}}}.Pack(reason)
	if err != nil {
		return nil, err
	}

	// Use the byte slice pool for the result
	result := GetByteSlice()
	if cap(result) < len(errorSig)+len(packedReason) {
		// If the slice from the pool is too small, allocate a new one
		result = make([]byte, 0, len(errorSig)+len(packedReason))
	}

	// Combine the error signature and packed reason
	result = append(result, errorSig...)
	result = append(result, packedReason...)

	// Cache the result
	sfcCache.AbiPackCache[cacheKey] = result
	log.Debug("SFC: Revert", "message", errorMessage)
	return result, nil
}

// Helper functions for calculating validator storage slots

// getValidatorStatusSlot calculates the storage slot for a validator's status
func getValidatorStatusSlot(validatorID *big.Int) (*big.Int, uint64) {
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

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int using the pool
	slot := GetBigInt().SetBytes(hash)

	// The status field is at slot + 0
	return slot, gasUsed
}

// getValidatorCreatedEpochSlot calculates the storage slot for a validator's created epoch
func getValidatorCreatedEpochSlot(validatorID *big.Int) (*big.Int, uint64) {
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

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int using the pool
	slot := GetBigInt().SetBytes(hash)

	// The createdEpoch field is at slot + 4
	// Use the pool for the result as well
	result := GetBigInt().Add(slot, big.NewInt(4))

	// Return the slot to the pool
	PutBigInt(slot)

	return result, gasUsed
}

// getValidatorCreatedTimeSlot calculates the storage slot for a validator's created time
func getValidatorCreatedTimeSlot(validatorID *big.Int) (*big.Int, uint64) {
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

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The createdTime field is at slot + 5
	return new(big.Int).Add(slot, big.NewInt(5)), gasUsed
}

// getValidatorDeactivatedEpochSlot calculates the storage slot for a validator's deactivated epoch
func getValidatorDeactivatedEpochSlot(validatorID *big.Int) (*big.Int, uint64) {
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

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The deactivatedEpoch field is at slot + 2
	return new(big.Int).Add(slot, big.NewInt(2)), gasUsed
}

// getValidatorDeactivatedTimeSlot calculates the storage slot for a validator's deactivated time
func getValidatorDeactivatedTimeSlot(validatorID *big.Int) (*big.Int, uint64) {
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

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The deactivatedTime field is at slot + 1
	return new(big.Int).Add(slot, big.NewInt(1)), gasUsed
}

// getValidatorCommissionSlot calculates the storage slot for a validator's commission
func getValidatorCommissionSlot(validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(uint256 => uint256), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorCommissionSlot))

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorCommissionSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
}

// getValidatorAuthSlot calculates the storage slot for a validator's auth address
func getValidatorAuthSlot(validatorID *big.Int) (*big.Int, uint64) {
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

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The auth field is at slot + 6
	return new(big.Int).Add(slot, big.NewInt(6)), gasUsed
}

// getValidatorPubkeySlot calculates the storage slot for a validator's pubkey
func getValidatorPubkeySlot(validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(uint256 => bytes), the slot is calculated as:
	// keccak256(abi.encode(validatorID, validatorPubkeySlot))

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorPubkeySlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
}

// getStakeSlot calculates the storage slot for a delegator's stake
func getStakeSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, stakeSlot))))

	// Create the inner hash input using cached padded values
	innerHashInput := CreateAddressHashInput(delegator, stakeSlot)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := CachedKeccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input using cached nested hash input
	outerHashInput := CreateNestedHashInput(toValidatorID, innerHash)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := CachedKeccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot, gasUsed
}

// getValidatorReceivedStakeSlot calculates the storage slot for a validator's received stake
func getValidatorReceivedStakeSlot(validatorID *big.Int) (*big.Int, uint64) {
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

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, validatorSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// The receivedStake field is at slot + 3
	return new(big.Int).Add(slot, big.NewInt(3)), gasUsed
}

// getWithdrawalRequestSlot calculates the storage slot for a withdrawal request
func getWithdrawalRequestSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => mapping(uint256 => WithdrawalRequest))), we need to calculate the slot in multiple steps

	// Step 1: Calculate keccak256(abi.encode(delegator, withdrawalRequestSlot))
	innerHashInput1 := CreateAddressHashInput(delegator, withdrawalRequestSlot)
	innerHash1 := CachedKeccak256(innerHashInput1)
	gasUsed += HashGasCost

	// Step 2: Calculate keccak256(abi.encode(toValidatorID, innerHash1))
	innerHashInput2 := CreateNestedHashInput(toValidatorID, innerHash1)
	innerHash2 := CachedKeccak256(innerHashInput2)
	gasUsed += HashGasCost

	// Step 3: Calculate keccak256(abi.encode(wrID, innerHash2))
	outerHashInput := CreateNestedHashInput(wrID, innerHash2) // We can reuse the nested hash input function for any uint256
	outerHash := CachedKeccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The WithdrawalRequest struct has the following fields:
	// uint256 amount;
	// uint256 epoch;
	// uint256 time;

	// We're returning the base slot for the struct
	return slot, gasUsed
}

// getWithdrawalRequestAmountSlot calculates the storage slot for a withdrawal request amount
func getWithdrawalRequestAmountSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The amount field is at the base slot (first field in the struct)
	return baseSlot, gasUsed
}

// getWithdrawalRequestEpochSlot calculates the storage slot for a withdrawal request epoch
func getWithdrawalRequestEpochSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The WithdrawalRequest struct has the following fields:
	// uint256 amount;
	// uint256 epoch;
	// uint256 time;

	// The epoch field is at slot + 1
	return new(big.Int).Add(baseSlot, big.NewInt(1)), gasUsed
}

// getWithdrawalRequestTimeSlot calculates the storage slot for a withdrawal request time
func getWithdrawalRequestTimeSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The WithdrawalRequest struct has the following fields:
	// uint256 amount;
	// uint256 epoch;
	// uint256 time;

	// The time field is at slot + 2
	return new(big.Int).Add(baseSlot, big.NewInt(2)), gasUsed
}

// getValidatorIDSlot calculates the storage slot for a validator ID
func getValidatorIDSlot(addr common.Address) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => uint256), the slot is calculated as:
	// keccak256(abi.encode(addr, validatorIDSlot))

	// Create the hash input using cached padded values
	hashInput := CreateAddressHashInput(addr, validatorIDSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
}

// getLockedStakeSlot calculates the storage slot for a delegation's locked stake
func getLockedStakeSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => LockedDelegation)), first we need to get the slot for the LockedDelegation struct
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, lockupInfoSlot))))

	// Create the inner hash input using cached padded values
	innerHashInput := CreateAddressHashInput(delegator, lockupInfoSlot)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := CachedKeccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input using cached nested hash input
	outerHashInput := CreateNestedHashInput(toValidatorID, innerHash)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := CachedKeccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The lockedStake field is at slot + 0
	return slot, gasUsed
}

// getLockupFromEpochSlot calculates the storage slot for a delegation's lockup from epoch
func getLockupFromEpochSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the locked delegation
	baseSlot, gasUsed := getLockedStakeSlot(delegator, toValidatorID)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The fromEpoch field is at slot + 1
	return new(big.Int).Add(baseSlot, big.NewInt(1)), gasUsed
}

// getLockupEndTimeSlot calculates the storage slot for a delegation's lockup end time
func getLockupEndTimeSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the locked delegation
	baseSlot, gasUsed := getLockedStakeSlot(delegator, toValidatorID)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The endTime field is at slot + 2
	return new(big.Int).Add(baseSlot, big.NewInt(2)), gasUsed
}

// getLockupDurationSlot calculates the storage slot for a delegation's lockup duration
func getLockupDurationSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the locked delegation
	baseSlot, gasUsed := getLockedStakeSlot(delegator, toValidatorID)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The duration field is at slot + 3
	return new(big.Int).Add(baseSlot, big.NewInt(3)), gasUsed
}

// getEarlyWithdrawalPenaltySlot calculates the storage slot for a delegation's early withdrawal penalty
func getEarlyWithdrawalPenaltySlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// TODO: Implement proper slot calculation with gas tracking
	// For now, just return a placeholder slot
	return stakeSlot, gasUsed
}

// getRewardsStashSlot calculates the storage slot for a delegation's rewards stash
func getRewardsStashSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => Rewards)), first we need to get the slot for the Rewards struct
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, rewardsStashSlot))))

	// Create the inner hash input using cached padded values
	innerHashInput := CreateAddressHashInput(delegator, rewardsStashSlot)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := CachedKeccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input using cached nested hash input
	outerHashInput := CreateNestedHashInput(toValidatorID, innerHash)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := CachedKeccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot, gasUsed
}

// getStashedLockupRewardsSlot calculates the storage slot for a delegation's stashed lockup rewards
func getStashedLockupRewardsSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, stashedLockupRewardsSlot))))

	// Create the inner hash input using cached padded values
	innerHashInput := CreateAddressHashInput(delegator, stashedLockupRewardsSlot)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := CachedKeccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input using cached nested hash input
	outerHashInput := CreateNestedHashInput(toValidatorID, innerHash)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := CachedKeccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot, gasUsed
}

// getStashedRewardsUntilEpochSlot calculates the storage slot for a delegation's stashed rewards until epoch
func getStashedRewardsUntilEpochSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, stashedRewardsUntilEpochSlot))))

	// Create the inner hash input using cached padded values
	innerHashInput := CreateAddressHashInput(delegator, stashedRewardsUntilEpochSlot)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := CachedKeccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input using cached nested hash input
	outerHashInput := CreateNestedHashInput(toValidatorID, innerHash)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := CachedKeccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot, gasUsed
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
	gasUsed += SloadGasCost // Add gas for SLOAD
	constantsManagerAddr := common.BytesToAddress(constantsManager.Bytes())

	// Pack the function call data
	data, err := CMAbi.Pack(methodName, args...)
	if err != nil {
		log.Error("SFC: Error packing ConstantsManager method", "method", methodName, "err", err)
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Make the call to the ConstantsManager contract
	result, leftOverGas, err := evm.Call(vm.AccountRef(ContractAddress), constantsManagerAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("SFC: Error calling ConstantsManager method", "method", methodName, "err", err, "reason", reason)
		return []interface{}{result}, gasUsed + (defaultGasLimit - leftOverGas), err
	}

	// Add the gas used by the call
	gasUsed += (defaultGasLimit - leftOverGas)

	// Unpack the result
	values, err := CMAbi.Methods[methodName].Outputs.Unpack(result)
	if err != nil {
		log.Error("SFC: Error unpacking ConstantsManager method", "method", methodName, "err", err)
		return values, gasUsed, vm.ErrExecutionReverted
	}

	return values, gasUsed, nil
}

// getCurrentEpoch returns the current epoch value (currentSealedEpoch + 1)
// This implements the logic from the currentEpoch() function in SFCBase.sol
func getCurrentEpoch(evm *vm.EVM) (*big.Int, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the current sealed epoch slot using a cached constant
	currentSealedEpochSlotHash := common.BigToHash(big.NewInt(currentSealedEpochSlot))

	// Get the current sealed epoch (SLOAD operation)
	currentSealedEpoch := evm.SfcStateDB.GetState(ContractAddress, currentSealedEpochSlotHash)
	gasUsed += SloadGasCost // Add gas for SLOAD

	// Use the big.Int pool
	currentSealedEpochBigInt := GetBigInt().SetBytes(currentSealedEpoch.Bytes())

	// Calculate current epoch as currentSealedEpoch + 1 using the pool
	currentEpochBigInt := GetBigInt().Add(currentSealedEpochBigInt, big.NewInt(1))

	// Return the sealed epoch to the pool
	PutBigInt(currentSealedEpochBigInt)

	return currentEpochBigInt, gasUsed, nil
}

// getEpochSnapshotSlot calculates the storage slot for an epoch snapshot
func getEpochSnapshotSlot(epoch *big.Int) (*big.Int, uint64) {
	// Check if the result is in the cache
	key := epoch.String()
	if slot, found := sfcCache.EpochSlot[key]; found {
		return slot, HashGasCost // Still account for gas even though we're using the cache
	}

	// Initialize gas used
	var gasUsed uint64 = 0

	// For a mapping(uint256 => EpochSnapshot), the slot is calculated as:
	// keccak256(abi.encode(epoch, epochSnapshotSlot))

	// Create the hash input: abi.encode(epoch, epochSnapshotSlot)
	epochBytes := common.LeftPadBytes(epoch.Bytes(), 32)                                     // Left-pad to 32 bytes
	epochSnapshotSlotBytes := common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(epochBytes, epochSnapshotSlotBytes...)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost
	log.Trace("SFC: Epoch Snapshot Slot", "hashslot", common.Bytes2Hex(hash))

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	// Store in cache
	sfcCache.EpochSlot[key] = slot

	return slot, gasUsed
}

// getEpochSnapshotFieldSlot calculates the storage slot for a field in an epoch snapshot
func getEpochSnapshotFieldSlot(epoch *big.Int, fieldOffset int64) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the base slot for the epoch snapshot
	baseSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// Add the field offset to the base slot
	fieldSlot := new(big.Int).Add(baseSlot, big.NewInt(fieldOffset))

	return fieldSlot, gasUsed
}

// getOfflinePenaltyThresholdBlocksNum gets the offline penalty threshold blocks number from the constants manager
func getOfflinePenaltyThresholdBlocksNum(evm *vm.EVM) (*big.Int, uint64, error) {
	result, gasUsed, err := callConstantManagerMethod(evm, "offlinePenaltyThresholdBlocksNum")
	if err != nil || len(result) == 0 {
		log.Error("SFC: Error getting offline penalty threshold blocks number from ConstantsManager", "err", err)
		return nil, gasUsed, err
	}

	threshold, ok := result[0].(*big.Int)
	if !ok {
		log.Error("SFC: Error typecasting offline penalty threshold blocks number from ConstantsManager")
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return threshold, gasUsed, nil
}

// getOfflinePenaltyThresholdTime gets the offline penalty threshold time from the constants manager
func getOfflinePenaltyThresholdTime(evm *vm.EVM) (*big.Int, uint64, error) {
	result, gasUsed, err := callConstantManagerMethod(evm, "offlinePenaltyThresholdTime")
	if err != nil || len(result) == 0 {
		return nil, gasUsed, err
	}

	threshold, ok := result[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return threshold, gasUsed, nil
}

// getValidatorStatusSlotByID calculates the storage slot for a validator's status by ID
func getValidatorStatusSlotByID(validatorID *big.Int) (*big.Int, uint64) {
	return getValidatorStatusSlot(validatorID)
}

// _setValidatorDeactivated sets a validator as deactivated with the specified status bit
func _setValidatorDeactivated(evm *vm.EVM, validatorID *big.Int, statusBit uint64) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the validator status slot
	statusSlot, slotGasUsed := getValidatorStatusSlotByID(validatorID)
	gasUsed += slotGasUsed

	// Get the current status
	status := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(statusSlot))
	gasUsed += SloadGasCost
	statusBigInt := new(big.Int).SetBytes(status.Bytes())

	// Check if the validator is already deactivated with this status
	currentStatus := statusBigInt.Uint64()
	if currentStatus == OK_STATUS && statusBit != OK_STATUS {
		// Get the validator's received stake slot
		receivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed

		// Get the validator's received stake
		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(receivedStakeSlot))
		gasUsed += SloadGasCost
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		// Update totalActiveStake
		totalActiveStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		gasUsed += SloadGasCost
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStake.Bytes())

		// Subtract the validator's stake from totalActiveStake
		newTotalActiveStake := new(big.Int).Sub(totalActiveStakeBigInt, receivedStakeBigInt)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
		gasUsed += SstoreGasCost
	}

	// Status as a number is proportional to severity
	if statusBit > currentStatus {
		// Update the validator status
		newStatus := new(big.Int).SetUint64(statusBit)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(statusSlot), common.BigToHash(newStatus))
		gasUsed += SstoreGasCost

		// Check if the validator is not already deactivated
		deactivatedEpochSlot, slotGasUsed := getValidatorDeactivatedEpochSlot(validatorID)
		gasUsed += slotGasUsed

		deactivatedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(deactivatedEpochSlot))
		gasUsed += SloadGasCost
		deactivatedEpochBigInt := new(big.Int).SetBytes(deactivatedEpoch.Bytes())

		if deactivatedEpochBigInt.Cmp(big.NewInt(0)) == 0 {
			// Set the deactivated epoch to the current epoch
			currentEpochBigInt, epochGasUsed, err := getCurrentEpoch(evm)
			gasUsed += epochGasUsed
			if err != nil {
				return gasUsed, err
			}

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(deactivatedEpochSlot), common.BigToHash(currentEpochBigInt))
			gasUsed += SstoreGasCost

			// Set the deactivated time to the current time
			deactivatedTimeSlot, slotGasUsed := getValidatorDeactivatedTimeSlot(validatorID)
			gasUsed += slotGasUsed

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(deactivatedTimeSlot), common.BigToHash(evm.Context.Time))
			gasUsed += SstoreGasCost

			// Emit DeactivatedValidator event
			topics := []common.Hash{
				SfcAbi.Events["DeactivatedValidator"].ID,
				common.BigToHash(validatorID), // indexed parameter (validatorID)
			}
			data := common.BigToHash(currentEpochBigInt).Bytes()
			data = append(data, common.BigToHash(evm.Context.Time).Bytes()...)

			evm.SfcStateDB.AddLog(&types.Log{
				Address:     ContractAddress,
				Topics:      topics,
				Data:        data,
				BlockNumber: evm.Context.BlockNumber.Uint64(),
			})
		}

		// Emit ChangedValidatorStatus event
		topics := []common.Hash{
			SfcAbi.Events["ChangedValidatorStatus"].ID,
			common.BigToHash(validatorID), // indexed parameter (validatorID)
		}
		data := common.BigToHash(new(big.Int).SetUint64(statusBit)).Bytes()

		evm.SfcStateDB.AddLog(&types.Log{
			Address:     ContractAddress,
			Topics:      topics,
			Data:        data,
			BlockNumber: evm.Context.BlockNumber.Uint64(),
		})
	}

	return gasUsed, nil
}

// _syncValidator syncs a validator's weight with the node
func _syncValidator(evm *vm.EVM, validatorID *big.Int, syncPubkey bool) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Check if validator exists
	validatorCreatedTimeSlot, slotGasUsed := getValidatorCreatedTimeSlot(validatorID)
	gasUsed += slotGasUsed

	validatorCreatedTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorCreatedTimeSlot))
	gasUsed += SloadGasCost

	if validatorCreatedTime.Big().Cmp(big.NewInt(0)) == 0 {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Get the validator's received stake
	receivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
	gasUsed += slotGasUsed

	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(receivedStakeSlot))
	gasUsed += SloadGasCost
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

	// Get the validator's status
	statusSlot, slotGasUsed := getValidatorStatusSlotByID(validatorID)
	gasUsed += slotGasUsed

	status := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(statusSlot))
	gasUsed += SloadGasCost
	statusBigInt := new(big.Int).SetBytes(status.Bytes())

	// Calculate weight
	weight := receivedStakeBigInt
	if statusBigInt.Uint64() != OK_STATUS {
		weight = big.NewInt(0)
	}

	// Call the node to update validator weight
	nodeDriverAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(nodeDriverAuthSlot)))
	gasUsed += SloadGasCost
	nodeDriverAuthAddr := common.BytesToAddress(nodeDriverAuth.Bytes())

	// Pack the function call data
	data, err := NodeDriverAbi.Pack("updateValidatorWeight", validatorID, weight)
	if err != nil {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Call the node driver
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return gasUsed, err
	}

	// If syncPubkey is true and weight is not zero, update validator pubkey
	if syncPubkey && weight.Cmp(big.NewInt(0)) != 0 {
		// Get the validator's pubkey
		pubkeySlot, slotGasUsed := getValidatorPubkeySlot(validatorID)
		gasUsed += slotGasUsed

		pubkeyHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(pubkeySlot))
		gasUsed += SloadGasCost

		// Pack the function call data
		data, err := NodeDriverAbi.Pack("updateValidatorPubkey", validatorID, pubkeyHash.Bytes())
		if err != nil {
			return gasUsed, vm.ErrExecutionReverted
		}

		// Call the node driver
		_, _, err = evm.Call(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, defaultGasLimit, big.NewInt(0))
		if err != nil {
			return gasUsed, err
		}
	}

	return gasUsed, nil
}

// getEpochValidatorOfflineTimeSlot calculates the storage slot for a validator's offline time in an epoch
func getEpochValidatorOfflineTimeSlot(epoch *big.Int, validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping field inside a struct, the correct approach is:
	// 1. Get the base slot of the struct
	// 2. Add the offset of the mapping field to get the mapping's slot
	// 3. Calculate the final slot for a specific key as keccak256(key . (struct_slot + offset))

	// Add the offset for the offlineTime mapping within the struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(offlineTimeOffset))

	// Use our helper function to create and hash the input
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
}

// getEpochValidatorOfflineBlocksSlot calculates the storage slot for a validator's offline blocks in an epoch
func getEpochValidatorOfflineBlocksSlot(epoch *big.Int, validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping field inside a struct, the correct approach is:
	// 1. Get the base slot of the struct
	// 2. Add the offset of the mapping field to get the mapping's slot
	// 3. Calculate the final slot for a specific key as keccak256(key . (struct_slot + offset))

	// Add the offset for the offlineBlocks mapping within the struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(offlineBlocksOffset))

	// Use our helper function to create and hash the input
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
}

// getEpochValidatorAccumulatedRewardPerTokenSlot calculates the storage slot for a validator's accumulated reward per token in an epoch
func getEpochValidatorAccumulatedRewardPerTokenSlot(epoch *big.Int, validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping field inside a struct, the correct approach is:
	// 1. Get the base slot of the struct
	// 2. Add the offset of the mapping field to get the mapping's slot
	// 3. Calculate the final slot for a specific key as keccak256(key . (struct_slot + offset))

	// Add the offset for the accumulatedRewardPerToken mapping within the struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(accumulatedRewardPerTokenOffset))

	// Use our helper function to create and hash the input
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
}

// getEpochValidatorAccumulatedUptimeSlot calculates the storage slot for a validator's accumulated uptime in an epoch
func getEpochValidatorAccumulatedUptimeSlot(epoch *big.Int, validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping field inside a struct, the correct approach is:
	// 1. Get the base slot of the struct
	// 2. Add the offset of the mapping field to get the mapping's slot
	// 3. Calculate the final slot for a specific key as keccak256(key . (struct_slot + offset))

	// Add the offset for the accumulatedUptime mapping within the struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(accumulatedUptimeOffset))

	// Use our helper function to create and hash the input
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
}

// getEpochValidatorAccumulatedOriginatedTxsFeeSlot calculates the storage slot for a validator's accumulated originated txs fee in an epoch
func getEpochValidatorAccumulatedOriginatedTxsFeeSlot(epoch *big.Int, validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping field inside a struct, the correct approach is:
	// 1. Get the base slot of the struct
	// 2. Add the offset of the mapping field to get the mapping's slot
	// 3. Calculate the final slot for a specific key as keccak256(key . (struct_slot + offset))

	// Add the offset for the accumulatedOriginatedTxsFee mapping within the struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(accumulatedOriginatedTxsFeeOffset))

	// Use our helper function to create and hash the input
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
}

// getEpochValidatorReceivedStakeSlot calculates the storage slot for a validator's received stake in an epoch
func getEpochValidatorReceivedStakeSlot(epoch *big.Int, validatorID *big.Int) (*big.Int, uint64) {
	var gasUsed uint64 = 0

	// Step 1: Calculate the slot for getEpochSnapshot[epoch]
	hashInput := CreateHashInput(epoch, epochSnapshotSlot)
	epochHash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Step 2: Add the offset for receivedStake within the struct
	// receivedStake is the first mapping in the struct, so its offset is 0
	structSlot := new(big.Int).SetBytes(epochHash)
	mappingSlot := new(big.Int).Add(structSlot, big.NewInt(receiveStakeOffset))

	// Step 3: Calculate the final slot for receivedStake[validatorID]
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)
	mappingSlotBytes := common.LeftPadBytes(mappingSlot.Bytes(), 32) // This is a computed value, not a constant

	// Use the byte slice pool for the result
	finalHashInput := GetByteSlice()
	if cap(finalHashInput) < len(validatorIDBytes)+len(mappingSlotBytes) {
		// If the slice from the pool is too small, allocate a new one
		finalHashInput = make([]byte, 0, len(validatorIDBytes)+len(mappingSlotBytes))
	}

	// Combine the bytes
	finalHashInput = append(finalHashInput, validatorIDBytes...)
	finalHashInput = append(finalHashInput, mappingSlotBytes...)
	finalHash := CachedKeccak256(finalHashInput)
	gasUsed += HashGasCost

	return new(big.Int).SetBytes(finalHash), gasUsed
}

// getDecimalUnit returns the decimal unit value (1e18) used for decimal calculations
func getDecimalUnit() *big.Int {
	// This is a pure function that returns 1e18, as defined in Decimal.sol
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
}

// trimGasPriceChangeRatio limits the gas price change ratio to be between 95% and 105%
// This is a port of the GP.trimGasPriceChangeRatio function from GasPriceConstants.sol
func trimGasPriceChangeRatio(x *big.Int) *big.Int {
	// Get the decimal unit
	unit := getDecimalUnit()

	// Calculate 105% of unit
	maxRatio := new(big.Int).Mul(unit, big.NewInt(105))
	maxRatio = new(big.Int).Div(maxRatio, big.NewInt(100))

	// Calculate 95% of unit
	minRatio := new(big.Int).Mul(unit, big.NewInt(95))
	minRatio = new(big.Int).Div(minRatio, big.NewInt(100))

	// Check if x is greater than 105% of unit
	if x.Cmp(maxRatio) > 0 {
		return maxRatio
	}

	// Check if x is less than 95% of unit
	if x.Cmp(minRatio) < 0 {
		return minRatio
	}

	// Return x unchanged if it's within the allowed range
	return x
}

// trimMinGasPrice limits the minimum gas price to be between 1 gwei and 1000000 gwei
// This is a port of the GP.trimMinGasPrice function from GasPriceConstants.sol
func trimMinGasPrice(x *big.Int) *big.Int {
	// Calculate 1 gwei (1e9)
	minGasPrice := new(big.Int).Mul(big.NewInt(1), big.NewInt(1000000000))

	// Calculate 1000000 gwei (1e15)
	maxGasPrice := new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1000000000))

	// Check if x is greater than 1000000 gwei
	if x.Cmp(maxGasPrice) > 0 {
		return maxGasPrice
	}

	// Check if x is less than 1 gwei
	if x.Cmp(minGasPrice) < 0 {
		return minGasPrice
	}

	// Return x unchanged if it's within the allowed range
	return x
}

// sumRewards adds three Rewards structs together and returns the result
// This is a port of the sumRewards function from SFCBase.sol
func sumRewards(a Rewards, b Rewards, c Rewards) Rewards {
	return Rewards{
		LockupExtraReward: new(big.Int).Add(new(big.Int).Add(a.LockupExtraReward, b.LockupExtraReward), c.LockupExtraReward),
		LockupBaseReward:  new(big.Int).Add(new(big.Int).Add(a.LockupBaseReward, b.LockupBaseReward), c.LockupBaseReward),
		UnlockedReward:    new(big.Int).Add(new(big.Int).Add(a.UnlockedReward, b.UnlockedReward), c.UnlockedReward),
	}
}

// _mintNativeToken mints native tokens to the specified address
// This is a helper function to call the _mintNativeToken handler
func _mintNativeToken(evm *vm.EVM, receiver common.Address, amount *big.Int) (uint64, error) {
	// Pack the arguments
	args := []interface{}{
		receiver,
		amount,
	}

	// Call the _mintNativeToken handler
	_, gasUsed, err := handle_mintNativeToken(evm, args)
	if err != nil {
		return gasUsed, err
	}

	return gasUsed, nil
}

// Rewards struct represents the different types of rewards
type Rewards struct {
	LockupExtraReward *big.Int
	LockupBaseReward  *big.Int
	UnlockedReward    *big.Int
}

// _scaleLockupReward scales the reward based on the lockup duration
// This is a port of the _scaleLockupReward function from SFCBase.sol
func _scaleLockupReward(evm *vm.EVM, fullReward *big.Int, lockupDuration *big.Int) (Rewards, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Initialize reward with zeros
	reward := Rewards{
		LockupExtraReward: big.NewInt(0),
		LockupBaseReward:  big.NewInt(0),
		UnlockedReward:    big.NewInt(0),
	}

	// Get the unlockedRewardRatio from the constants manager
	unlockedRewardRatio, cmGasUsed, err := callConstantManagerMethod(evm, "unlockedRewardRatio")
	gasUsed += cmGasUsed
	if err != nil || len(unlockedRewardRatio) == 0 {
		return reward, gasUsed, err
	}
	unlockedRewardRatioBigInt, ok := unlockedRewardRatio[0].(*big.Int)
	if !ok {
		return reward, gasUsed, vm.ErrExecutionReverted
	}

	// Check if lockupDuration is not zero
	if lockupDuration.Cmp(big.NewInt(0)) != 0 {
		// Get the decimal unit
		decimalUnit := getDecimalUnit()

		// Calculate maxLockupExtraRatio = Decimal.unit() - unlockedRewardRatio
		maxLockupExtraRatio := new(big.Int).Sub(decimalUnit, unlockedRewardRatioBigInt)

		// Get the maxLockupDuration from the constants manager
		maxLockupDuration, cmGasUsed, err := callConstantManagerMethod(evm, "maxLockupDuration")
		gasUsed += cmGasUsed
		if err != nil || len(maxLockupDuration) == 0 {
			return reward, gasUsed, err
		}
		maxLockupDurationBigInt, ok := maxLockupDuration[0].(*big.Int)
		if !ok {
			return reward, gasUsed, vm.ErrExecutionReverted
		}

		// Calculate lockupExtraRatio = maxLockupExtraRatio * lockupDuration / maxLockupDuration
		lockupExtraRatio := new(big.Int).Mul(maxLockupExtraRatio, lockupDuration)
		lockupExtraRatio = new(big.Int).Div(lockupExtraRatio, maxLockupDurationBigInt)

		// Calculate totalScaledReward = fullReward * (unlockedRewardRatio + lockupExtraRatio) / Decimal.unit()
		totalRewardRatio := new(big.Int).Add(unlockedRewardRatioBigInt, lockupExtraRatio)
		totalScaledReward := new(big.Int).Mul(fullReward, totalRewardRatio)
		totalScaledReward = new(big.Int).Div(totalScaledReward, decimalUnit)

		// Calculate lockupBaseReward = fullReward * unlockedRewardRatio / Decimal.unit()
		reward.LockupBaseReward = new(big.Int).Mul(fullReward, unlockedRewardRatioBigInt)
		reward.LockupBaseReward = new(big.Int).Div(reward.LockupBaseReward, decimalUnit)

		// Calculate lockupExtraReward = totalScaledReward - lockupBaseReward
		reward.LockupExtraReward = new(big.Int).Sub(totalScaledReward, reward.LockupBaseReward)
	} else {
		// Calculate unlockedReward = fullReward * unlockedRewardRatio / Decimal.unit()
		reward.UnlockedReward = new(big.Int).Mul(fullReward, unlockedRewardRatioBigInt)
		reward.UnlockedReward = new(big.Int).Div(reward.UnlockedReward, getDecimalUnit())
	}

	return reward, gasUsed, nil
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
