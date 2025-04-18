package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/rlp"
)

// Gas costs for storage operations
const (
	SloadGasCost  uint64 = 2100  // Cost of SLOAD (GetState) operation (ColdSloadCostEIP2929)
	SstoreGasCost uint64 = 20000 // Cost of SSTORE (SetState) operation (SstoreSetGasEIP2200)
	HashGasCost   uint64 = 30    // Cost of hash operation (Keccak256)
)

// EpochSnapshot struct offsets
const (
	endTimeOffset                     int64 = 0  // uint256 endTime
	epochFeeOffset                    int64 = 1  // uint256 epochFee
	totalBaseRewardOffset             int64 = 2  // uint256 totalBaseRewardWeight
	totalTxRewardOffset               int64 = 3  // uint256 totalTxRewardWeight
	baseRewardPerSecondOffset         int64 = 4  // uint256 baseRewardPerSecond
	totalStakeOffset                  int64 = 5  // uint256 totalStake
	totalSupplyOffset                 int64 = 6  // uint256 totalSupply
	validatorIDsOffset                int64 = 7  // uint256[] validatorIDs
	offlineTimeOffset                 int64 = 8  // mapping(uint256 => uint256) offlineTime
	offlineBlocksOffset               int64 = 9  // mapping(uint256 => uint256) offlineBlocks
	accumulatedRewardPerTokenOffset   int64 = 10 // mapping(uint256 => uint256) accumulatedRewardPerToken
	accumulatedUptimeOffset           int64 = 11 // mapping(uint256 => uint256) accumulatedUptime
	accumulatedOriginatedTxsFeeOffset int64 = 12 // mapping(uint256 => uint256) accumulatedOriginatedTxsFee
	receiveStakeOffset                int64 = 13 // mapping(uint256 => uint256) receivedStake
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
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(statusSlot)))
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
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(statusSlot)))
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
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(statusSlot)))
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
	delegation := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlotNum)))
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
	delegation := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlotNum)))
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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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
	gasUsed += HashGasCost

	// Step 2: Calculate keccak256(abi.encode(toValidatorID, innerHash1))
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput2 := append(toValidatorIDBytes, innerHash1...)
	innerHash2 := crypto.Keccak256(innerHashInput2)
	gasUsed += HashGasCost

	// Step 3: Calculate keccak256(abi.encode(wrID, innerHash2))
	wrIDBytes := common.LeftPadBytes(wrID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(wrIDBytes, innerHash2...)
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

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
func getWithdrawalRequestAmountSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (int64, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The amount field is at the base slot (first field in the struct)
	return baseSlot, gasUsed
}

// getWithdrawalRequestEpochSlot calculates the storage slot for a withdrawal request epoch
func getWithdrawalRequestEpochSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (int64, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The WithdrawalRequest struct has the following fields:
	// uint256 amount;
	// uint256 epoch;
	// uint256 time;

	// The epoch field is at slot + 1
	return new(big.Int).Add(big.NewInt(baseSlot), big.NewInt(1)).Int64(), gasUsed
}

// getWithdrawalRequestTimeSlot calculates the storage slot for a withdrawal request time
func getWithdrawalRequestTimeSlot(delegator common.Address, toValidatorID *big.Int, wrID *big.Int) (int64, uint64) {
	// Get the base slot for the withdrawal request
	baseSlot, gasUsed := getWithdrawalRequestSlot(delegator, toValidatorID, wrID)

	// The WithdrawalRequest struct has the following fields:
	// uint256 amount;
	// uint256 epoch;
	// uint256 time;

	// The time field is at slot + 2
	return new(big.Int).Add(big.NewInt(baseSlot), big.NewInt(2)).Int64(), gasUsed
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

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

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

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

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
	// Get the base slot for the locked delegation
	baseSlot, gasUsed := getLockedStakeSlot(delegator, toValidatorID)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The fromEpoch field is at slot + 1
	return new(big.Int).Add(big.NewInt(baseSlot), big.NewInt(1)).Int64(), gasUsed
}

// getLockupEndTimeSlot calculates the storage slot for a delegation's lockup end time
func getLockupEndTimeSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Get the base slot for the locked delegation
	baseSlot, gasUsed := getLockedStakeSlot(delegator, toValidatorID)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The endTime field is at slot + 2
	return new(big.Int).Add(big.NewInt(baseSlot), big.NewInt(2)).Int64(), gasUsed
}

// getLockupDurationSlot calculates the storage slot for a delegation's lockup duration
func getLockupDurationSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Get the base slot for the locked delegation
	baseSlot, gasUsed := getLockedStakeSlot(delegator, toValidatorID)

	// The LockedDelegation struct has the following fields:
	// uint256 lockedStake;
	// uint256 fromEpoch;
	// uint256 endTime;
	// uint256 duration;

	// The duration field is at slot + 3
	return new(big.Int).Add(big.NewInt(baseSlot), big.NewInt(3)).Int64(), gasUsed
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
func getRewardsStashSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => Rewards)), first we need to get the slot for the Rewards struct
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, rewardsStashSlot))))

	// Create the inner hash input: abi.encode(delegator, rewardsStashSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                           // Left-pad to 32 bytes
	rewardsStashSlotBytes := common.LeftPadBytes(big.NewInt(rewardsStashSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, rewardsStashSlotBytes...)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// getStashedLockupRewardsSlot calculates the storage slot for a delegation's stashed lockup rewards
func getStashedLockupRewardsSlot(delegator common.Address, toValidatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// For a mapping(address => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(toValidatorID, keccak256(abi.encode(delegator, stashedLockupRewardsSlot))))

	// Create the inner hash input: abi.encode(delegator, stashedLockupRewardsSlot)
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)                                           // Left-pad to 32 bytes
	stashedLockupRewardsSlotBytes := common.LeftPadBytes(big.NewInt(stashedLockupRewardsSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(delegatorBytes, stashedLockupRewardsSlotBytes...)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

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

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(toValidatorID, innerHash)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(toValidatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

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
	gasUsed += SloadGasCost // Add gas for SLOAD
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
	gasUsed += SloadGasCost // Add gas for SLOAD
	currentSealedEpochBigInt := new(big.Int).SetBytes(currentSealedEpoch.Bytes())

	// Calculate current epoch as currentSealedEpoch + 1
	currentEpochBigInt := new(big.Int).Add(currentSealedEpochBigInt, big.NewInt(1))

	return currentEpochBigInt, gasUsed, nil
}

// getEpochSnapshotSlot calculates the storage slot for an epoch snapshot
func getEpochSnapshotSlot(epoch *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// For a mapping(uint256 => EpochSnapshot), the slot is calculated as:
	// keccak256(abi.encode(epoch, epochSnapshotSlot))

	// Create the hash input: abi.encode(epoch, epochSnapshotSlot)
	epochBytes := common.LeftPadBytes(epoch.Bytes(), 32)                                     // Left-pad to 32 bytes
	epochSnapshotSlotBytes := common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32) // Left-pad to 32 bytes
	hashInput := append(epochBytes, epochSnapshotSlotBytes...)

	// Calculate the hash - add gas cost for hashing
	hash := crypto.Keccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(hash)

	return slot.Int64(), gasUsed
}

// getOfflinePenaltyThresholdBlocksNum gets the offline penalty threshold blocks number from the constants manager
func getOfflinePenaltyThresholdBlocksNum(evm *vm.EVM) (*big.Int, uint64, error) {
	result, gasUsed, err := callConstantManagerMethod(evm, "offlinePenaltyThresholdBlocksNum")
	if err != nil || len(result) == 0 {
		return nil, gasUsed, err
	}

	threshold, ok := result[0].(*big.Int)
	if !ok {
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

// decodeValidatorIDs decodes validator IDs from storage
func decodeValidatorIDs(data []byte) ([]*big.Int, error) {
	// If data is empty, return an empty array
	if len(data) == 0 {
		return []*big.Int{}, nil
	}

	// Decode the validator IDs
	// The data is expected to be an RLP-encoded array of big.Int
	var validatorIDs []*big.Int
	err := rlp.DecodeBytes(data, &validatorIDs)
	if err != nil {
		return nil, err
	}

	return validatorIDs, nil
}

// getValidatorStatusSlotByID calculates the storage slot for a validator's status by ID
func getValidatorStatusSlotByID(validatorID *big.Int) (int64, uint64) {
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
	status := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(statusSlot)))
	gasUsed += SloadGasCost
	statusBigInt := new(big.Int).SetBytes(status.Bytes())

	// Check if the validator is already deactivated with this status
	currentStatus := statusBigInt.Uint64()
	if currentStatus == OK_STATUS && statusBit != OK_STATUS {
		// Get the validator's received stake slot
		receivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed

		// Get the validator's received stake
		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(receivedStakeSlot)))
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
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(statusSlot)), common.BigToHash(newStatus))
		gasUsed += SstoreGasCost

		// Check if the validator is not already deactivated
		deactivatedEpochSlot, slotGasUsed := getValidatorDeactivatedEpochSlot(validatorID)
		gasUsed += slotGasUsed

		deactivatedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(deactivatedEpochSlot)))
		gasUsed += SloadGasCost
		deactivatedEpochBigInt := new(big.Int).SetBytes(deactivatedEpoch.Bytes())

		if deactivatedEpochBigInt.Cmp(big.NewInt(0)) == 0 {
			// Set the deactivated epoch to the current epoch
			currentEpochBigInt, epochGasUsed, err := getCurrentEpoch(evm)
			gasUsed += epochGasUsed
			if err != nil {
				return gasUsed, err
			}

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(deactivatedEpochSlot)), common.BigToHash(currentEpochBigInt))
			gasUsed += SstoreGasCost

			// Set the deactivated time to the current time
			deactivatedTimeSlot, slotGasUsed := getValidatorDeactivatedTimeSlot(validatorID)
			gasUsed += slotGasUsed

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(deactivatedTimeSlot)), common.BigToHash(evm.Context.Time))
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

	validatorCreatedTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorCreatedTimeSlot)))
	gasUsed += SloadGasCost

	if validatorCreatedTime.Big().Cmp(big.NewInt(0)) == 0 {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Get the validator's received stake
	receivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
	gasUsed += slotGasUsed

	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(receivedStakeSlot)))
	gasUsed += SloadGasCost
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

	// Get the validator's status
	statusSlot, slotGasUsed := getValidatorStatusSlotByID(validatorID)
	gasUsed += slotGasUsed

	status := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(statusSlot)))
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
	data, err := DriverAbi.Pack("updateValidatorWeight", validatorID, weight)
	if err != nil {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Call the node driver
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return gasUsed, err
	}

	// If syncPubkey is true and weight is not zero, update validator pubkey
	if syncPubkey && weight.Cmp(big.NewInt(0)) != 0 {
		// Get the validator's pubkey
		pubkeySlot, slotGasUsed := getValidatorPubkeySlot(validatorID)
		gasUsed += slotGasUsed

		pubkeyHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(pubkeySlot)))
		gasUsed += SloadGasCost

		// Pack the function call data
		data, err := DriverAbi.Pack("updateValidatorPubkey", validatorID, pubkeyHash.Bytes())
		if err != nil {
			return gasUsed, vm.ErrExecutionReverted
		}

		// Call the node driver
		_, _, err = evm.Call(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, 50000, big.NewInt(0))
		if err != nil {
			return gasUsed, err
		}
	}

	return gasUsed, nil
}

// getEpochValidatorOfflineTimeSlot calculates the storage slot for a validator's offline time in an epoch
func getEpochValidatorOfflineTimeSlot(epoch *big.Int, validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping(uint256 => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(validatorID, keccak256(abi.encode(offlineTimeOffset, epochSnapshotSlot))))

	// Create the inner hash input: abi.encode(offlineTimeOffset, epochSnapshotSlot)
	offlineTimeOffsetBytes := common.LeftPadBytes(big.NewInt(offlineTimeOffset).Bytes(), 32) // Left-pad to 32 bytes
	epochSnapshotSlotBytes := common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32) // Left-pad to 32 bytes
	innerHashInput := append(offlineTimeOffsetBytes, epochSnapshotSlotBytes...)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(validatorID, innerHash)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(validatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// getEpochValidatorOfflineBlocksSlot calculates the storage slot for a validator's offline blocks in an epoch
func getEpochValidatorOfflineBlocksSlot(epoch *big.Int, validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping(uint256 => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(validatorID, keccak256(abi.encode(offlineBlocksOffset, epochSnapshotSlot))))

	// Create the inner hash input: abi.encode(offlineBlocksOffset, epochSnapshotSlot)
	offlineBlocksOffsetBytes := common.LeftPadBytes(big.NewInt(offlineBlocksOffset).Bytes(), 32) // Left-pad to 32 bytes
	epochSnapshotSlotBytes := common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32)     // Left-pad to 32 bytes
	innerHashInput := append(offlineBlocksOffsetBytes, epochSnapshotSlotBytes...)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(validatorID, innerHash)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(validatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// getEpochValidatorAccumulatedRewardPerTokenSlot calculates the storage slot for a validator's accumulated reward per token in an epoch
func getEpochValidatorAccumulatedRewardPerTokenSlot(epoch *big.Int, validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping(uint256 => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(validatorID, keccak256(abi.encode(accumulatedRewardPerTokenOffset, epochSnapshotSlot))))

	// Create the inner hash input: abi.encode(accumulatedRewardPerTokenOffset, epochSnapshotSlot)
	accumulatedRewardPerTokenOffsetBytes := common.LeftPadBytes(big.NewInt(accumulatedRewardPerTokenOffset).Bytes(), 32) // Left-pad to 32 bytes
	epochSnapshotSlotBytes := common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32)                             // Left-pad to 32 bytes
	innerHashInput := append(accumulatedRewardPerTokenOffsetBytes, epochSnapshotSlotBytes...)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(validatorID, innerHash)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(validatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// getEpochValidatorAccumulatedUptimeSlot calculates the storage slot for a validator's accumulated uptime in an epoch
func getEpochValidatorAccumulatedUptimeSlot(epoch *big.Int, validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping(uint256 => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(validatorID, keccak256(abi.encode(accumulatedUptimeOffset, epochSnapshotSlot))))

	// Create the inner hash input: abi.encode(accumulatedUptimeOffset, epochSnapshotSlot)
	accumulatedUptimeOffsetBytes := common.LeftPadBytes(big.NewInt(accumulatedUptimeOffset).Bytes(), 32) // Left-pad to 32 bytes
	epochSnapshotSlotBytes := common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32)             // Left-pad to 32 bytes
	innerHashInput := append(accumulatedUptimeOffsetBytes, epochSnapshotSlotBytes...)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(validatorID, innerHash)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(validatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// getEpochValidatorAccumulatedOriginatedTxsFeeSlot calculates the storage slot for a validator's accumulated originated txs fee in an epoch
func getEpochValidatorAccumulatedOriginatedTxsFeeSlot(epoch *big.Int, validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping(uint256 => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(validatorID, keccak256(abi.encode(accumulatedOriginatedTxsFeeOffset, epochSnapshotSlot))))

	// Create the inner hash input: abi.encode(accumulatedOriginatedTxsFeeOffset, epochSnapshotSlot)
	accumulatedOriginatedTxsFeeOffsetBytes := common.LeftPadBytes(big.NewInt(accumulatedOriginatedTxsFeeOffset).Bytes(), 32) // Left-pad to 32 bytes
	epochSnapshotSlotBytes := common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32)                                 // Left-pad to 32 bytes
	innerHashInput := append(accumulatedOriginatedTxsFeeOffsetBytes, epochSnapshotSlotBytes...)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(validatorID, innerHash)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(validatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
}

// getEpochValidatorReceivedStakeSlot calculates the storage slot for a validator's received stake in an epoch
func getEpochValidatorReceivedStakeSlot(epoch *big.Int, validatorID *big.Int) (int64, uint64) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// First get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// For a mapping(uint256 => mapping(uint256 => uint256)), the slot is calculated as:
	// keccak256(abi.encode(validatorID, keccak256(abi.encode(receiveStakeOffset, epochSnapshotSlot))))

	// Create the inner hash input: abi.encode(receiveStakeOffset, epochSnapshotSlot)
	receiveStakeOffsetBytes := common.LeftPadBytes(big.NewInt(receiveStakeOffset).Bytes(), 32) // Left-pad to 32 bytes
	epochSnapshotSlotBytes := common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32)   // Left-pad to 32 bytes
	innerHashInput := append(receiveStakeOffsetBytes, epochSnapshotSlotBytes...)

	// Calculate the inner hash - add gas cost for hashing
	innerHash := crypto.Keccak256(innerHashInput)
	gasUsed += HashGasCost

	// Create the outer hash input: abi.encode(validatorID, innerHash)
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32) // Left-pad to 32 bytes
	outerHashInput := append(validatorIDBytes, innerHash...)

	// Calculate the outer hash - add gas cost for hashing
	outerHash := crypto.Keccak256(outerHashInput)
	gasUsed += HashGasCost

	// Convert the hash to a big.Int
	slot := new(big.Int).SetBytes(outerHash)

	return slot.Int64(), gasUsed
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
