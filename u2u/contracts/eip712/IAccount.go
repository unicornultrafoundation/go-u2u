// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package eip712

import (
	"errors"
	"math/big"
	"strings"

	u2u "github.com/unicornultrafoundation/go-u2u"
	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/accounts/abi/bind"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = u2u.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// Transaction is an auto generated low-level Go binding around an user-defined struct.
type Transaction struct {
	TxType                 *big.Int
	From                   *big.Int
	To                     *big.Int
	GasLimit               *big.Int
	GasPerPubdataByteLimit *big.Int
	MaxFeePerGas           *big.Int
	MaxPriorityFeePerGas   *big.Int
	Paymaster              *big.Int
	Nonce                  *big.Int
	Value                  *big.Int
	Reserved               [4]*big.Int
	Data                   []byte
	Signature              []byte
	PaymasterInput         []byte
}

// IAccountMetaData contains all meta data concerning the IAccount contract.
var IAccountMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_txHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_suggestedSignedHash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"}],\"name\":\"executeTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"}],\"name\":\"executeTransactionFromOutside\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_txHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_suggestedSignedHash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"}],\"name\":\"payForTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_txHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_possibleSignedHash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"}],\"name\":\"prepareForPaymaster\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_txHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_suggestedSignedHash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"}],\"name\":\"validateTransaction\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"magic\",\"type\":\"bytes4\"}],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// IAccountABI is the input ABI used to generate the binding from.
// Deprecated: Use IAccountMetaData.ABI instead.
var IAccountABI = IAccountMetaData.ABI

// IAccountBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use IAccountMetaData.Bin instead.
var IAccountBin = IAccountMetaData.Bin

// DeployIAccount deploys a new Ethereum contract, binding an instance of IAccount to it.
func DeployIAccount(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *IAccount, error) {
	parsed, err := IAccountMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(IAccountBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &IAccount{IAccountCaller: IAccountCaller{contract: contract}, IAccountTransactor: IAccountTransactor{contract: contract}, IAccountFilterer: IAccountFilterer{contract: contract}}, nil
}

// IAccount is an auto generated Go binding around an Ethereum contract.
type IAccount struct {
	IAccountCaller     // Read-only binding to the contract
	IAccountTransactor // Write-only binding to the contract
	IAccountFilterer   // Log filterer for contract events
}

// IAccountCaller is an auto generated read-only Go binding around an Ethereum contract.
type IAccountCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IAccountTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IAccountTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IAccountFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IAccountFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IAccountSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IAccountSession struct {
	Contract     *IAccount         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IAccountCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IAccountCallerSession struct {
	Contract *IAccountCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// IAccountTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IAccountTransactorSession struct {
	Contract     *IAccountTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// IAccountRaw is an auto generated low-level Go binding around an Ethereum contract.
type IAccountRaw struct {
	Contract *IAccount // Generic contract binding to access the raw methods on
}

// IAccountCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IAccountCallerRaw struct {
	Contract *IAccountCaller // Generic read-only contract binding to access the raw methods on
}

// IAccountTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IAccountTransactorRaw struct {
	Contract *IAccountTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIAccount creates a new instance of IAccount, bound to a specific deployed contract.
func NewIAccount(address common.Address, backend bind.ContractBackend) (*IAccount, error) {
	contract, err := bindIAccount(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IAccount{IAccountCaller: IAccountCaller{contract: contract}, IAccountTransactor: IAccountTransactor{contract: contract}, IAccountFilterer: IAccountFilterer{contract: contract}}, nil
}

// NewIAccountCaller creates a new read-only instance of IAccount, bound to a specific deployed contract.
func NewIAccountCaller(address common.Address, caller bind.ContractCaller) (*IAccountCaller, error) {
	contract, err := bindIAccount(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IAccountCaller{contract: contract}, nil
}

// NewIAccountTransactor creates a new write-only instance of IAccount, bound to a specific deployed contract.
func NewIAccountTransactor(address common.Address, transactor bind.ContractTransactor) (*IAccountTransactor, error) {
	contract, err := bindIAccount(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IAccountTransactor{contract: contract}, nil
}

// NewIAccountFilterer creates a new log filterer instance of IAccount, bound to a specific deployed contract.
func NewIAccountFilterer(address common.Address, filterer bind.ContractFilterer) (*IAccountFilterer, error) {
	contract, err := bindIAccount(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IAccountFilterer{contract: contract}, nil
}

// bindIAccount binds a generic wrapper to an already deployed contract.
func bindIAccount(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IAccountMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IAccount *IAccountRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IAccount.Contract.IAccountCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IAccount *IAccountRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IAccount.Contract.IAccountTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IAccount *IAccountRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IAccount.Contract.IAccountTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IAccount *IAccountCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IAccount.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IAccount *IAccountTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IAccount.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IAccount *IAccountTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IAccount.Contract.contract.Transact(opts, method, params...)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0x53e38206.
//
// Solidity: function executeTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountTransactor) ExecuteTransaction(opts *bind.TransactOpts, _txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.contract.Transact(opts, "executeTransaction", _txHash, _suggestedSignedHash, _transaction)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0x53e38206.
//
// Solidity: function executeTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountSession) ExecuteTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.ExecuteTransaction(&_IAccount.TransactOpts, _txHash, _suggestedSignedHash, _transaction)
}

// ExecuteTransaction is a paid mutator transaction binding the contract method 0x53e38206.
//
// Solidity: function executeTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountTransactorSession) ExecuteTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.ExecuteTransaction(&_IAccount.TransactOpts, _txHash, _suggestedSignedHash, _transaction)
}

// ExecuteTransactionFromOutside is a paid mutator transaction binding the contract method 0x494290c6.
//
// Solidity: function executeTransactionFromOutside((uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountTransactor) ExecuteTransactionFromOutside(opts *bind.TransactOpts, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.contract.Transact(opts, "executeTransactionFromOutside", _transaction)
}

// ExecuteTransactionFromOutside is a paid mutator transaction binding the contract method 0x494290c6.
//
// Solidity: function executeTransactionFromOutside((uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountSession) ExecuteTransactionFromOutside(_transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.ExecuteTransactionFromOutside(&_IAccount.TransactOpts, _transaction)
}

// ExecuteTransactionFromOutside is a paid mutator transaction binding the contract method 0x494290c6.
//
// Solidity: function executeTransactionFromOutside((uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountTransactorSession) ExecuteTransactionFromOutside(_transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.ExecuteTransactionFromOutside(&_IAccount.TransactOpts, _transaction)
}

// PayForTransaction is a paid mutator transaction binding the contract method 0xa0b4c91a.
//
// Solidity: function payForTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountTransactor) PayForTransaction(opts *bind.TransactOpts, _txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.contract.Transact(opts, "payForTransaction", _txHash, _suggestedSignedHash, _transaction)
}

// PayForTransaction is a paid mutator transaction binding the contract method 0xa0b4c91a.
//
// Solidity: function payForTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountSession) PayForTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.PayForTransaction(&_IAccount.TransactOpts, _txHash, _suggestedSignedHash, _transaction)
}

// PayForTransaction is a paid mutator transaction binding the contract method 0xa0b4c91a.
//
// Solidity: function payForTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountTransactorSession) PayForTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.PayForTransaction(&_IAccount.TransactOpts, _txHash, _suggestedSignedHash, _transaction)
}

// PrepareForPaymaster is a paid mutator transaction binding the contract method 0x07192996.
//
// Solidity: function prepareForPaymaster(bytes32 _txHash, bytes32 _possibleSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountTransactor) PrepareForPaymaster(opts *bind.TransactOpts, _txHash [32]byte, _possibleSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.contract.Transact(opts, "prepareForPaymaster", _txHash, _possibleSignedHash, _transaction)
}

// PrepareForPaymaster is a paid mutator transaction binding the contract method 0x07192996.
//
// Solidity: function prepareForPaymaster(bytes32 _txHash, bytes32 _possibleSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountSession) PrepareForPaymaster(_txHash [32]byte, _possibleSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.PrepareForPaymaster(&_IAccount.TransactOpts, _txHash, _possibleSignedHash, _transaction)
}

// PrepareForPaymaster is a paid mutator transaction binding the contract method 0x07192996.
//
// Solidity: function prepareForPaymaster(bytes32 _txHash, bytes32 _possibleSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns()
func (_IAccount *IAccountTransactorSession) PrepareForPaymaster(_txHash [32]byte, _possibleSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.PrepareForPaymaster(&_IAccount.TransactOpts, _txHash, _possibleSignedHash, _transaction)
}

// ValidateTransaction is a paid mutator transaction binding the contract method 0x9c27706d.
//
// Solidity: function validateTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic)
func (_IAccount *IAccountTransactor) ValidateTransaction(opts *bind.TransactOpts, _txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.contract.Transact(opts, "validateTransaction", _txHash, _suggestedSignedHash, _transaction)
}

// ValidateTransaction is a paid mutator transaction binding the contract method 0x9c27706d.
//
// Solidity: function validateTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic)
func (_IAccount *IAccountSession) ValidateTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.ValidateTransaction(&_IAccount.TransactOpts, _txHash, _suggestedSignedHash, _transaction)
}

// ValidateTransaction is a paid mutator transaction binding the contract method 0x9c27706d.
//
// Solidity: function validateTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic)
func (_IAccount *IAccountTransactorSession) ValidateTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _transaction Transaction) (*types.Transaction, error) {
	return _IAccount.Contract.ValidateTransaction(&_IAccount.TransactOpts, _txHash, _suggestedSignedHash, _transaction)
}
