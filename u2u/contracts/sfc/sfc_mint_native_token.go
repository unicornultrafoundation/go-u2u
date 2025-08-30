package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// handleMintU2U mints U2U tokens
func handleMintU2U(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Check if caller is the owner
	revertData, checkGasUsed, err := checkOnlyOwner(evm, caller, "mintU2U")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 3 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	receiver, ok := args[0].(common.Address)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	amount, ok := args[1].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	justification, ok := args[2].(string)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Mint the native token to the contract itself
	mintArgs := []interface{}{
		ContractAddress,
		amount,
	}
	_, mintGasUsed, err := handle_mintNativeToken(evm, mintArgs)
	gasUsed += mintGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Transfer the tokens from the SFC contract to the receiver
	evm.SfcStateDB.SubBalance(ContractAddress, amount) // Subtract from SFC contract
	if SfcPrecompiles[receiver] {
		evm.SfcStateDB.AddBalance(receiver, amount) // Only maintain balance of sfc precompiled addresses
	}

	// Emit InflatedU2U event
	topics := []common.Hash{
		SfcLibAbi.Events["InflatedU2U"].ID,
		common.BytesToHash(common.LeftPadBytes(receiver.Bytes(), 32)), // indexed parameter (receiver)
	}
	data, err := SfcLibAbi.Events["InflatedU2U"].Inputs.NonIndexed().Pack(
		amount,
		justification,
	)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
	})

	return nil, gasUsed, nil
}

// _mintNativeToken is an internal function to mint native tokens
func handle_mintNativeToken(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the arguments
	if len(args) != 2 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	receiver, ok := args[0].(common.Address)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	amount, ok := args[1].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Implement the _mintNativeToken logic
	// 1. Call node.incBalance to increase the balance of the receiver
	nodeDriverAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(nodeDriverAuthSlot)))
	gasUsed += SloadGasCost
	nodeDriverAuthAddr := common.BytesToAddress(nodeDriverAuth.Bytes())

	// Pack the function call data for incBalance
	data, err := NodeDriverAuthAbi.Pack("incBalance", receiver, amount)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Call the node driver
	result, _, err := evm.CallSFC(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("SFC: Error calling NodeDriverAuth method", "method", "incBalance", "err", err, "reason", reason)
		return nil, gasUsed, err
	}

	// 2. Update the total supply
	totalSupply := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)))
	gasUsed += SloadGasCost
	totalSupplyBigInt := new(big.Int).SetBytes(totalSupply.Bytes())

	// Add the amount to the total supply
	newTotalSupply := new(big.Int).Add(totalSupplyBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)), common.BigToHash(newTotalSupply))
	gasUsed += SstoreGasCost

	return nil, gasUsed, nil
}
