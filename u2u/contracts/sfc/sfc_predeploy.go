package sfc

import (
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/hexutil"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/sfc100"
)

// GetContractBin is SFC contract genesis implementation bin code
// Has to be compiled with flag bin-runtime
// Built from u2u-sfc 76c17565a891e241b09de0a9c1693d0ab3689c17, solc 0.5.17+commit.d19bba13.Emscripten.clang, optimize-runs 200
func GetContractBin() []byte {
	return hexutil.MustDecode(sfc100.ContractBinRuntime)
}

// ContractAddress is the SFC contract address
var ContractAddress = common.HexToAddress("0xfc00face00000000000000000000000000000000")

const (
	ConstantManagerABIStr = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"constant":true,"inputs":[],"name":"baseRewardPerSecond","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"burntFeeShare","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"gasPriceBalancingCounterweight","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"initialize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"isOwner","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"maxDelegatedRatio","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"maxLockupDuration","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"minLockupDuration","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"minSelfStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"offlinePenaltyThresholdBlocksNum","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"offlinePenaltyThresholdTime","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"renounceOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"targetGasPowerPerSecond","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"treasuryFeeShare","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"unlockedRewardRatio","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateBaseRewardPerSecond","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateBurntFeeShare","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateGasPriceBalancingCounterweight","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateMaxDelegatedRatio","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateMaxLockupDuration","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateMinLockupDuration","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateMinSelfStake","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateOfflinePenaltyThresholdBlocksNum","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateOfflinePenaltyThresholdTime","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateTargetGasPowerPerSecond","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateTreasuryFeeShare","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateUnlockedRewardRatio","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateValidatorCommission","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateWithdrawalPeriodEpochs","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"v","type":"uint256"}],"name":"updateWithdrawalPeriodTime","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"validatorCommission","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"withdrawalPeriodEpochs","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"withdrawalPeriodTime","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`
	DriverABIStr          = `[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"num","type":"uint256"}],"name":"AdvanceEpochs","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes","name":"diff","type":"bytes"}],"name":"UpdateNetworkRules","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"version","type":"uint256"}],"name":"UpdateNetworkVersion","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"validatorID","type":"uint256"},{"indexed":false,"internalType":"bytes","name":"pubkey","type":"bytes"}],"name":"UpdateValidatorPubkey","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"validatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"weight","type":"uint256"}],"name":"UpdateValidatorWeight","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"backend","type":"address"}],"name":"UpdatedBackend","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"_backend","type":"address"}],"name":"setBackend","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_backend","type":"address"},{"internalType":"address","name":"_evmWriterAddress","type":"address"}],"name":"initialize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"acc","type":"address"},{"internalType":"uint256","name":"value","type":"uint256"}],"name":"setBalance","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"acc","type":"address"},{"internalType":"address","name":"from","type":"address"}],"name":"copyCode","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"acc","type":"address"},{"internalType":"address","name":"with","type":"address"}],"name":"swapCode","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"acc","type":"address"},{"internalType":"bytes32","name":"key","type":"bytes32"},{"internalType":"bytes32","name":"value","type":"bytes32"}],"name":"setStorage","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"acc","type":"address"},{"internalType":"uint256","name":"diff","type":"uint256"}],"name":"incNonce","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes","name":"diff","type":"bytes"}],"name":"updateNetworkRules","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"version","type":"uint256"}],"name":"updateNetworkVersion","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],"name":"advanceEpochs","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"validatorID","type":"uint256"},{"internalType":"uint256","name":"value","type":"uint256"}],"name":"updateValidatorWeight","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"validatorID","type":"uint256"},{"internalType":"bytes","name":"pubkey","type":"bytes"}],"name":"updateValidatorPubkey","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"_auth","type":"address"},{"internalType":"uint256","name":"validatorID","type":"uint256"},{"internalType":"bytes","name":"pubkey","type":"bytes"},{"internalType":"uint256","name":"status","type":"uint256"},{"internalType":"uint256","name":"createdEpoch","type":"uint256"},{"internalType":"uint256","name":"createdTime","type":"uint256"},{"internalType":"uint256","name":"deactivatedEpoch","type":"uint256"},{"internalType":"uint256","name":"deactivatedTime","type":"uint256"}],"name":"setGenesisValidator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"delegator","type":"address"},{"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"lockedStake","type":"uint256"},{"internalType":"uint256","name":"lockupFromEpoch","type":"uint256"},{"internalType":"uint256","name":"lockupEndTime","type":"uint256"},{"internalType":"uint256","name":"lockupDuration","type":"uint256"},{"internalType":"uint256","name":"earlyUnlockPenalty","type":"uint256"},{"internalType":"uint256","name":"rewards","type":"uint256"}],"name":"setGenesisDelegation","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"validatorID","type":"uint256"},{"internalType":"uint256","name":"status","type":"uint256"}],"name":"deactivateValidator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256[]","name":"nextValidatorIDs","type":"uint256[]"}],"name":"sealEpochValidators","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256[]","name":"offlineTimes","type":"uint256[]"},{"internalType":"uint256[]","name":"offlineBlocks","type":"uint256[]"},{"internalType":"uint256[]","name":"uptimes","type":"uint256[]"},{"internalType":"uint256[]","name":"originatedTxsFee","type":"uint256[]"}],"name":"sealEpoch","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256[]","name":"offlineTimes","type":"uint256[]"},{"internalType":"uint256[]","name":"offlineBlocks","type":"uint256[]"},{"internalType":"uint256[]","name":"uptimes","type":"uint256[]"},{"internalType":"uint256[]","name":"originatedTxsFee","type":"uint256[]"},{"internalType":"uint256","name":"usedGas","type":"uint256"}],"name":"sealEpochV1","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
)
