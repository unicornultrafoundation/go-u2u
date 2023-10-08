package sfclib

import (
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/sfclib100"
	"github.com/unicornultrafoundation/go-u2u/libs/common"
	"github.com/unicornultrafoundation/go-u2u/libs/common/hexutil"
)

// GetContractBin is SFCLib contract genesis implementation bin code
// Has to be compiled with flag bin-runtime
func GetContractBin() []byte {
	return hexutil.MustDecode(sfclib100.ContractBinRuntime)
}

// ContractAddress is the SFCLib contract address
var ContractAddress = common.HexToAddress("0xfc01face00000000000000000000000000000000")
