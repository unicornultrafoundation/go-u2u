package sfclib

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/sfclib100"
)

// GetContractBin is SFCLib contract genesis implementation bin code
// Has to be compiled with flag bin-runtime
// Built from opera-sfc 76c17565a891e241b09de0a9c1693d0ab3689c17, solc 0.5.17+commit.d19bba13.Emscripten.clang, optimize-runs 200
func GetContractBin() []byte {
	return hexutil.MustDecode(sfclib100.ContractBinRuntime)
}

// ContractAddress is the SFCLib contract address
var ContractAddress = common.HexToAddress("0xfc01face00000000000000000000000000000000")
