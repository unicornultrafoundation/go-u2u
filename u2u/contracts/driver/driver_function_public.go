package driver

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// Handler functions for NodeDriver contract public functions

// handleSetBackend sets the backend address
func handleSetBackend(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "setBackend")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new backend address from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	newBackend, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the new backend is not the zero address
	emptyAddr := common.Address{}
	if newBackend.Cmp(emptyAddr) == 0 {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the new backend address
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot)), newBackend.Hash())

	// Emit UpdatedBackend event
	topics := []common.Hash{
		DriverAbi.Events["UpdatedBackend"].ID,
		common.BytesToHash(common.LeftPadBytes(newBackend.Bytes(), 32)), // indexed parameter
	}
	var data []byte
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleInitialize initializes the NodeDriver contract
func handleInitialize(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if contract is already initialized
	backend := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot)))
	evmWriter := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(evmWriterSlot)))
	if backend.Cmp(common.Hash{}) != 0 || evmWriter.Cmp(common.Hash{}) != 0 {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the backend and evmWriter addresses from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	newBackend, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	newEvmWriter, ok := args[1].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the new backend and evmWriter are not the zero address
	emptyAddr := common.Address{}
	if newBackend.Cmp(emptyAddr) == 0 || newEvmWriter.Cmp(emptyAddr) == 0 {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the backend and evmWriter addresses
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot)), newBackend.Hash())
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(evmWriterSlot)), newEvmWriter.Hash())

	// Emit UpdatedBackend event
	topics := []common.Hash{
		DriverAbi.Events["UpdatedBackend"].ID,
		common.BytesToHash(common.LeftPadBytes(newBackend.Bytes(), 32)), // indexed parameter
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleSetBalance sets the balance of an account
func handleSetBalance(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "setBalance")
	if err != nil {
		return revertData, 0, err
	}

	// Get the account and value from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	account, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract to set the balance
	evmWriterAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(evmWriterSlot))).Bytes())

	// Pack the function call data
	data, err := EvmWriterAbi.Pack("setBalance", account, value)
	if err != nil {
		log.Error("Driver SetBalance: Error packing function call data", "method", "setBalance",
			"account", account.Hex(), "value", value, "error", err)
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract
	result, _, err := evm.Call(vm.AccountRef(ContractAddress), evmWriterAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("Driver SetBalance: Error calling EVMWriter", "error", err, "method", "setBalance", "reason", reason)
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleCopyCode copies code from one account to another
func handleCopyCode(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "copyCode")
	if err != nil {
		return revertData, 0, err
	}

	// Get the accounts from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	account, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	from, ok := args[1].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract to copy the code
	evmWriterAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(evmWriterSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("copyCode", account, from)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), evmWriterAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSwapCode swaps code between two accounts
func handleSwapCode(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "swapCode")
	if err != nil {
		return revertData, 0, err
	}

	// Get the accounts from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	account, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	with, ok := args[1].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract to swap the code
	evmWriterAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(evmWriterSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("swapCode", account, with)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), evmWriterAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSetStorage sets the storage of an account
func handleSetStorage(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "setStorage")
	if err != nil {
		return revertData, 0, err
	}

	// Get the account, key, and value from args
	if len(args) != 3 {
		return nil, 0, vm.ErrExecutionReverted
	}
	account, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	key, ok := args[1].(common.Hash)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[2].(common.Hash)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract to set the storage
	evmWriterAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(evmWriterSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("setStorage", account, key, value)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), evmWriterAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleIncNonce increments the nonce of an account
func handleIncNonce(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "incNonce")
	if err != nil {
		return revertData, 0, err
	}

	// Get the account and diff from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	account, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	diff, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract to increment the nonce
	evmWriterAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(evmWriterSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("incNonce", account, diff)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the EVMWriter contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), evmWriterAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleUpdateNetworkRules updates the network rules
func handleUpdateNetworkRules(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "updateNetworkRules")
	if err != nil {
		return revertData, 0, err
	}

	// Get the diff from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	diff, ok := args[0].([]byte)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Emit UpdateNetworkRules event
	topics := []common.Hash{
		DriverAbi.Events["UpdateNetworkRules"].ID,
	}

	// Pack the event data
	data, err := DriverAbi.Events["UpdateNetworkRules"].Inputs.NonIndexed().Pack(diff)
	if err != nil {
		log.Error("Driver: Error packing UpdateNetworkRules event data")
		return nil, 0, vm.ErrExecutionReverted
	}

	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleUpdateNetworkVersion updates the network version
func handleUpdateNetworkVersion(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "updateNetworkVersion")
	if err != nil {
		return revertData, 0, err
	}

	// Get the version from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	version, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Emit UpdateNetworkVersion event
	topics := []common.Hash{
		DriverAbi.Events["UpdateNetworkVersion"].ID,
	}

	// Pack the event data
	data, err := DriverAbi.Events["UpdateNetworkVersion"].Inputs.NonIndexed().Pack(version)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleAdvanceEpochs advances the epochs
func handleAdvanceEpochs(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "advanceEpochs")
	if err != nil {
		return revertData, 0, err
	}

	// Get the num from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	num, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Emit AdvanceEpochs event
	topics := []common.Hash{
		DriverAbi.Events["AdvanceEpochs"].ID,
	}

	// Pack the event data
	data, err := DriverAbi.Events["AdvanceEpochs"].Inputs.NonIndexed().Pack(num)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleUpdateValidatorWeight updates the validator weight
func handleUpdateValidatorWeight(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "updateValidatorWeight")
	if err != nil {
		return revertData, 0, err
	}

	// Get the validatorID and value from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Emit UpdateValidatorWeight event
	topics := []common.Hash{
		DriverAbi.Events["UpdateValidatorWeight"].ID,
		common.BigToHash(validatorID), // indexed parameter
	}

	// Pack the event data
	data, err := DriverAbi.Events["UpdateValidatorWeight"].Inputs.NonIndexed().Pack(value)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleUpdateValidatorPubkey updates the validator pubkey
func handleUpdateValidatorPubkey(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the backend
	revertData, err := checkOnlyBackend(evm, caller, "updateValidatorPubkey")
	if err != nil {
		return revertData, 0, err
	}

	// Get the validatorID and pubkey from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	pubkey, ok := args[1].([]byte)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Emit UpdateValidatorPubkey event
	topics := []common.Hash{
		DriverAbi.Events["UpdateValidatorPubkey"].ID,
		common.BigToHash(validatorID), // indexed parameter
	}

	// Pack the event data
	data, err := DriverAbi.Events["UpdateValidatorPubkey"].Inputs.NonIndexed().Pack(pubkey)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}
