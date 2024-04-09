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

//// Transaction is an auto generated low-level Go binding around an user-defined struct.
//type Transaction struct {
//	TxType                 *big.Int
//	From                   *big.Int
//	To                     *big.Int
//	GasLimit               *big.Int
//	GasPerPubdataByteLimit *big.Int
//	MaxFeePerGas           *big.Int
//	MaxPriorityFeePerGas   *big.Int
//	Paymaster              *big.Int
//	Nonce                  *big.Int
//	Value                  *big.Int
//	Reserved               [4]*big.Int
//	Data                   []byte
//	Signature              []byte
//	PaymasterInput         []byte
//}

// IPaymasterMetaData contains all meta data concerning the IPaymaster contract.
var IPaymasterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_context\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_txHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_suggestedSignedHash\",\"type\":\"bytes32\"},{\"internalType\":\"enumExecutionResult\",\"name\":\"_txResult\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"_maxRefundedGas\",\"type\":\"uint256\"}],\"name\":\"postTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_txHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_suggestedSignedHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_authenticationSignature\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"}],\"name\":\"validateAndPayForPaymasterTransaction\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"magic\",\"type\":\"bytes4\"},{\"internalType\":\"bytes\",\"name\":\"context\",\"type\":\"bytes\"}],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// IPaymasterABI is the input ABI used to generate the binding from.
// Deprecated: Use IPaymasterMetaData.ABI instead.
var IPaymasterABI = IPaymasterMetaData.ABI

// IPaymaster is an auto generated Go binding around an Ethereum contract.
type IPaymaster struct {
	IPaymasterCaller     // Read-only binding to the contract
	IPaymasterTransactor // Write-only binding to the contract
	IPaymasterFilterer   // Log filterer for contract events
}

// IPaymasterCaller is an auto generated read-only Go binding around an Ethereum contract.
type IPaymasterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IPaymasterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IPaymasterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IPaymasterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IPaymasterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IPaymasterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IPaymasterSession struct {
	Contract     *IPaymaster       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IPaymasterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IPaymasterCallerSession struct {
	Contract *IPaymasterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// IPaymasterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IPaymasterTransactorSession struct {
	Contract     *IPaymasterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// IPaymasterRaw is an auto generated low-level Go binding around an Ethereum contract.
type IPaymasterRaw struct {
	Contract *IPaymaster // Generic contract binding to access the raw methods on
}

// IPaymasterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IPaymasterCallerRaw struct {
	Contract *IPaymasterCaller // Generic read-only contract binding to access the raw methods on
}

// IPaymasterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IPaymasterTransactorRaw struct {
	Contract *IPaymasterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIPaymaster creates a new instance of IPaymaster, bound to a specific deployed contract.
func NewIPaymaster(address common.Address, backend bind.ContractBackend) (*IPaymaster, error) {
	contract, err := bindIPaymaster(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IPaymaster{IPaymasterCaller: IPaymasterCaller{contract: contract}, IPaymasterTransactor: IPaymasterTransactor{contract: contract}, IPaymasterFilterer: IPaymasterFilterer{contract: contract}}, nil
}

// NewIPaymasterCaller creates a new read-only instance of IPaymaster, bound to a specific deployed contract.
func NewIPaymasterCaller(address common.Address, caller bind.ContractCaller) (*IPaymasterCaller, error) {
	contract, err := bindIPaymaster(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IPaymasterCaller{contract: contract}, nil
}

// NewIPaymasterTransactor creates a new write-only instance of IPaymaster, bound to a specific deployed contract.
func NewIPaymasterTransactor(address common.Address, transactor bind.ContractTransactor) (*IPaymasterTransactor, error) {
	contract, err := bindIPaymaster(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IPaymasterTransactor{contract: contract}, nil
}

// NewIPaymasterFilterer creates a new log filterer instance of IPaymaster, bound to a specific deployed contract.
func NewIPaymasterFilterer(address common.Address, filterer bind.ContractFilterer) (*IPaymasterFilterer, error) {
	contract, err := bindIPaymaster(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IPaymasterFilterer{contract: contract}, nil
}

// bindIPaymaster binds a generic wrapper to an already deployed contract.
func bindIPaymaster(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IPaymasterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IPaymaster *IPaymasterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IPaymaster.Contract.IPaymasterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IPaymaster *IPaymasterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IPaymaster.Contract.IPaymasterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IPaymaster *IPaymasterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IPaymaster.Contract.IPaymasterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IPaymaster *IPaymasterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IPaymaster.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IPaymaster *IPaymasterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IPaymaster.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IPaymaster *IPaymasterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IPaymaster.Contract.contract.Transact(opts, method, params...)
}

// PostTransaction is a paid mutator transaction binding the contract method 0x9f3cb0b0.
//
// Solidity: function postTransaction(bytes _context, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction, bytes32 _txHash, bytes32 _suggestedSignedHash, uint8 _txResult, uint256 _maxRefundedGas) payable returns()
func (_IPaymaster *IPaymasterTransactor) PostTransaction(opts *bind.TransactOpts, _context []byte, _transaction Transaction, _txHash [32]byte, _suggestedSignedHash [32]byte, _txResult uint8, _maxRefundedGas *big.Int) (*types.Transaction, error) {
	return _IPaymaster.contract.Transact(opts, "postTransaction", _context, _transaction, _txHash, _suggestedSignedHash, _txResult, _maxRefundedGas)
}

// PostTransaction is a paid mutator transaction binding the contract method 0x9f3cb0b0.
//
// Solidity: function postTransaction(bytes _context, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction, bytes32 _txHash, bytes32 _suggestedSignedHash, uint8 _txResult, uint256 _maxRefundedGas) payable returns()
func (_IPaymaster *IPaymasterSession) PostTransaction(_context []byte, _transaction Transaction, _txHash [32]byte, _suggestedSignedHash [32]byte, _txResult uint8, _maxRefundedGas *big.Int) (*types.Transaction, error) {
	return _IPaymaster.Contract.PostTransaction(&_IPaymaster.TransactOpts, _context, _transaction, _txHash, _suggestedSignedHash, _txResult, _maxRefundedGas)
}

// PostTransaction is a paid mutator transaction binding the contract method 0x9f3cb0b0.
//
// Solidity: function postTransaction(bytes _context, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction, bytes32 _txHash, bytes32 _suggestedSignedHash, uint8 _txResult, uint256 _maxRefundedGas) payable returns()
func (_IPaymaster *IPaymasterTransactorSession) PostTransaction(_context []byte, _transaction Transaction, _txHash [32]byte, _suggestedSignedHash [32]byte, _txResult uint8, _maxRefundedGas *big.Int) (*types.Transaction, error) {
	return _IPaymaster.Contract.PostTransaction(&_IPaymaster.TransactOpts, _context, _transaction, _txHash, _suggestedSignedHash, _txResult, _maxRefundedGas)
}

// ValidateAndPayForPaymasterTransaction is a paid mutator transaction binding the contract method 0x391c91d2.
//
// Solidity: function validateAndPayForPaymasterTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, bytes _authenticationSignature, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic, bytes context)
func (_IPaymaster *IPaymasterTransactor) ValidateAndPayForPaymasterTransaction(opts *bind.TransactOpts, _txHash [32]byte, _suggestedSignedHash [32]byte, _authenticationSignature []byte, _transaction Transaction) (*types.Transaction, error) {
	return _IPaymaster.contract.Transact(opts, "validateAndPayForPaymasterTransaction", _txHash, _suggestedSignedHash, _authenticationSignature, _transaction)
}

// ValidateAndPayForPaymasterTransaction is a paid mutator transaction binding the contract method 0x391c91d2.
//
// Solidity: function validateAndPayForPaymasterTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, bytes _authenticationSignature, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic, bytes context)
func (_IPaymaster *IPaymasterSession) ValidateAndPayForPaymasterTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _authenticationSignature []byte, _transaction Transaction) (*types.Transaction, error) {
	return _IPaymaster.Contract.ValidateAndPayForPaymasterTransaction(&_IPaymaster.TransactOpts, _txHash, _suggestedSignedHash, _authenticationSignature, _transaction)
}

// ValidateAndPayForPaymasterTransaction is a paid mutator transaction binding the contract method 0x391c91d2.
//
// Solidity: function validateAndPayForPaymasterTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, bytes _authenticationSignature, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic, bytes context)
func (_IPaymaster *IPaymasterTransactorSession) ValidateAndPayForPaymasterTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _authenticationSignature []byte, _transaction Transaction) (*types.Transaction, error) {
	return _IPaymaster.Contract.ValidateAndPayForPaymasterTransaction(&_IPaymaster.TransactOpts, _txHash, _suggestedSignedHash, _authenticationSignature, _transaction)
}
