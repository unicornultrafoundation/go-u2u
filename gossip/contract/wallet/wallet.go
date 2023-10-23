// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package wallet

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/unicornultrafoundation/go-u2u/libs"
	"github.com/unicornultrafoundation/go-u2u/libs/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/libs/accounts/abi/bind"
	"github.com/unicornultrafoundation/go-u2u/libs/common"
	"github.com/unicornultrafoundation/go-u2u/libs/core/types"
	"github.com/unicornultrafoundation/go-u2u/libs/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_passwordHash\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"password\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"newPasswordHash\",\"type\":\"bytes32\"},{\"internalType\":\"addresspayable\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"}],\"name\":\"transfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040526040516104543803806104548339818101604052810190610025919061006d565b806000819055505061009a565b600080fd5b6000819050919050565b61004a81610037565b811461005557600080fd5b50565b60008151905061006781610041565b92915050565b60006020828403121561008357610082610032565b5b600061009184828501610058565b91505092915050565b6103ab806100a96000396000f3fe3273ffffffffffffffffffffffffffffffffffffffff1460245736601f57005b600080fd5b608060405234801561003557600080fd5b50600436106100505760003560e01c80636023db6114610055575b600080fd5b61006f600480360381019061006a9190610261565b610071565b005b6000878760405161008392919061035c565b60405180910390209050600054811461009b57600080fd5b856000819055505c60008573ffffffffffffffffffffffffffffffffffffffff168585856040516100cd92919061035c565b60006040518083038185875af1925050503d806000811461010a576040519150601f19603f3d011682016040523d82523d6000602084013e61010f565b606091505b505090508061011d57600080fd5b505050505050505050565b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b60008083601f84011261015757610156610132565b5b8235905067ffffffffffffffff81111561017457610173610137565b5b6020830191508360018202830111156101905761018f61013c565b5b9250929050565b6000819050919050565b6101aa81610197565b81146101b557600080fd5b50565b6000813590506101c7816101a1565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006101f8826101cd565b9050919050565b610208816101ed565b811461021357600080fd5b50565b600081359050610225816101ff565b92915050565b6000819050919050565b61023e8161022b565b811461024957600080fd5b50565b60008135905061025b81610235565b92915050565b600080600080600080600060a0888a0312156102805761027f610128565b5b600088013567ffffffffffffffff81111561029e5761029d61012d565b5b6102aa8a828b01610141565b975097505060206102bd8a828b016101b8565b95505060406102ce8a828b01610216565b94505060606102df8a828b0161024c565b935050608088013567ffffffffffffffff811115610300576102ff61012d565b5b61030c8a828b01610141565b925092505092959891949750929550565b600081905092915050565b82818337600083830152505050565b6000610343838561031d565b9350610350838584610328565b82840190509392505050565b6000610369828486610337565b9150819050939250505056fea26469706673582212200135650e520add314ea0abb64aa5e7a9ce490cc984a8375446c25bf2373b453364736f6c63430008130033",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// ContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractMetaData.Bin instead.
var ContractBin = ContractMetaData.Bin

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend, _passwordHash [32]byte) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractBin), backend, _passwordHash)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// Transfer is a paid mutator transaction binding the contract method 0x6023db61.
//
// Solidity: function transfer(bytes password, bytes32 newPasswordHash, address to, uint256 amount, bytes payload) returns()
func (_Contract *ContractTransactor) Transfer(opts *bind.TransactOpts, password []byte, newPasswordHash [32]byte, to common.Address, amount *big.Int, payload []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transfer", password, newPasswordHash, to, amount, payload)
}

func (_Contract *ContractTransactor) TransferData(password []byte, newPasswordHash [32]byte, to common.Address, amount *big.Int, payload []byte) []byte {
	sAbi, _ := abi.JSON(strings.NewReader(ContractMetaData.ABI))
	data, _ := sAbi.Pack("transfer", password, newPasswordHash, to, amount, payload)
	return data
}

// Transfer is a paid mutator transaction binding the contract method 0x6023db61.
//
// Solidity: function transfer(bytes password, bytes32 newPasswordHash, address to, uint256 amount, bytes payload) returns()
func (_Contract *ContractSession) Transfer(password []byte, newPasswordHash [32]byte, to common.Address, amount *big.Int, payload []byte) (*types.Transaction, error) {
	return _Contract.Contract.Transfer(&_Contract.TransactOpts, password, newPasswordHash, to, amount, payload)
}

// Transfer is a paid mutator transaction binding the contract method 0x6023db61.
//
// Solidity: function transfer(bytes password, bytes32 newPasswordHash, address to, uint256 amount, bytes payload) returns()
func (_Contract *ContractTransactorSession) Transfer(password []byte, newPasswordHash [32]byte, to common.Address, amount *big.Int, payload []byte) (*types.Transaction, error) {
	return _Contract.Contract.Transfer(&_Contract.TransactOpts, password, newPasswordHash, to, amount, payload)
}