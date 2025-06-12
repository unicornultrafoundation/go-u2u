package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleInitialize initializes the SFC contract
func handleInitialize(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if contract is already initialized
	revertData, checkGasUsed, err := checkAlreadyInitialized(evm, "initialize")
	var gasUsed uint64 = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 6 {
		return nil, 0, vm.ErrExecutionReverted
	}
	sealedEpoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	_totalSupply, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	nodeDriver, ok := args[2].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	lib, ok := args[3].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	_c, ok := args[4].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	owner, ok := args[5].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the addresses are not zero
	emptyAddr := common.Address{}
	if nodeDriver.Cmp(emptyAddr) == 0 || lib.Cmp(emptyAddr) == 0 || _c.Cmp(emptyAddr) == 0 || owner.Cmp(emptyAddr) == 0 {
		revertData, err := encodeRevertReason("initialize", "zero address")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Set the owner
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), common.BytesToHash(owner.Bytes()))

	// Set the current sealed epoch
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)), common.BigToHash(sealedEpoch))

	// Set the node driver
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(nodeDriverAuthSlot)), common.BytesToHash(nodeDriver.Bytes()))

	// Set the lib address
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(libAddressSlot)), common.BytesToHash(lib.Bytes()))

	// Set the constants manager
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(constantsManagerSlot)), common.BytesToHash(_c.Bytes()))

	// Set the total supply
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)), common.BigToHash(_totalSupply))

	// Set the min gas price
	initialMinGasPrice := big.NewInt(1000000000) // 1 gwei
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)), common.BigToHash(initialMinGasPrice))

	// Set the epoch snapshot end time
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(epochSnapshotSlot)), common.BigToHash(evm.Context.Time))

	// Set the initialized flag
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(isInitialized)), common.BigToHash(big.NewInt(1)))

	// Emit OwnershipTransferred event
	emptyHash := common.Hash{}
	topics := []common.Hash{
		SfcAbi.Events["OwnershipTransferred"].ID,
		emptyHash, // indexed parameter (previous owner - zero address)
		common.BytesToHash(common.LeftPadBytes(owner.Bytes(), 32)), // indexed parameter (new owner)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleUpdateStakeTokenizerAddress updates the stake tokenizer address
func handleUpdateStakeTokenizerAddress(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, checkGasUsed, err := checkOnlyOwner(evm, caller, "updateStakeTokenizerAddress")
	var gasUsed uint64 = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the stake tokenizer address
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(stakeTokenizerAddressSlot)), common.BytesToHash(addr.Bytes()))

	return nil, 0, nil
}

// handleUpdateLibAddress updates the lib address
func handleUpdateLibAddress(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, checkGasUsed, err := checkOnlyOwner(evm, caller, "updateLibAddress")
	var gasUsed uint64 = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	v, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the lib address
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(libAddressSlot)), common.BytesToHash(v.Bytes()))

	return nil, 0, nil
}

// UpdateTreasuryAddress updates the address of the treasury
func handleUpdateTreasuryAddress(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateTreasuryAddress handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UpdateConstsAddress updates the address of the ConstantsManager contract
func handleUpdateConstsAddress(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateConstsAddress handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// ConstsAddress returns the address of the ConstantsManager contract
func handleConstsAddress(evm *vm.EVM) ([]byte, uint64, error) {
	// TODO: Implement constsAddress handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UpdateVoteBookAddress updates the address of the vote book
func handleUpdateVoteBookAddress(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateVoteBookAddress handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}
