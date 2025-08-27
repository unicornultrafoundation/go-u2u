package constant_manager

import (
	"math/big"
	"sync"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// CMCache stores cached values from the ConstantManager contract
type CMCache struct {
	// Public variables stored in a map for convenience
	Values map[string]*big.Int
	// Owner address stored separately since it's not a big.Int
	Owner common.Address

	// Mutex for thread safety
	mutex sync.RWMutex

	// Invalidate cache signal
	NeedInvalidating bool
}

// Package-level cache instance
var cmCache = &CMCache{
	Values:           make(map[string]*big.Int),
	NeedInvalidating: true,
}

// GetCMCache returns the ConstantManager cache instance
func GetCMCache() *CMCache {
	return cmCache
}

// InvalidateCmCache invalidates the cache with the correct values from CM contract.
func InvalidateCmCache(evm *vm.EVM) {
	log.Info("Invalidating ConstantsManager cache...")
	cmCache.mutex.Lock()
	defer cmCache.mutex.Unlock()

	// Initialize the map if it doesn't exist
	if cmCache.Values == nil {
		cmCache.Values = make(map[string]*big.Int)
	}

	// Initialize all cache values
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minSelfStakeSlot)))
	cmCache.Values[MinSelfStakeKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(maxDelegatedRatioSlot)))
	cmCache.Values[MaxDelegatedRatioKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorCommissionSlot)))
	cmCache.Values[ValidatorCommissionKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(burntFeeShareSlot)))
	cmCache.Values[BurntFeeShareKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(treasuryFeeShareSlot)))
	cmCache.Values[TreasuryFeeShareKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(unlockedRewardRatioSlot)))
	cmCache.Values[UnlockedRewardRatioKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minLockupDurationSlot)))
	cmCache.Values[MinLockupDurationKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(maxLockupDurationSlot)))
	cmCache.Values[MaxLockupDurationKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalPeriodEpochsSlot)))
	cmCache.Values[WithdrawalPeriodEpochsKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalPeriodTimeSlot)))
	cmCache.Values[WithdrawalPeriodTimeKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(baseRewardPerSecondSlot)))
	cmCache.Values[BaseRewardPerSecondKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(offlinePenaltyThresholdBlocksNumSlot)))
	cmCache.Values[OfflinePenaltyThresholdBlocksNumKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(offlinePenaltyThresholdTimeSlot)))
	cmCache.Values[OfflinePenaltyThresholdTimeKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(targetGasPowerPerSecondSlot)))
	cmCache.Values[TargetGasPowerPerSecondKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(gasPriceBalancingCounterweightSlot)))
	cmCache.Values[GasPriceBalancingCounterweightKey] = val.Big()

	val = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	cmCache.Owner = common.BytesToAddress(val.Bytes())

	cmCache.NeedInvalidating = false
	log.Info("ConstantsManager cache invalidated successfully")
}

// UpdateCacheValue updates a specific value in the cache
// Only updates during real transactions, not during API calls (gas estimation, etc.)
func UpdateCacheValue(evm *vm.EVM, slot int64, value *big.Int) {
	// Skip cache updates for API calls (NoBaseFee indicates API call/simulation)
	if evm != nil && evm.Config.NoBaseFee {
		return
	}

	cmCache.mutex.Lock()
	defer cmCache.mutex.Unlock()

	// Initialize the map if it doesn't exist
	if cmCache.Values == nil {
		cmCache.Values = make(map[string]*big.Int)
	}

	switch slot {
	case minSelfStakeSlot:
		cmCache.Values[MinSelfStakeKey] = new(big.Int).Set(value)
	case maxDelegatedRatioSlot:
		cmCache.Values[MaxDelegatedRatioKey] = new(big.Int).Set(value)
	case validatorCommissionSlot:
		cmCache.Values[ValidatorCommissionKey] = new(big.Int).Set(value)
	case burntFeeShareSlot:
		cmCache.Values[BurntFeeShareKey] = new(big.Int).Set(value)
	case treasuryFeeShareSlot:
		cmCache.Values[TreasuryFeeShareKey] = new(big.Int).Set(value)
	case unlockedRewardRatioSlot:
		cmCache.Values[UnlockedRewardRatioKey] = new(big.Int).Set(value)
	case minLockupDurationSlot:
		cmCache.Values[MinLockupDurationKey] = new(big.Int).Set(value)
	case maxLockupDurationSlot:
		cmCache.Values[MaxLockupDurationKey] = new(big.Int).Set(value)
	case withdrawalPeriodEpochsSlot:
		cmCache.Values[WithdrawalPeriodEpochsKey] = new(big.Int).Set(value)
	case withdrawalPeriodTimeSlot:
		cmCache.Values[WithdrawalPeriodTimeKey] = new(big.Int).Set(value)
	case baseRewardPerSecondSlot:
		cmCache.Values[BaseRewardPerSecondKey] = new(big.Int).Set(value)
	case offlinePenaltyThresholdBlocksNumSlot:
		cmCache.Values[OfflinePenaltyThresholdBlocksNumKey] = new(big.Int).Set(value)
	case offlinePenaltyThresholdTimeSlot:
		cmCache.Values[OfflinePenaltyThresholdTimeKey] = new(big.Int).Set(value)
	case targetGasPowerPerSecondSlot:
		cmCache.Values[TargetGasPowerPerSecondKey] = new(big.Int).Set(value)
	case gasPriceBalancingCounterweightSlot:
		cmCache.Values[GasPriceBalancingCounterweightKey] = new(big.Int).Set(value)
	}
}

// UpdateOwner updates the owner address in the cache
// Only updates during real transactions, not during API calls (gas estimation, etc.)
func UpdateOwner(evm *vm.EVM, owner common.Address) {
	// Skip cache updates for API calls (NoBaseFee indicates API call/simulation)
	if evm != nil && evm.Config.NoBaseFee {
		return
	}

	cmCache.mutex.Lock()
	defer cmCache.mutex.Unlock()

	cmCache.Owner = owner
}

// Getter functions for cached values

// GetValue returns a cached value by key
func (c *CMCache) GetValue(key string) *big.Int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, exists := c.Values[key]
	if !exists || value == nil {
		log.Warn("ConstantsManager cache is nil", "key", key)
		return big.NewInt(0)
	}

	return new(big.Int).Set(value)
}

// GetOwner returns the cached owner of the contract
func (c *CMCache) GetOwner() common.Address {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Owner
}

// InitializeCMCacheAtStartup initializes the CM cache once at node startup
// This function creates a temporary EVM instance to call InvalidateCmCache
func InitializeCMCacheAtStartup(statedb vm.StateDB, sfcStatedb vm.StateDB, chainConfig *params.ChainConfig) {
	if !cmCache.NeedInvalidating {
		log.Info("CM cache already initialized, skipping")
		return
	}
	
	log.Info("Initializing CM cache at node startup...")
	
	// Create a temporary block context for initialization
	blockCtx := vm.BlockContext{
		CanTransfer: func(vm.StateDB, common.Address, *big.Int) bool { return true },
		Transfer:    func(vm.StateDB, common.Address, common.Address, *big.Int) {},
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		Coinbase:    common.Address{},
		BlockNumber: big.NewInt(0),
		Time:        big.NewInt(0),
		Difficulty:  big.NewInt(0),
		GasLimit:    0,
		BaseFee:     big.NewInt(0),
	}
	
	// Create a temporary transaction context for initialization
	txCtx := vm.TxContext{
		Origin:   common.HexToAddress("0xfc00face00000000000000000000000000000000"),
		GasPrice: big.NewInt(0),
	}
	
	// Create temporary EVM config for initialization (NoBaseFee = false for real transaction)
	config := vm.Config{NoBaseFee: false}
	
	// Create temporary EVM instance
	evm := vm.NewEVM(blockCtx, txCtx, statedb, sfcStatedb, chainConfig, config)
	
	// Initialize the cache
	InvalidateCmCache(evm)
}
