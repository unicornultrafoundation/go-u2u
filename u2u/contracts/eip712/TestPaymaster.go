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

// TestPaymasterMetaData contains all meta data concerning the TestPaymaster contract.
var TestPaymasterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_context\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_txHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_suggestedSignedHash\",\"type\":\"bytes32\"},{\"internalType\":\"enumExecutionResult\",\"name\":\"_txResult\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"_maxRefundedGas\",\"type\":\"uint256\"}],\"name\":\"postTransaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_txHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_suggestedSignedHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_authenticationSignature\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"txType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"from\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"to\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasPerPubdataByteLimit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paymaster\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"reserved\",\"type\":\"uint256[4]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"paymasterInput\",\"type\":\"bytes\"}],\"internalType\":\"structTransaction\",\"name\":\"_transaction\",\"type\":\"tuple\"}],\"name\":\"validateAndPayForPaymasterTransaction\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"magic\",\"type\":\"bytes4\"},{\"internalType\":\"bytes\",\"name\":\"context\",\"type\":\"bytes\"}],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
	Bin: "0x61014060405234801561001157600080fd5b5060405161096c38038061096c83398101604081905261003091610191565b815160209283012081519183019190912060e08290526101008190524660a0818152604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f818801819052818301969096526060810194909452608080850193909352308483018190528151808603909301835260c09485019091528151919095012090529190915261012052600080546001600160a01b031916331790556101f4565b634e487b7160e01b600052604160045260246000fd5b600082601f8301126100fd57600080fd5b81516001600160401b0380821115610117576101176100d6565b604051601f8301601f19908116603f0116810190828211818310171561013f5761013f6100d6565b8160405283815260209250868385880101111561015b57600080fd5b600091505b8382101561017d5785820183015181830184015290820190610160565b600093810190920192909252949350505050565b600080604083850312156101a457600080fd5b82516001600160401b03808211156101bb57600080fd5b6101c7868387016100ec565b935060208501519150808211156101dd57600080fd5b506101ea858286016100ec565b9150509250929050565b60805160a05160c05160e051610100516101205161070061026c6000396000818161011001526103c7015260008181610164015261041601526000818161021d01526103f1015260008181610251015261034a01526000818160a801526103740152600081816101e9015261039e01526107006000f3fe6080604052600436106100915760003560e01c80638da5cb5b116100595780638da5cb5b146101865780639f3cb0b0146101be578063a9e91e54146101d7578063caac6c821461020b578063da28b5271461023f57600080fd5b80632b437d4814610096578063391c91d2146100dd5780635d2dab0b146100fe5780636b03e22214610132578063712ac56d14610152575b600080fd5b3480156100a257600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b6040519081526020015b60405180910390f35b6100f06100eb366004610493565b610273565b6040516100d492919061057c565b34801561010a57600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b34801561013e57600080fd5b506100ca61014d3660046105db565b61029b565b34801561015e57600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b34801561019257600080fd5b506000546101a6906001600160a01b031681565b6040516001600160a01b0390911681526020016100d4565b6101d56101cc366004610608565b50505050505050565b005b3480156101e357600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b34801561021757600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b34801561024b57600080fd5b506101a67f000000000000000000000000000000000000000000000000000000000000000081565b600060608482036102875760009150610292565b631c8e48e960e11b91505b94509492505050565b60006102f07f2bb156903cb7269fcabf8c90e4a20f3c4eaf1dde6b407dbc50a24ac21c9fab696040805160208101929092528101849052606001604051602081830303815290604052805190602001206102f6565b92915050565b600061030061033d565b60405161190160f01b6020820152602281019190915260428101839052606201604051602081830303815290604052805190602001209050919050565b6000306001600160a01b037f00000000000000000000000000000000000000000000000000000000000000001614801561039657507f000000000000000000000000000000000000000000000000000000000000000046145b156103c057507f000000000000000000000000000000000000000000000000000000000000000090565b50604080517f00000000000000000000000000000000000000000000000000000000000000006020808301919091527f0000000000000000000000000000000000000000000000000000000000000000828401527f000000000000000000000000000000000000000000000000000000000000000060608301524660808301523060a0808401919091528351808403909101815260c0909201909252805191012090565b634e487b7160e01b600052604160045260246000fd5b6000610220828403121561048d57600080fd5b50919050565b600080600080608085870312156104a957600080fd5b8435935060208501359250604085013567ffffffffffffffff808211156104cf57600080fd5b818701915087601f8301126104e357600080fd5b8135818111156104f5576104f5610464565b604051601f8201601f19908116603f0116810190838211818310171561051d5761051d610464565b816040528281528a602084870101111561053657600080fd5b82602086016020830137600060208483010152809650505050606087013591508082111561056357600080fd5b506105708782880161047a565b91505092959194509250565b63ffffffff60e01b8316815260006020604081840152835180604085015260005b818110156105b95785810183015185820160600152820161059d565b506000606082860101526060601f19601f830116850101925050509392505050565b6000602082840312156105ed57600080fd5b5035919050565b80356002811061060357600080fd5b919050565b600080600080600080600060c0888a03121561062357600080fd5b873567ffffffffffffffff8082111561063b57600080fd5b818a0191508a601f83011261064f57600080fd5b81358181111561065e57600080fd5b8b602082850101111561067057600080fd5b60209283019950975090890135908082111561068b57600080fd5b506106988a828b0161047a565b95505060408801359350606088013592506106b5608089016105f4565b915060a088013590509295989194975092955056fea2646970667358221220cfe90f67e26a4eecde3eb5c017d883a935bf1b6444733a7a4dd8e53e3805c9f364736f6c63430008140033",
}

// TestPaymasterABI is the input ABI used to generate the binding from.
// Deprecated: Use TestPaymasterMetaData.ABI instead.
var TestPaymasterABI = TestPaymasterMetaData.ABI

// TestPaymasterBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use TestPaymasterMetaData.Bin instead.
var TestPaymasterBin = TestPaymasterMetaData.Bin

// DeployTestPaymaster deploys a new Ethereum contract, binding an instance of TestPaymaster to it.
func DeployTestPaymaster(auth *bind.TransactOpts, backend bind.ContractBackend, name string, version string) (common.Address, *types.Transaction, *TestPaymaster, error) {
	parsed, err := TestPaymasterMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TestPaymasterBin), backend, name, version)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &TestPaymaster{TestPaymasterCaller: TestPaymasterCaller{contract: contract}, TestPaymasterTransactor: TestPaymasterTransactor{contract: contract}, TestPaymasterFilterer: TestPaymasterFilterer{contract: contract}}, nil
}

// TestPaymaster is an auto generated Go binding around an Ethereum contract.
type TestPaymaster struct {
	TestPaymasterCaller     // Read-only binding to the contract
	TestPaymasterTransactor // Write-only binding to the contract
	TestPaymasterFilterer   // Log filterer for contract events
}

// TestPaymasterCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestPaymasterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestPaymasterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestPaymasterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestPaymasterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestPaymasterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestPaymasterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestPaymasterSession struct {
	Contract     *TestPaymaster    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestPaymasterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestPaymasterCallerSession struct {
	Contract *TestPaymasterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// TestPaymasterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestPaymasterTransactorSession struct {
	Contract     *TestPaymasterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// TestPaymasterRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestPaymasterRaw struct {
	Contract *TestPaymaster // Generic contract binding to access the raw methods on
}

// TestPaymasterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestPaymasterCallerRaw struct {
	Contract *TestPaymasterCaller // Generic read-only contract binding to access the raw methods on
}

// TestPaymasterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestPaymasterTransactorRaw struct {
	Contract *TestPaymasterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestPaymaster creates a new instance of TestPaymaster, bound to a specific deployed contract.
func NewTestPaymaster(address common.Address, backend bind.ContractBackend) (*TestPaymaster, error) {
	contract, err := bindTestPaymaster(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TestPaymaster{TestPaymasterCaller: TestPaymasterCaller{contract: contract}, TestPaymasterTransactor: TestPaymasterTransactor{contract: contract}, TestPaymasterFilterer: TestPaymasterFilterer{contract: contract}}, nil
}

// NewTestPaymasterCaller creates a new read-only instance of TestPaymaster, bound to a specific deployed contract.
func NewTestPaymasterCaller(address common.Address, caller bind.ContractCaller) (*TestPaymasterCaller, error) {
	contract, err := bindTestPaymaster(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestPaymasterCaller{contract: contract}, nil
}

// NewTestPaymasterTransactor creates a new write-only instance of TestPaymaster, bound to a specific deployed contract.
func NewTestPaymasterTransactor(address common.Address, transactor bind.ContractTransactor) (*TestPaymasterTransactor, error) {
	contract, err := bindTestPaymaster(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestPaymasterTransactor{contract: contract}, nil
}

// NewTestPaymasterFilterer creates a new log filterer instance of TestPaymaster, bound to a specific deployed contract.
func NewTestPaymasterFilterer(address common.Address, filterer bind.ContractFilterer) (*TestPaymasterFilterer, error) {
	contract, err := bindTestPaymaster(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestPaymasterFilterer{contract: contract}, nil
}

// bindTestPaymaster binds a generic wrapper to an already deployed contract.
func bindTestPaymaster(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TestPaymasterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestPaymaster *TestPaymasterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestPaymaster.Contract.TestPaymasterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestPaymaster *TestPaymasterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestPaymaster.Contract.TestPaymasterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestPaymaster *TestPaymasterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestPaymaster.Contract.TestPaymasterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TestPaymaster *TestPaymasterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TestPaymaster.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TestPaymaster *TestPaymasterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TestPaymaster.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TestPaymaster *TestPaymasterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TestPaymaster.Contract.contract.Transact(opts, method, params...)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TestPaymaster *TestPaymasterCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TestPaymaster.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TestPaymaster *TestPaymasterSession) Owner() (common.Address, error) {
	return _TestPaymaster.Contract.Owner(&_TestPaymaster.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TestPaymaster *TestPaymasterCallerSession) Owner() (common.Address, error) {
	return _TestPaymaster.Contract.Owner(&_TestPaymaster.CallOpts)
}

// PostTransaction is a paid mutator transaction binding the contract method 0x9f3cb0b0.
//
// Solidity: function postTransaction(bytes _context, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction, bytes32 _txHash, bytes32 _suggestedSignedHash, uint8 _txResult, uint256 _maxRefundedGas) payable returns()
func (_TestPaymaster *TestPaymasterTransactor) PostTransaction(opts *bind.TransactOpts, _context []byte, _transaction Transaction, _txHash [32]byte, _suggestedSignedHash [32]byte, _txResult uint8, _maxRefundedGas *big.Int) (*types.Transaction, error) {
	return _TestPaymaster.contract.Transact(opts, "postTransaction", _context, _transaction, _txHash, _suggestedSignedHash, _txResult, _maxRefundedGas)
}

// PostTransaction is a paid mutator transaction binding the contract method 0x9f3cb0b0.
//
// Solidity: function postTransaction(bytes _context, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction, bytes32 _txHash, bytes32 _suggestedSignedHash, uint8 _txResult, uint256 _maxRefundedGas) payable returns()
func (_TestPaymaster *TestPaymasterSession) PostTransaction(_context []byte, _transaction Transaction, _txHash [32]byte, _suggestedSignedHash [32]byte, _txResult uint8, _maxRefundedGas *big.Int) (*types.Transaction, error) {
	return _TestPaymaster.Contract.PostTransaction(&_TestPaymaster.TransactOpts, _context, _transaction, _txHash, _suggestedSignedHash, _txResult, _maxRefundedGas)
}

// PostTransaction is a paid mutator transaction binding the contract method 0x9f3cb0b0.
//
// Solidity: function postTransaction(bytes _context, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction, bytes32 _txHash, bytes32 _suggestedSignedHash, uint8 _txResult, uint256 _maxRefundedGas) payable returns()
func (_TestPaymaster *TestPaymasterTransactorSession) PostTransaction(_context []byte, _transaction Transaction, _txHash [32]byte, _suggestedSignedHash [32]byte, _txResult uint8, _maxRefundedGas *big.Int) (*types.Transaction, error) {
	return _TestPaymaster.Contract.PostTransaction(&_TestPaymaster.TransactOpts, _context, _transaction, _txHash, _suggestedSignedHash, _txResult, _maxRefundedGas)
}

// ValidateAndPayForPaymasterTransaction is a paid mutator transaction binding the contract method 0x391c91d2.
//
// Solidity: function validateAndPayForPaymasterTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, bytes _authenticationSignature, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic, bytes context)
func (_TestPaymaster *TestPaymasterTransactor) ValidateAndPayForPaymasterTransaction(opts *bind.TransactOpts, _txHash [32]byte, _suggestedSignedHash [32]byte, _authenticationSignature []byte, _transaction Transaction) (*types.Transaction, error) {
	return _TestPaymaster.contract.Transact(opts, "validateAndPayForPaymasterTransaction", _txHash, _suggestedSignedHash, _authenticationSignature, _transaction)
}

// ValidateAndPayForPaymasterTransaction is a paid mutator transaction binding the contract method 0x391c91d2.
//
// Solidity: function validateAndPayForPaymasterTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, bytes _authenticationSignature, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic, bytes context)
func (_TestPaymaster *TestPaymasterSession) ValidateAndPayForPaymasterTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _authenticationSignature []byte, _transaction Transaction) (*types.Transaction, error) {
	return _TestPaymaster.Contract.ValidateAndPayForPaymasterTransaction(&_TestPaymaster.TransactOpts, _txHash, _suggestedSignedHash, _authenticationSignature, _transaction)
}

// ValidateAndPayForPaymasterTransaction is a paid mutator transaction binding the contract method 0x391c91d2.
//
// Solidity: function validateAndPayForPaymasterTransaction(bytes32 _txHash, bytes32 _suggestedSignedHash, bytes _authenticationSignature, (uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256[4],bytes,bytes,bytes) _transaction) payable returns(bytes4 magic, bytes context)
func (_TestPaymaster *TestPaymasterTransactorSession) ValidateAndPayForPaymasterTransaction(_txHash [32]byte, _suggestedSignedHash [32]byte, _authenticationSignature []byte, _transaction Transaction) (*types.Transaction, error) {
	return _TestPaymaster.Contract.ValidateAndPayForPaymasterTransaction(&_TestPaymaster.TransactOpts, _txHash, _suggestedSignedHash, _authenticationSignature, _transaction)
}
