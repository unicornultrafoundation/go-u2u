package sfc

import (
	"fmt"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/constant_manager"
)

// Gas costs and limits
const (
	defaultGasLimit uint64 = 7000000 // Default gas limit for contract calls

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

// Cached values for gas price ratio limits
var (
	// gasMaxRatio represents 105% of the decimal unit (1.05e18)
	gasMaxRatio *big.Int
	// gasMinRatio represents 95% of the decimal unit (0.95e18)
	gasMinRatio *big.Int

	unit             = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	constMinGasPrice = new(big.Int).Mul(big.NewInt(1), big.NewInt(1000000000))
	constMaxGasPrice = new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1000000000))

	SfcPrecompiles = map[common.Address]bool{
		common.HexToAddress("0xFC00FACE00000000000000000000000000000000"): true,
		common.HexToAddress("0xD100ae0000000000000000000000000000000000"): true,
		common.HexToAddress("0xd100A01E00000000000000000000000000000000"): true,
		common.HexToAddress("0x6CA548f6DF5B540E72262E935b6Fe3e72cDd68C9"): true,
		common.HexToAddress("0xFC01fACE00000000000000000000000000000000"): true, // SFCLib
	}
)

// init calculates and caches the gas price ratio limits
func init() {
	// Calculate 105% of unit
	gasMaxRatio = new(big.Int).Mul(unit, big.NewInt(105))
	gasMaxRatio = new(big.Int).Div(gasMaxRatio, big.NewInt(100))

	// Calculate 95% of unit
	gasMinRatio = new(big.Int).Mul(unit, big.NewInt(95))
	gasMinRatio = new(big.Int).Div(gasMinRatio, big.NewInt(100))
}

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
	if caller.Cmp(NodeDriverAuthAddr) != 0 {
		// Return ABI-encoded revert reason: "caller is not the NodeDriverAuth contract"
		revertData, err := encodeRevertReason(methodName, "caller is not the NodeDriverAuth contract")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		log.Error("SFC: Caller is not the NodeDriverAuth contract",
			"caller", caller.Hex(), "nodeDriverAuth", NodeDriverAuthAddr.Hex())
		return revertData, gasUsed, vm.ErrExecutionReverted
	}
	return nil, gasUsed, nil
}

// checkValidatorExists checks if a validator with the given ID exists
// Returns nil if the validator exists, otherwise returns an ABI-encoded revert reason
func checkValidatorExists(evm *vm.EVM, validatorID *big.Int, methodName string) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Calculate validator createdTime slot - this matches the _validatorExists function in SFCBase.sol
	// which checks if getValidator[validatorID].createdTime != 0
	createdTimeSlot, slotGasUsed := getValidatorCreatedTimeSlot(validatorID)
	gasUsed += slotGasUsed

	// Check if validator exists (SLOAD operation)
	validatorCreatedTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(createdTimeSlot))
	gasUsed += SloadGasCost

	// Check if createdTime is zero
	if validatorCreatedTime.Big().Cmp(big.NewInt(0)) == 0 {
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

	// Check if validator is active (status == OK_STATUS) (SLOAD operation)
	// This matches the Solidity: require(getValidator[toValidatorID].status == OK_STATUS, "validator isn't active");
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(statusSlot))
	gasUsed += SloadGasCost

	statusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if validator status is exactly OK_STATUS (0) to match Solidity behavior
	if statusBigInt.Cmp(big.NewInt(int64(OK_STATUS))) != 0 {
		// Return ABI-encoded revert reason: "validator isn't active"
		revertReason := "validator isn't active"
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
	if cachedValue, ok := sfcCache.AbiPackCache.Get(cacheKey); ok {
		log.Info("SFC: Revert", "message", errorMessage)
		return cachedValue.([]byte), nil
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
	sfcCache.AbiPackCache.Add(cacheKey, result)
	log.Info("SFC: Revert", "message", errorMessage)
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
	var gasUsed uint64 = 0
	hashInput := CreateHashInput(validatorID, validatorSlot)
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost
	slot := new(big.Int).SetBytes(hash)

	// The auth field is at slot + 6
	return new(big.Int).Add(slot, big.NewInt(6)), gasUsed
}

// getValidatorPubkeySlot calculates the storage slot for a validator's pubkey
func getValidatorPubkeySlot(validatorID *big.Int) (*big.Int, uint64) {
	var gasUsed uint64 = 0
	hashInput := CreateHashInput(validatorID, validatorPubkeySlot)
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost
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

	// Convert the hash to a big.Int using the pool
	slot := GetBigInt().SetBytes(hash)

	// The receivedStake field is at slot + 3
	result := GetBigInt().Add(slot, big.NewInt(3))

	// Return the slot to the pool
	PutBigInt(slot)

	return result, gasUsed
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

	// The WithdrawalRequest struct has the following fields (in Solidity storage order):
	// uint256 epoch;   // First field (slot + 0)
	// uint256 time;    // Second field (slot + 1)
	// uint256 amount;  // Third field (slot + 2)

	// We're returning the base slot for the struct
	return slot, gasUsed
}

// getWithdrawalRequestAmountSlot calculates the storage slot for a withdrawal request amount
func getWithdrawalRequestAmountSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The amount field is at slot + 2 (third field in the struct)
	return new(big.Int).Add(baseSlot, big.NewInt(2)), gasUsed
}

// getWithdrawalRequestEpochSlot calculates the storage slot for a withdrawal request epoch
func getWithdrawalRequestEpochSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The epoch field is at slot + 0 (first field in the struct)
	return baseSlot, gasUsed
}

// getWithdrawalRequestTimeSlot calculates the storage slot for a withdrawal request time
func getWithdrawalRequestTimeSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The time field is at slot + 1 (second field in the struct)
	return new(big.Int).Add(baseSlot, big.NewInt(1)), gasUsed
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
	// For a mapping(address => mapping(uint256 => Rewards)), first we need to get the slot for the Rewards struct
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

// getStashedLockupExtraRewardSlot calculates the storage slot for a delegation's stashed lockup extra reward
func getStashedLockupExtraRewardSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the Rewards struct
	baseSlot, gasUsed := getStashedLockupRewardsSlot(delegator, toValidatorID)
	// The lockupExtraReward field is at the base slot
	return baseSlot, gasUsed
}

// getStashedLockupBaseRewardSlot calculates the storage slot for a delegation's stashed lockup base reward
func getStashedLockupBaseRewardSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the Rewards struct
	baseSlot, gasUsed := getStashedLockupRewardsSlot(delegator, toValidatorID)
	// The lockupBaseReward field is at baseSlot + 1
	return new(big.Int).Add(baseSlot, big.NewInt(1)), gasUsed
}

// getStashedUnlockedRewardSlot calculates the storage slot for a delegation's stashed unlocked reward
func getStashedUnlockedRewardSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the Rewards struct
	baseSlot, gasUsed := getStashedLockupRewardsSlot(delegator, toValidatorID)
	// The unlockedReward field is at baseSlot + 2
	return new(big.Int).Add(baseSlot, big.NewInt(2)), gasUsed
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

// getRewardsStashLockupExtraRewardSlot calculates the storage slot for a delegation's rewards stash lockup extra reward
func getRewardsStashLockupExtraRewardSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the Rewards struct
	baseSlot, gasUsed := getRewardsStashSlot(delegator, toValidatorID)
	// The lockupExtraReward field is at baseSlot + 0
	return baseSlot, gasUsed
}

// getRewardsStashLockupBaseRewardSlot calculates the storage slot for a delegation's rewards stash lockup base reward
func getRewardsStashLockupBaseRewardSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the Rewards struct
	baseSlot, gasUsed := getRewardsStashSlot(delegator, toValidatorID)
	// The lockupBaseReward field is at baseSlot + 1
	return new(big.Int).Add(baseSlot, big.NewInt(1)), gasUsed
}

// getRewardsStashUnlockedRewardSlot calculates the storage slot for a delegation's rewards stash unlocked reward
func getRewardsStashUnlockedRewardSlot(delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64) {
	// Get the base slot for the Rewards struct
	baseSlot, gasUsed := getRewardsStashSlot(delegator, toValidatorID)
	// The unlockedReward field is at baseSlot + 2
	return new(big.Int).Add(baseSlot, big.NewInt(2)), gasUsed
}

// callConstantManagerMethod calls a method on the ConstantManager contract and returns the result
// methodName: the name of the method to call
// args: the arguments to pass to the method
// Returns: the result of the method call, the gas used, or an error if the call failed
func callConstantManagerMethod(evm *vm.EVM, methodName string, args ...interface{}) ([]interface{}, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Pack the function call data
	data, err := CMAbi.Pack(methodName, args...)
	if err != nil {
		log.Error("SFC: Error packing ConstantsManager method", "method", methodName, "err", err)
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Make the call to the ConstantsManager contract
	result, leftOverGas, err := evm.CallSFC(vm.AccountRef(ContractAddress), CMAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("SFC: Error calling ConstantsManager method", "method", methodName, "err", err, "reason", reason)
		return []interface{}{result}, gasUsed + (defaultGasLimit - leftOverGas), err
	}

	// Add the gas used by the call
	gasUsed += defaultGasLimit - leftOverGas

	// Unpack the result
	values, err := CMAbi.Methods[methodName].Outputs.Unpack(result)
	if err != nil {
		log.Error("SFC: Error unpacking ConstantsManager method", "method", methodName, "err", err)
		return values, gasUsed, vm.ErrExecutionReverted
	}

	return values, gasUsed, nil
}

func getConstantsManagerVariable(stateDb vm.StateDB, methodName string) *big.Int {
	return stateDb.GetState(CMAddr, common.BigToHash(constant_manager.ConstantManagerSlots[methodName])).Big()
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
	currentEpochBigInt := new(big.Int).Add(currentSealedEpochBigInt, big.NewInt(1))

	// Return the sealed epoch to the pool
	PutBigInt(currentSealedEpochBigInt)

	return currentEpochBigInt, gasUsed, nil
}

// getEpochSnapshotSlot calculates the storage slot for an epoch snapshot
func getEpochSnapshotSlot(epoch *big.Int) (*big.Int, uint64) {
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

	return slot, gasUsed
}

// getValidatorStatusSlotByID calculates the storage slot for a validator's status by ID
func getValidatorStatusSlotByID(validatorID *big.Int) (*big.Int, uint64) {
	return getValidatorStatusSlot(validatorID)
}

// handleInternalSetValidatorDeactivated sets a validator as deactivated with the specified status bit
func handleInternalSetValidatorDeactivated(evm *vm.EVM, validatorID *big.Int, statusBit uint64) (uint64, error) {
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
				BlockNumber: evm.Context.BlockNumber.Uint64(),
				Address:     ContractAddress,
				Topics:      topics,
				Data:        data,
			})
		}

		// Emit ChangedValidatorStatus event
		topics := []common.Hash{
			SfcAbi.Events["ChangedValidatorStatus"].ID,
			common.BigToHash(validatorID), // indexed parameter (validatorID)
		}
		data := common.BigToHash(new(big.Int).SetUint64(statusBit)).Bytes()

		evm.SfcStateDB.AddLog(&types.Log{
			BlockNumber: evm.Context.BlockNumber.Uint64(),
			Address:     ContractAddress,
			Topics:      topics,
			Data:        data,
		})
	}

	return gasUsed, nil
}

// handleInternalSyncValidator syncs a validator's weight with the node
func handleInternalSyncValidator(evm *vm.EVM, validatorID *big.Int, syncPubkey bool) (uint64, error) {
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
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return gasUsed, err
	}

	// If syncPubkey is true and weight is not zero, update validator pubkey
	if syncPubkey && weight.Cmp(big.NewInt(0)) != 0 {
		// Get the validator's pubkey (dynamic bytes)
		pubkeySlot, slotGasUsed := getValidatorPubkeySlot(validatorID)
		gasUsed += slotGasUsed

		pubkeyBytes, readBytesGasUsed, err := readDynamicBytes(evm, pubkeySlot)
		if err != nil {
			return gasUsed, err
		}
		gasUsed += readBytesGasUsed

		// Pack the function call data
		data, err := NodeDriverAbi.Pack("updateValidatorPubkey", validatorID, pubkeyBytes)
		if err != nil {
			return gasUsed, vm.ErrExecutionReverted
		}

		// Call the node driver
		_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, defaultGasLimit, big.NewInt(0))
		if err != nil {
			return gasUsed, err
		}
	}

	return gasUsed, nil
}

// getEpochValidatorReceivedStakeSlot calculates the storage slot for a validator's received stake in an epoch
func getEpochValidatorReceivedStakeSlot(epoch *big.Int, validatorID *big.Int) (*big.Int, uint64) {
	var gasUsed uint64 = 0

	// Step 1: Calculate the slot for getEpochSnapshot[epoch]
	hashInput := CreateHashInput(epoch, epochSnapshotSlot)
	epochHash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

	// Step 2: Add the offset for receivedStake within the struct
	// receivedStake is the first mapping in the struct, so its offset is 0
	structSlot := GetBigInt().SetBytes(epochHash)
	mappingSlot := GetBigInt().Add(structSlot, big.NewInt(receiveStakeOffset))

	// Step 3: Calculate the final slot for receivedStake[validatorID]
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)
	mappingSlotBytes := common.LeftPadBytes(mappingSlot.Bytes(), 32) // This is a computed value, not a constant

	// Use the byte slice pool for the result
	finalHashInput := GetByteSlice()
	// Reset the slice but keep the capacity
	finalHashInput = finalHashInput[:0]
	// Ensure the slice has enough capacity
	if cap(finalHashInput) < len(validatorIDBytes)+len(mappingSlotBytes) {
		// If the slice from the pool is too small, allocate a new one with extra capacity for future use
		finalHashInput = make([]byte, 0, len(validatorIDBytes)+len(mappingSlotBytes)+32)
	}

	// Combine the bytes
	finalHashInput = append(finalHashInput, validatorIDBytes...)
	finalHashInput = append(finalHashInput, mappingSlotBytes...)
	finalHash := crypto.Keccak256(finalHashInput)
	gasUsed += HashGasCost

	// Don't forget to return the byte slice to the pool
	PutByteSlice(finalHashInput)

	result := new(big.Int).SetBytes(finalHash)

	// Return the temporary big.Ints to the pool
	PutBigInt(structSlot)
	PutBigInt(mappingSlot)

	return result, gasUsed
}

// getDecimalUnit returns the decimal unit value (1e18) used for decimal calculations
func getDecimalUnit() *big.Int {
	// This is a pure function that returns 1e18, as defined in Decimal.sol
	return unit
}

// trimGasPriceChangeRatio limits the gas price change ratio to be between 95% and 105%
// This is a port of the GP.trimGasPriceChangeRatio function from GasPriceConstants.sol
func trimGasPriceChangeRatio(x *big.Int) *big.Int {
	// Check if x is greater than 105% of unit
	if x.Cmp(gasMaxRatio) > 0 {
		return gasMaxRatio
	}

	// Check if x is less than 95% of unit
	if x.Cmp(gasMinRatio) < 0 {
		return gasMinRatio
	}

	// Return x unchanged if it's within the allowed range
	return x
}

// trimMinGasPrice limits the minimum gas price to be between 1 gwei and 1000000 gwei
// This is a port of the GP.trimMinGasPrice function from GasPriceConstants.sol
func trimMinGasPrice(x *big.Int) *big.Int {
	// Check if x is greater than 1000000 gwei
	if x.Cmp(constMaxGasPrice) > 0 {
		return constMaxGasPrice
	}

	// Check if x is less than 1 gwei
	if x.Cmp(constMinGasPrice) < 0 {
		return constMinGasPrice
	}

	// Return x unchanged if it's within the allowed range
	return x
}

// sumRewards adds two Rewards structs together and returns the result
// This is a port of the sumRewards function from SFCBase.sol
func sumRewards(a Rewards, b Rewards) Rewards {
	return Rewards{
		LockupExtraReward: new(big.Int).Add(a.LockupExtraReward, b.LockupExtraReward),
		LockupBaseReward:  new(big.Int).Add(a.LockupBaseReward, b.LockupBaseReward),
		UnlockedReward:    new(big.Int).Add(a.UnlockedReward, b.UnlockedReward),
	}
}

// sumRewards3 adds three Rewards structs together and returns the result
// This is used when we need to sum three rewards (like in _newRewards)
func sumRewards3(a Rewards, b Rewards, c Rewards) Rewards {
	return sumRewards(sumRewards(a, b), c)
}

// getSlashingPenalty calculates the slashing penalty for a validator
// This is a port of the getSlashingPenalty function from SFCLib.sol
func getSlashingPenalty(evm *vm.EVM, amount *big.Int, isCheater bool, validatorID *big.Int) (*big.Int, uint64, error) {
	var gasUsed uint64 = 0

	// If not a cheater, no penalty
	if !isCheater {
		return big.NewInt(0), gasUsed, nil
	}

	// Get the slashing refund ratio for this validator
	slashingRefundRatioSlot, slotGasUsed := getSlashingRefundRatioSlot(validatorID)
	gasUsed += slotGasUsed
	refundRatioState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slashingRefundRatioSlot))
	gasUsed += SloadGasCost
	refundRatio := new(big.Int).SetBytes(refundRatioState.Bytes())

	// Get decimal unit
	decimalUnit := getDecimalUnit()

	// If refundRatio >= Decimal.unit(), no penalty
	if refundRatio.Cmp(decimalUnit) >= 0 {
		return big.NewInt(0), gasUsed, nil
	}

	// Calculate penalty = amount * (Decimal.unit() - refundRatio) / Decimal.unit() + 1
	// Round penalty upwards (ceiling) to prevent dust amount attacks
	numerator := new(big.Int).Sub(decimalUnit, refundRatio)
	penalty := new(big.Int).Mul(amount, numerator)
	penalty = new(big.Int).Div(penalty, decimalUnit)
	penalty = new(big.Int).Add(penalty, big.NewInt(1))

	// If penalty > amount, return amount
	if penalty.Cmp(amount) > 0 {
		return amount, gasUsed, nil
	}

	return penalty, gasUsed, nil
}

// getSlashingRefundRatioSlot calculates the storage slot for a validator's slashing refund ratio
func getSlashingRefundRatioSlot(validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(uint256 => uint256), the slot is calculated as:
	// keccak256(abi.encode(validatorID, slashingRefundRatioSlot))

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, slashingRefundRatioSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot, gasUsed
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

func (r Rewards) String() string {
	return fmt.Sprintf("LockupExtraReward: %s, LockupBaseReward: %s, UnlockedReward: %s",
		common.Bytes2Hex(r.LockupExtraReward.Bytes()),
		common.Bytes2Hex(r.LockupBaseReward.Bytes()),
		common.Bytes2Hex(r.UnlockedReward.Bytes()))
}

// _highestPayableEpoch returns the highest epoch for which rewards can be paid
// This is a port of the _highestPayableEpoch function from SFCLib.sol
func _highestPayableEpoch(evm *vm.EVM, validatorID *big.Int) (*big.Int, uint64, error) {
	var gasUsed uint64 = 0

	// Get the validator's deactivated epoch
	validatorDeactivatedEpochSlot, slotGasUsed := getValidatorDeactivatedEpochSlot(validatorID)
	gasUsed += slotGasUsed
	validatorDeactivatedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorDeactivatedEpochSlot))
	gasUsed += SloadGasCost
	validatorDeactivatedEpochBigInt := new(big.Int).SetBytes(validatorDeactivatedEpoch.Bytes())

	// Get the current sealed epoch
	currentSealedEpochSlot := common.BigToHash(big.NewInt(currentSealedEpochSlot))
	currentSealedEpoch := evm.SfcStateDB.GetState(ContractAddress, currentSealedEpochSlot)
	gasUsed += SloadGasCost
	currentSealedEpochBigInt := new(big.Int).SetBytes(currentSealedEpoch.Bytes())

	// If the validator is deactivated (deactivatedEpoch != 0)
	if validatorDeactivatedEpochBigInt.Cmp(big.NewInt(0)) != 0 {
		// If currentSealedEpoch < deactivatedEpoch, return currentSealedEpoch
		// Otherwise return deactivatedEpoch
		if currentSealedEpochBigInt.Cmp(validatorDeactivatedEpochBigInt) < 0 {
			return currentSealedEpochBigInt, gasUsed, nil
		}
		return validatorDeactivatedEpochBigInt, gasUsed, nil
	}

	// If validator is not deactivated, return currentSealedEpoch
	return currentSealedEpochBigInt, gasUsed, nil
}

// _isLockedUpAtEpoch checks if a delegation is locked up at a specific epoch
// This is a port of the _isLockedUpAtEpoch function from SFCLib.sol
func _isLockedUpAtEpoch(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, epoch *big.Int) (bool, uint64, error) {
	var gasUsed uint64 = 0

	// Get the lockup from epoch
	lockupFromEpochSlot, slotGasUsed := getLockupFromEpochSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	lockupFromEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupFromEpochSlot))
	gasUsed += SloadGasCost
	lockupFromEpochBigInt := new(big.Int).SetBytes(lockupFromEpoch.Bytes())

	// Get the lockup end time
	lockupEndTimeSlot, slotGasUsed := getLockupEndTimeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	lockupEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupEndTimeSlot))
	gasUsed += SloadGasCost
	lockupEndTimeBigInt := new(big.Int).SetBytes(lockupEndTime.Bytes())

	// Get the epoch end time
	epochEndTimeSlot, slotGasUsed := getEpochEndTimeSlot(epoch)
	gasUsed += slotGasUsed
	epochEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(epochEndTimeSlot))
	gasUsed += SloadGasCost
	epochEndTimeBigInt := new(big.Int).SetBytes(epochEndTime.Bytes())

	// Check if the delegation is locked up at the specified epoch
	// lockupFromEpoch <= epoch && epochEndTime <= lockupEndTime
	return lockupFromEpochBigInt.Cmp(epoch) <= 0 && epochEndTimeBigInt.Cmp(lockupEndTimeBigInt) <= 0, gasUsed, nil
}

// _highestLockupEpoch returns the highest epoch for which a delegation is locked up
// This is a port of the _highestLockupEpoch function from SFCLib.sol
func _highestLockupEpoch(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int) (*big.Int, uint64, error) {
	var gasUsed uint64 = 0

	// Get the lockup from epoch
	lockupFromEpochSlot, slotGasUsed := getLockupFromEpochSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	lockupFromEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupFromEpochSlot))
	gasUsed += SloadGasCost
	lockupFromEpochBigInt := new(big.Int).SetBytes(lockupFromEpoch.Bytes())

	// Get the current sealed epoch
	currentSealedEpochSlot := common.BigToHash(big.NewInt(currentSealedEpochSlot))
	currentSealedEpoch := evm.SfcStateDB.GetState(ContractAddress, currentSealedEpochSlot)
	gasUsed += SloadGasCost
	currentSealedEpochBigInt := new(big.Int).SetBytes(currentSealedEpoch.Bytes())

	// Check if the delegation is locked up at the current sealed epoch
	isLockedAtCurrentEpoch, lockedGasUsed, err := _isLockedUpAtEpoch(evm, delegator, toValidatorID, currentSealedEpochBigInt)
	if err != nil {
		return nil, gasUsed, err
	}
	gasUsed += lockedGasUsed

	if isLockedAtCurrentEpoch {
		return currentSealedEpochBigInt, gasUsed, nil
	}

	// Check if the delegation is locked up at the from epoch
	isLockedAtFromEpoch, lockedGasUsed, err := _isLockedUpAtEpoch(evm, delegator, toValidatorID, lockupFromEpochBigInt)
	if err != nil {
		return nil, gasUsed, err
	}
	gasUsed += lockedGasUsed

	if !isLockedAtFromEpoch {
		return big.NewInt(0), gasUsed, nil
	}

	// Binary search to find the highest epoch for which the delegation is locked up
	l := lockupFromEpochBigInt
	r := currentSealedEpochBigInt

	if l.Cmp(r) > 0 {
		return big.NewInt(0), gasUsed, nil
	}

	for l.Cmp(r) < 0 {
		m := new(big.Int).Add(l, r)
		m = m.Div(m, big.NewInt(2))

		isLockedAtM, lockedGasUsed, err := _isLockedUpAtEpoch(evm, delegator, toValidatorID, m)
		if err != nil {
			return nil, gasUsed, err
		}
		gasUsed += lockedGasUsed

		if isLockedAtM {
			l = new(big.Int).Add(m, big.NewInt(1))
		} else {
			r = m
		}
	}

	if r.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0), gasUsed, nil
	}

	return new(big.Int).Sub(r, big.NewInt(1)), gasUsed, nil
}

// _newRewardsOf calculates the new rewards for a delegation between two epochs
// This is a port of the _newRewardsOf function from SFCLib.sol
func _newRewardsOf(evm *vm.EVM, stakeAmount *big.Int, toValidatorID *big.Int, fromEpoch *big.Int, toEpoch *big.Int) (*big.Int, uint64, error) {
	var gasUsed uint64 = 0

	// If fromEpoch >= toEpoch, return 0
	if fromEpoch.Cmp(toEpoch) >= 0 {
		return big.NewInt(0), gasUsed, nil
	}

	// Get the stashed rate
	stashedRateSlot, slotGasUsed := getEpochAccumulatedRewardPerTokenSlot(fromEpoch, toValidatorID)
	gasUsed += slotGasUsed
	stashedRate := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stashedRateSlot))
	gasUsed += SloadGasCost
	stashedRateBigInt := new(big.Int).SetBytes(stashedRate.Bytes())

	// Get the current rate
	currentRateSlot, slotGasUsed := getEpochAccumulatedRewardPerTokenSlot(toEpoch, toValidatorID)
	gasUsed += slotGasUsed
	currentRate := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(currentRateSlot))
	gasUsed += SloadGasCost
	currentRateBigInt := new(big.Int).SetBytes(currentRate.Bytes())

	// Calculate the reward
	// return currentRate.sub(stashedRate).mul(stakeAmount).div(Decimal.unit());
	reward := new(big.Int).Sub(currentRateBigInt, stashedRateBigInt)
	reward = new(big.Int).Mul(reward, stakeAmount)
	reward = new(big.Int).Div(reward, getDecimalUnit())

	return reward, gasUsed, nil
}

// _newRewards calculates the new rewards for a delegation
// This is a port of the _newRewards function from SFCLib.sol
func _newRewards(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int) (Rewards, uint64, error) {
	var gasUsed uint64 = 0

	// Get the stashed rewards until epoch
	stashedRewardsUntilEpochSlot, slotGasUsed := getStashedRewardsUntilEpochSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	stashedRewardsUntilEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stashedRewardsUntilEpochSlot))
	gasUsed += SloadGasCost
	stashedUntil := new(big.Int).SetBytes(stashedRewardsUntilEpoch.Bytes())

	// Get the highest payable epoch
	payableUntil, epochGasUsed, err := _highestPayableEpoch(evm, toValidatorID)
	if err != nil {
		return Rewards{}, gasUsed, err
	}
	gasUsed += epochGasUsed

	// Get the highest lockup epoch
	lockedUntil, lockupGasUsed, err := _highestLockupEpoch(evm, delegator, toValidatorID)
	if err != nil {
		return Rewards{}, gasUsed, err
	}
	gasUsed += lockupGasUsed

	// Adjust lockedUntil if necessary
	if lockedUntil.Cmp(payableUntil) > 0 {
		lockedUntil = payableUntil
	}
	if lockedUntil.Cmp(stashedUntil) < 0 {
		lockedUntil = stashedUntil
	}

	// Get the locked delegation info
	lockedStakeSlot, slotGasUsed := getLockedStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	gasUsed += SloadGasCost
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Get the lockup duration
	lockupDurationSlot, slotGasUsed := getLockupDurationSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	lockupDuration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupDurationSlot))
	gasUsed += SloadGasCost
	lockupDurationBigInt := new(big.Int).SetBytes(lockupDuration.Bytes())

	// Get the whole stake
	stakeSlot, slotGasUsed := getStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stakeSlot))
	gasUsed += SloadGasCost
	wholeStake := new(big.Int).SetBytes(stake.Bytes())

	// Calculate the unlocked stake
	unlockedStake := new(big.Int).Sub(wholeStake, lockedStakeBigInt)
	if unlockedStake.Cmp(big.NewInt(0)) < 0 {
		unlockedStake = big.NewInt(0)
	}

	// Calculate rewards for locked stake during lockup epochs
	fullReward, rewardsGasUsed, err := _newRewardsOf(evm, lockedStakeBigInt, toValidatorID, stashedUntil, lockedUntil)
	if err != nil {
		return Rewards{}, gasUsed, err
	}
	gasUsed += rewardsGasUsed

	plReward, scaleGasUsed, err := handleInternalScaleLockupReward(evm, fullReward, lockupDurationBigInt)
	if err != nil {
		return Rewards{}, gasUsed, err
	}
	gasUsed += scaleGasUsed
	log.Debug("_newRewards: plReward", "slot", common.Bytes2Hex(stakeSlot.Bytes()), "plReward", plReward)

	// Calculate rewards for unlocked stake during lockup epochs
	fullReward, rewardsGasUsed, err = _newRewardsOf(evm, unlockedStake, toValidatorID, stashedUntil, lockedUntil)
	if err != nil {
		return Rewards{}, gasUsed, err
	}
	gasUsed += rewardsGasUsed

	puReward, scaleGasUsed, err := handleInternalScaleLockupReward(evm, fullReward, big.NewInt(0))
	if err != nil {
		return Rewards{}, gasUsed, err
	}
	gasUsed += scaleGasUsed
	log.Debug("_newRewards: puReward", "slot", common.Bytes2Hex(stakeSlot.Bytes()), "puReward", puReward)

	// Calculate rewards for whole stake during unlocked epochs
	fullReward, rewardsGasUsed, err = _newRewardsOf(evm, wholeStake, toValidatorID, lockedUntil, payableUntil)
	if err != nil {
		return Rewards{}, gasUsed, err
	}
	gasUsed += rewardsGasUsed

	wuReward, scaleGasUsed, err := handleInternalScaleLockupReward(evm, fullReward, big.NewInt(0))
	if err != nil {
		return Rewards{}, gasUsed, err
	}
	gasUsed += scaleGasUsed
	log.Debug("_newRewards: wuReward", "slot", common.Bytes2Hex(stakeSlot.Bytes()), "wuReward", wuReward)

	finalRewards := sumRewards3(plReward, puReward, wuReward)
	log.Debug("_newRewards: finalRewards", "finalRewards", finalRewards)
	return finalRewards, gasUsed, nil
}

// getEpochEndTimeSlot returns the storage slot for an epoch's end time
func getEpochEndTimeSlot(epoch *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(uint256 => EpochSnapshot), first we need to get the slot for the EpochSnapshot struct
	// keccak256(abi.encode(epoch, epochSnapshotSlot))
	// Then we need to add the offset for the endTime field (7)

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(epoch, epochSnapshotSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int and add the offset for the endTime field
	slot := new(big.Int).SetBytes(hash)
	slot = new(big.Int).Add(slot, big.NewInt(endTimeOffset))

	return slot, gasUsed
}

// getEpochAccumulatedRewardPerTokenSlot returns the storage slot for an epoch's accumulated reward per token
func getEpochAccumulatedRewardPerTokenSlot(epoch *big.Int, validatorID *big.Int) (*big.Int, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(uint256 => EpochSnapshot), first we need to get the slot for the EpochSnapshot struct
	// keccak256(abi.encode(epoch, epochSnapshotSlot))
	// Then we need to get the slot for the accumulatedRewardPerToken mapping
	// keccak256(abi.encode(validatorID, keccak256(abi.encode(epoch, epochSnapshotSlot)) + accumulatedRewardPerTokenOffset))

	// Create the inner hash input using cached padded values
	thisEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// Update accumulated reward per token
	// For a mapping within a struct, we need to calculate the slot as:
	// keccak256(key . (struct_slot + offset))
	// Add the offset for the accumulatedRewardPerToken mapping within the struct
	mappingSlot := new(big.Int).Add(thisEpochSnapshotSlot, big.NewInt(accumulatedRewardPerTokenOffset))

	// Then, calculate the slot for the specific key using our helper function
	// Use CreateValidatorMappingHashInput to create the hash input
	// Declare outerHashInput at the beginning of the function
	outerHashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
	// Use cached hash calculation
	accumulatedRewardPerTokenSlotHash := CachedKeccak256Hash(outerHashInput)

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(accumulatedRewardPerTokenSlotHash.Bytes())

	return slot, gasUsed
}

// handleInternalScaleLockupReward scales the reward based on the lockup duration
// This is a port of the _scaleLockupReward function from SFCBase.sol
func handleInternalScaleLockupReward(evm *vm.EVM, fullReward *big.Int, lockupDuration *big.Int) (Rewards, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Initialize reward with zeros
	reward := Rewards{
		LockupExtraReward: big.NewInt(0),
		LockupBaseReward:  big.NewInt(0),
		UnlockedReward:    big.NewInt(0),
	}

	// Get the unlockedRewardRatio from the constants manager
	unlockedRewardRatioBigInt := getConstantsManagerVariable(evm.SfcStateDB, "unlockedRewardRatio")

	// Check if lockupDuration is not zero
	if lockupDuration.Cmp(big.NewInt(0)) != 0 {
		// Get the decimal unit
		decimalUnit := getDecimalUnit()

		// Calculate maxLockupExtraRatio = Decimal.unit() - unlockedRewardRatio
		maxLockupExtraRatio := new(big.Int).Sub(decimalUnit, unlockedRewardRatioBigInt)

		// Get the maxLockupDuration from the constants manager
		maxLockupDurationBigInt := getConstantsManagerVariable(evm.SfcStateDB, "maxLockupDuration")

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

// packRewards manually packs a Rewards struct into bytes for storage
// Since internal functions don't have ABI, we need to pack the struct manually
// The Rewards struct has three fields stored in consecutive slots:
// - LockupExtraReward at slot N
// - LockupBaseReward at slot N+1
// - UnlockedReward at slot N+2
func packRewards(rewards Rewards) [][]byte {
	result := make([][]byte, 3)
	// Pack each field into 32 bytes
	result[0] = common.BigToHash(rewards.LockupExtraReward).Bytes() // First slot
	result[1] = common.BigToHash(rewards.LockupBaseReward).Bytes()  // Second slot
	result[2] = common.BigToHash(rewards.UnlockedReward).Bytes()    // Third slot
	return result
}

// unpackRewards manually unpacks bytes into a Rewards struct
// This is the reverse operation of packRewards
func unpackRewards(packedData [][]byte) Rewards {
	if len(packedData) != 3 {
		return Rewards{
			LockupExtraReward: big.NewInt(0),
			LockupBaseReward:  big.NewInt(0),
			UnlockedReward:    big.NewInt(0),
		}
	}
	return Rewards{
		LockupExtraReward: new(big.Int).SetBytes(packedData[0]),
		LockupBaseReward:  new(big.Int).SetBytes(packedData[1]),
		UnlockedReward:    new(big.Int).SetBytes(packedData[2]),
	}
}

// handleInternalCalcRawValidatorEpochBaseReward calculates the raw base reward for a validator in an epoch
// This is a port of the _calcRawValidatorEpochBaseReward function from SFCBase.sol
func handleInternalCalcRawValidatorEpochBaseReward(epochDuration *big.Int, baseRewardPerSecond *big.Int, baseRewardWeight *big.Int, totalBaseRewardWeight *big.Int) *big.Int {
	if baseRewardWeight.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	if totalBaseRewardWeight.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	totalReward := new(big.Int).Mul(epochDuration, baseRewardPerSecond)
	return new(big.Int).Div(new(big.Int).Mul(totalReward, baseRewardWeight), totalBaseRewardWeight)
}

// handleInternalCalcRawValidatorEpochTxReward calculates the raw transaction reward for a validator in an epoch
// This is a port of the _calcRawValidatorEpochTxReward function from SFCBase.sol
func handleInternalCalcRawValidatorEpochTxReward(stateDb vm.StateDB, epochFee *big.Int, txRewardWeight *big.Int, totalTxRewardWeight *big.Int) *big.Int {
	if txRewardWeight.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	if totalTxRewardWeight.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	txReward := new(big.Int).Div(new(big.Int).Mul(epochFee, txRewardWeight), totalTxRewardWeight)
	// fee reward except burntFeeShare and treasuryFeeShare
	burntFeeShare := getConstantsManagerVariable(stateDb, "burntFeeShare")
	treasuryFeeShare := getConstantsManagerVariable(stateDb, "treasuryFeeShare")
	shareToSubtract := new(big.Int).Add(burntFeeShare, treasuryFeeShare)
	shareToKeep := new(big.Int).Sub(getDecimalUnit(), shareToSubtract)
	return new(big.Int).Div(new(big.Int).Mul(txReward, shareToKeep), getDecimalUnit())
}

// handleInternalCalcValidatorCommission calculates the validator commission from raw reward
// This is a port of the _calcValidatorCommission function from SFCBase.sol
func handleInternalCalcValidatorCommission(rawReward *big.Int, commission *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Mul(rawReward, commission), getDecimalUnit())
}

// readDynamicBytes reads a dynamic bytes value from storage following Solidity's storage layout
// In Solidity, dynamic bytes are stored as:
// - At the main slot: the length of the bytes
// - If length <= 31: data is packed with length in the same slot
// - If length > 31: data starts at keccak256(main_slot) for as many slots as needed
func readDynamicBytes(evm *vm.EVM, mainSlot *big.Int) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Read the main slot to get length and possibly data
	mainSlotValue := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(mainSlot))
	gasUsed += SloadGasCost

	// Convert to 32-byte array to ensure proper indexing
	mainSlotBytes := common.LeftPadBytes(mainSlotValue.Bytes(), 32)

	// Extract the length from the least significant byte
	// In Solidity, if length is odd, the last bit is 1 and length = (value & 0xFF) / 2
	// If length is even, the data is stored inline and length = (value & 0xFF) / 2
	lastByte := mainSlotBytes[31] // Get the last byte

	if lastByte&1 == 0 {
		// Data is stored inline (length <= 31)
		length := lastByte / 2
		if length == 0 {
			return []byte{}, gasUsed, nil
		}

		// Extract the data from the same slot (first 'length' bytes)
		if len(mainSlotBytes) >= int(length) {
			return mainSlotBytes[:length], gasUsed, nil
		}
		return mainSlotBytes, gasUsed, nil
	} else {
		// Data is stored in separate slots (length > 31)
		// Length is stored in the main slot as: length * 2 + 1
		length := new(big.Int).SetBytes(mainSlotBytes)
		length = length.Sub(length, big.NewInt(1)) // subtract 1
		length = length.Div(length, big.NewInt(2)) // length = (stored_value - 1) / 2

		if length.Cmp(big.NewInt(0)) == 0 {
			return []byte{}, gasUsed, nil
		}

		// Calculate the data starting slot: keccak256(main_slot)
		mainSlotBytes := common.LeftPadBytes(mainSlot.Bytes(), 32)
		dataStartSlotHash := CachedKeccak256(mainSlotBytes)
		gasUsed += HashGasCost
		dataStartSlot := new(big.Int).SetBytes(dataStartSlotHash)

		// Calculate how many slots we need to read
		lengthInt := length.Uint64()
		slotsNeeded := (lengthInt + 31) / 32 // Round up to nearest 32-byte slot

		// Read the data from storage
		data := make([]byte, 0, lengthInt)
		for i := uint64(0); i < slotsNeeded; i++ {
			slotToRead := new(big.Int).Add(dataStartSlot, big.NewInt(int64(i)))
			slotValue := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotToRead))
			gasUsed += SloadGasCost

			// Convert to 32-byte array to ensure proper padding
			slotBytes := common.LeftPadBytes(slotValue.Bytes(), 32)

			// Append the slot data (32 bytes max)
			if len(data)+len(slotBytes) > int(lengthInt) {
				// Last slot - only take what we need
				remaining := int(lengthInt) - len(data)
				data = append(data, slotBytes[:remaining]...)
			} else {
				data = append(data, slotBytes...)
			}
		}

		return data[:lengthInt], gasUsed, nil
	}
}

// writeDynamicBytes writes a dynamic bytes value to storage following Solidity's storage layout
func writeDynamicBytes(evm *vm.EVM, mainSlot *big.Int, data []byte) (uint64, error) {
	var gasUsed uint64 = 0

	dataLength := len(data)

	if dataLength <= 31 {
		// Store data inline in the main slot
		// Format: [data...] [length * 2] (length in last byte, even number)
		slotValue := make([]byte, 32)
		copy(slotValue, data)
		slotValue[31] = byte(dataLength * 2) // Store length * 2 in last byte

		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(mainSlot), common.BytesToHash(slotValue))
		gasUsed += SstoreGasCost
	} else {
		// Store length in main slot and data in separate slots
		// Main slot stores: length * 2 + 1
		lengthValue := big.NewInt(int64(dataLength*2 + 1))
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(mainSlot), common.BigToHash(lengthValue))
		gasUsed += SstoreGasCost

		// Calculate the data starting slot: keccak256(main_slot)
		mainSlotBytes := common.LeftPadBytes(mainSlot.Bytes(), 32)
		dataStartSlotHash := CachedKeccak256(mainSlotBytes)
		gasUsed += HashGasCost
		dataStartSlot := new(big.Int).SetBytes(dataStartSlotHash)

		// Calculate how many slots we need to write
		slotsNeeded := (dataLength + 31) / 32 // Round up to nearest 32-byte slot

		// Write the data to storage
		for i := 0; i < slotsNeeded; i++ {
			slotToWrite := new(big.Int).Add(dataStartSlot, big.NewInt(int64(i)))

			// Prepare the data for this slot (32 bytes max)
			start := i * 32
			end := start + 32
			if end > dataLength {
				end = dataLength
			}

			slotData := make([]byte, 32)
			copy(slotData, data[start:end])

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(slotToWrite), common.BytesToHash(slotData))
			gasUsed += SstoreGasCost
		}
	}

	return gasUsed, nil
}
