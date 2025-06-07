package driverauth

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// handleInitialize initializes the NodeDriverAuth contract
func handleInitialize(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	log.Info("handleInitialize", "caller", caller.Hex(), "args", args)
	// Check if contract is already initialized
	sfc := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot)))
	driver := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot)))
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	emptyHash := common.Hash{}
	if sfc.Cmp(emptyHash) != 0 || driver.Cmp(emptyHash) != 0 || owner.Cmp(emptyHash) != 0 {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the sfc, driver, and owner addresses from args
	if len(args) != 3 {
		return nil, 0, vm.ErrExecutionReverted
	}
	sfcAddr, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	driverAddr, ok := args[1].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	ownerAddr, ok := args[2].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the sfc, driver, and owner addresses
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot)), sfcAddr.Hash())
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot)), driverAddr.Hash())
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), ownerAddr.Hash())

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		DriverAuthAbi.Events["OwnershipTransferred"].ID,
		emptyHash, // indexed parameter (previous owner - zero address)
		common.BytesToHash(common.LeftPadBytes(ownerAddr.Bytes(), 32)), // indexed parameter (new owner)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleMigrateTo migrates to a new driver auth contract
func handleMigrateTo(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the new driver auth address from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	newDriverAuth, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract to set the backend
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("setBackend", newDriverAuth)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleExecute executes a contract
func handleExecute(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the executable address from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	executable, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the code hash of this contract and the driver contract
	selfCodeHash := getCodeHash(evm, ContractAddress)
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())
	driverCodeHash := getCodeHash(evm, driverAddr)

	// Call mutExecute with the current owner
	return handleMutExecute(evm, caller, []interface{}{executable, common.BytesToAddress(owner.Bytes()), selfCodeHash, driverCodeHash})
}

// handleMutExecute executes a contract with a new owner
func handleMutExecute(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the arguments
	if len(args) != 4 {
		return nil, 0, vm.ErrExecutionReverted
	}
	executable, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	newOwner, ok := args[1].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	selfCodeHash, ok := args[2].(common.Hash)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	driverCodeHash, ok := args[3].(common.Hash)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Transfer ownership to the executable
	oldOwner := common.BytesToAddress(owner.Bytes())
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), executable.Hash())

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		DriverAuthAbi.Events["OwnershipTransferred"].ID,
		common.BytesToHash(common.LeftPadBytes(oldOwner.Bytes(), 32)),   // indexed parameter (previous owner)
		common.BytesToHash(common.LeftPadBytes(executable.Bytes(), 32)), // indexed parameter (new owner)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	// Call the executable contract's execute function
	execData, err := DriverAuthAbi.Pack("execute")
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), executable, execData, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	// Transfer ownership to the new owner
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), newOwner.Hash())

	// Emit OwnershipTransferred event
	topics = []common.Hash{
		DriverAuthAbi.Events["OwnershipTransferred"].ID,
		common.BytesToHash(common.LeftPadBytes(executable.Bytes(), 32)), // indexed parameter (previous owner)
		common.BytesToHash(common.LeftPadBytes(newOwner.Bytes(), 32)),   // indexed parameter (new owner)
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	// Check that the code hashes match
	currentSelfCodeHash := getCodeHash(evm, ContractAddress)
	if currentSelfCodeHash.Cmp(selfCodeHash) != 0 {
		return nil, 0, fmt.Errorf("self code hash doesn't match")
	}
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())
	currentDriverCodeHash := getCodeHash(evm, driverAddr)
	if currentDriverCodeHash.Cmp(driverCodeHash) != 0 {
		return nil, 0, fmt.Errorf("driver code hash doesn't match")
	}

	return nil, 0, nil
}

// handleIncBalance increments the balance of an account
func handleIncBalance(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the SFC contract
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())
	if caller != sfcAddr {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the account and diff from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	acc, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	diff, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the recipient is the SFC contract
	if acc != sfcAddr {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Calculate the new balance
	balance := new(big.Int).Add(evm.SfcStateDB.GetBalance(acc), diff)

	// Call the driver contract to set the balance
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("setBalance", acc, balance)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	result, _, err := evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("DriverAuth Precompiled: Error calling driver contract", "error", err, "method", "incBalance", "reason", reason)
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleUpgradeCode upgrades the code of a contract
func handleUpgradeCode(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the accounts from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	acc, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	from, ok := args[1].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that both addresses are contracts
	if !isContract(evm, acc) || !isContract(evm, from) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract to copy the code
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("copyCode", acc, from)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleCopyCode copies code from one account to another
func handleCopyCode(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the accounts from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	acc, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	from, ok := args[1].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract to copy the code
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("copyCode", acc, from)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleIncNonce increments the nonce of an account
func handleIncNonce(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the account and diff from args
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	acc, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	diff, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract to increment the nonce
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("incNonce", acc, diff)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleUpdateNetworkRules updates the network rules
func handleUpdateNetworkRules(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the diff from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	diff, ok := args[0].([]byte)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract to update the network rules
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("updateNetworkRules", diff)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleUpdateMinGasPrice updates the minimum gas price
func handleUpdateMinGasPrice(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the SFC contract
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())
	if caller != sfcAddr {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the minGasPrice from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	minGasPrice, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Format the JSON string
	jsonStr := fmt.Sprintf("{\"Economy\":{\"MinGasPrice\":%s}}", minGasPrice.String())

	// Call the driver contract to update the network rules
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("updateNetworkRules", []byte(jsonStr))
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleUpdateNetworkVersion updates the network version
func handleUpdateNetworkVersion(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the version from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	version, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract to update the network version
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("updateNetworkVersion", version)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleAdvanceEpochs advances the epochs
func handleAdvanceEpochs(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the num from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	num, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract to advance the epochs
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("advanceEpochs", num)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleUpdateValidatorWeight updates the validator weight
func handleUpdateValidatorWeight(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the SFC contract
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())
	if caller != sfcAddr {
		return nil, 0, vm.ErrExecutionReverted
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

	// Call the driver contract to update the validator weight
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("updateValidatorWeight", validatorID, value)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleUpdateValidatorPubkey updates the validator pubkey
func handleUpdateValidatorPubkey(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the SFC contract
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())
	if caller != sfcAddr {
		return nil, 0, vm.ErrExecutionReverted
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

	// Call the driver contract to update the validator pubkey
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("updateValidatorPubkey", validatorID, pubkey)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the driver contract
	_, _, err = evm.CallSFC(vm.AccountRef(ContractAddress), driverAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleTransferOwnership transfers ownership of the contract
func handleTransferOwnership(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the new owner from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	newOwner, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the new owner is not the zero address
	emptyAddr := common.Address{}
	if bytes.Equal(newOwner.Bytes(), emptyAddr.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the new owner
	oldOwner := common.BytesToAddress(owner.Bytes())
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), newOwner.Hash())

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		DriverAuthAbi.Events["OwnershipTransferred"].ID,
		common.BytesToHash(common.LeftPadBytes(oldOwner.Bytes(), 32)), // indexed parameter (previous owner)
		common.BytesToHash(common.LeftPadBytes(newOwner.Bytes(), 32)), // indexed parameter (new owner)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleRenounceOwnership renounces ownership of the contract
func handleRenounceOwnership(evm *vm.EVM, caller common.Address) ([]byte, uint64, error) {
	// Check if caller is the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if caller != common.BytesToAddress(owner.Bytes()) {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the owner to the zero address
	oldOwner := common.BytesToAddress(owner.Bytes())
	emptyHash := common.Hash{}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), emptyHash)

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		DriverAuthAbi.Events["OwnershipTransferred"].ID,
		common.BytesToHash(common.LeftPadBytes(oldOwner.Bytes(), 32)), // indexed parameter (previous owner)
		emptyHash, // indexed parameter (new owner - zero address)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// Helper functions

// isContract checks if an address is a contract
func isContract(evm *vm.EVM, addr common.Address) bool {
	code := evm.SfcStateDB.GetCode(addr)
	return len(code) > 0
}

// getCodeHash returns the code hash of an address
func getCodeHash(evm *vm.EVM, addr common.Address) common.Hash {
	return evm.SfcStateDB.GetCodeHash(addr)
}
