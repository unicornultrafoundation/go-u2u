package netinit

import (
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/netinit100"
	"github.com/unicornultrafoundation/go-u2u/libs/common"
	"github.com/unicornultrafoundation/go-u2u/libs/common/hexutil"
)

// GetContractBin is NetworkInitializer contract genesis implementation bin code
// Has to be compiled with flag bin-runtime
// Built from u2u-sfc 76c17565a891e241b09de0a9c1693d0ab3689c17, solc 0.5.17+commit.d19bba13.Emscripten.clang, optimize-runs 10000
func GetContractBin() []byte {
	return hexutil.MustDecode(netinit100.ContractBinRuntime)
}

// ContractAddress is the NetworkInitializer contract address
var ContractAddress = common.HexToAddress("0xd1005eed00000000000000000000000000000000")
