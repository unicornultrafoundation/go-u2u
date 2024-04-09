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
	Bin: "0x61014060405234801562000011575f80fd5b50604051620012d5380380620012d5833981810160405281019062000037919062000314565b81815f828051906020012090505f828051906020012090505f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f90508260e08181525050816101008181525050620000946200013e60201b60201c565b60a08181525050620000ae8184846200014560201b60201c565b608081815250503073ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff16815250508061012081815250505050505050335f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550505062000469565b5f46905090565b5f838383620001596200013e60201b60201c565b30604051602001620001709594939291906200040e565b6040516020818303038152906040528051906020012090509392505050565b5f604051905090565b5f80fd5b5f80fd5b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b620001f082620001a8565b810181811067ffffffffffffffff82111715620002125762000211620001b8565b5b80604052505050565b5f620002266200018f565b9050620002348282620001e5565b919050565b5f67ffffffffffffffff821115620002565762000255620001b8565b5b6200026182620001a8565b9050602081019050919050565b5f5b838110156200028d57808201518184015260208101905062000270565b5f8484015250505050565b5f620002ae620002a88462000239565b6200021b565b905082815260208101848484011115620002cd57620002cc620001a4565b5b620002da8482856200026e565b509392505050565b5f82601f830112620002f957620002f8620001a0565b5b81516200030b84826020860162000298565b91505092915050565b5f80604083850312156200032d576200032c62000198565b5b5f83015167ffffffffffffffff8111156200034d576200034c6200019c565b5b6200035b85828601620002e2565b925050602083015167ffffffffffffffff8111156200037f576200037e6200019c565b5b6200038d85828601620002e2565b9150509250929050565b5f819050919050565b620003ab8162000397565b82525050565b5f819050919050565b620003c581620003b1565b82525050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f620003f682620003cb565b9050919050565b6200040881620003ea565b82525050565b5f60a082019050620004235f830188620003a0565b620004326020830187620003a0565b620004416040830186620003a0565b620004506060830185620003ba565b6200045f6080830184620003fd565b9695505050505050565b60805160a05160c05160e0516101005161012051610e24620004b15f395f6104c701525f61050901525f6104e801525f61041601525f61046c01525f61049c0152610e245ff3fe608060405260043610610033575f3560e01c8063391c91d2146100375780638da5cb5b146100685780639f3cb0b014610092575b5f80fd5b610051600480360381019061004c9190610804565b6100ae565b60405161005f929190610954565b60405180910390f35b348015610073575f80fd5b5061007c610167565b60405161008991906109c1565b60405180910390f35b6100ac60048036038101906100a79190610a8d565b61018a565b005b5f60605f6100c46100be87610193565b866101d3565b90505f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614610153576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161014a90610bd3565b60405180910390fd5b63391c91d260e01b92505094509492505050565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b50505050505050565b5f6101cc61019f6101fd565b836040516020016101b1929190610c00565b60405160208183030381529060405280519060200120610224565b9050919050565b5f805f806101e1868661025c565b9250925092506101f182826102b1565b82935050505092915050565b5f7f2bb156903cb7269fcabf8c90e4a20f3c4eaf1dde6b407dbc50a24ac21c9fab69905090565b5f61022d610413565b8260405160200161023f929190610c9b565b604051602081830303815290604052805190602001209050919050565b5f805f604184510361029c575f805f602087015192506040870151915060608701515f1a905061028e88828585610533565b9550955095505050506102aa565b5f600285515f1b9250925092505b9250925092565b5f60038111156102c4576102c3610cd1565b5b8260038111156102d7576102d6610cd1565b5b031561040f57600160038111156102f1576102f0610cd1565b5b82600381111561030457610303610cd1565b5b0361033b576040517ff645eedf00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6002600381111561034f5761034e610cd1565b5b82600381111561036257610361610cd1565b5b036103a657805f1c6040517ffce698f700000000000000000000000000000000000000000000000000000000815260040161039d9190610d0d565b60405180910390fd5b6003808111156103b9576103b8610cd1565b5b8260038111156103cc576103cb610cd1565b5b0361040e57806040517fd78bce0c0000000000000000000000000000000000000000000000000000000081526004016104059190610d26565b60405180910390fd5b5b5050565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163073ffffffffffffffffffffffffffffffffffffffff1614801561049557507f000000000000000000000000000000000000000000000000000000000000000061049361061a565b145b156104c2577f00000000000000000000000000000000000000000000000000000000000000009050610530565b61052d7f00000000000000000000000000000000000000000000000000000000000000007f00000000000000000000000000000000000000000000000000000000000000007f0000000000000000000000000000000000000000000000000000000000000000610621565b90505b90565b5f805f7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0845f1c111561056f575f600385925092509250610610565b5f6001888888886040515f81526020016040526040516105929493929190610d5a565b6020604051602081039080840390855afa1580156105b2573d5f803e3d5ffd5b5050506020604051035190505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610603575f60015f801b93509350935050610610565b805f805f1b935093509350505b9450945094915050565b5f46905090565b5f83838361062d61061a565b30604051602001610642959493929190610d9d565b6040516020818303038152906040528051906020012090509392505050565b5f604051905090565b5f80fd5b5f80fd5b5f819050919050565b61068481610672565b811461068e575f80fd5b50565b5f8135905061069f8161067b565b92915050565b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6106f3826106ad565b810181811067ffffffffffffffff82111715610712576107116106bd565b5b80604052505050565b5f610724610661565b905061073082826106ea565b919050565b5f67ffffffffffffffff82111561074f5761074e6106bd565b5b610758826106ad565b9050602081019050919050565b828183375f83830152505050565b5f61078561078084610735565b61071b565b9050828152602081018484840111156107a1576107a06106a9565b5b6107ac848285610765565b509392505050565b5f82601f8301126107c8576107c76106a5565b5b81356107d8848260208601610773565b91505092915050565b5f80fd5b5f61022082840312156107fb576107fa6107e1565b5b81905092915050565b5f805f806080858703121561081c5761081b61066a565b5b5f61082987828801610691565b945050602061083a87828801610691565b935050604085013567ffffffffffffffff81111561085b5761085a61066e565b5b610867878288016107b4565b925050606085013567ffffffffffffffff8111156108885761088761066e565b5b610894878288016107e5565b91505092959194509250565b5f7fffffffff0000000000000000000000000000000000000000000000000000000082169050919050565b6108d4816108a0565b82525050565b5f81519050919050565b5f82825260208201905092915050565b5f5b838110156109115780820151818401526020810190506108f6565b5f8484015250505050565b5f610926826108da565b61093081856108e4565b93506109408185602086016108f4565b610949816106ad565b840191505092915050565b5f6040820190506109675f8301856108cb565b8181036020830152610979818461091c565b90509392505050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6109ab82610982565b9050919050565b6109bb816109a1565b82525050565b5f6020820190506109d45f8301846109b2565b92915050565b5f80fd5b5f80fd5b5f8083601f8401126109f7576109f66106a5565b5b8235905067ffffffffffffffff811115610a1457610a136109da565b5b602083019150836001820283011115610a3057610a2f6109de565b5b9250929050565b60028110610a43575f80fd5b50565b5f81359050610a5481610a37565b92915050565b5f819050919050565b610a6c81610a5a565b8114610a76575f80fd5b50565b5f81359050610a8781610a63565b92915050565b5f805f805f805f60c0888a031215610aa857610aa761066a565b5b5f88013567ffffffffffffffff811115610ac557610ac461066e565b5b610ad18a828b016109e2565b9750975050602088013567ffffffffffffffff811115610af457610af361066e565b5b610b008a828b016107e5565b9550506040610b118a828b01610691565b9450506060610b228a828b01610691565b9350506080610b338a828b01610a46565b92505060a0610b448a828b01610a79565b91505092959891949750929550565b5f82825260208201905092915050565b7f5061796d61737465723a20556e617574686f72697a6564207369676e617475725f8201527f6500000000000000000000000000000000000000000000000000000000000000602082015250565b5f610bbd602183610b53565b9150610bc882610b63565b604082019050919050565b5f6020820190508181035f830152610bea81610bb1565b9050919050565b610bfa81610672565b82525050565b5f604082019050610c135f830185610bf1565b610c206020830184610bf1565b9392505050565b5f81905092915050565b7f19010000000000000000000000000000000000000000000000000000000000005f82015250565b5f610c65600283610c27565b9150610c7082610c31565b600282019050919050565b5f819050919050565b610c95610c9082610672565b610c7b565b82525050565b5f610ca582610c59565b9150610cb18285610c84565b602082019150610cc18284610c84565b6020820191508190509392505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffd5b610d0781610a5a565b82525050565b5f602082019050610d205f830184610cfe565b92915050565b5f602082019050610d395f830184610bf1565b92915050565b5f60ff82169050919050565b610d5481610d3f565b82525050565b5f608082019050610d6d5f830187610bf1565b610d7a6020830186610d4b565b610d876040830185610bf1565b610d946060830184610bf1565b95945050505050565b5f60a082019050610db05f830188610bf1565b610dbd6020830187610bf1565b610dca6040830186610bf1565b610dd76060830185610cfe565b610de460808301846109b2565b969550505050505056fea2646970667358221220072ef40a10e6cee80466e8db32779bb3e99e9138e59300d541cb378055838be664736f6c63430008140033",
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
