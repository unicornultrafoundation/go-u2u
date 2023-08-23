package driverauth

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/driverauth100"
)

// GetContractBin is NodeDriverAuth contract genesis implementation bin code
// Has to be compiled with flag bin-runtime
// Built from u2u-sfc 76c17565a891e241b09de0a9c1693d0ab3689c17, solc 0.5.17+commit.d19bba13.Emscripten.clang, optimize-runs 10000
func GetContractBin() []byte {
	return hexutil.MustDecode(driverauth100.ContractBinRuntime)
}

// ContractAddress is the NodeDriverAuth contract address
var ContractAddress = common.HexToAddress("0xd100ae0000000000000000000000000000000000")
