package constant_manager

import (
	"math/big"
	"sync"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
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
	Values: make(map[string]*big.Int),
}

// GetCMCache returns the ConstantManager cache instance
func GetCMCache() *CMCache {
	return cmCache
}

// InitCache initializes the cache with values from the contract
func InitCache(evm *vm.EVM) {
	log.Info("Initializing ConstantManager cache")
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
	log.Info("ConstantManager cache initialized successfully")
}

// UpdateCacheValue updates a specific value in the cache
func UpdateCacheValue(slot int64, value *big.Int) {
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

	cmCache.NeedInvalidating = true
}

// UpdateOwner updates the owner address in the cache
func UpdateOwner(owner common.Address) {
	cmCache.mutex.Lock()
	defer cmCache.mutex.Unlock()

	cmCache.Owner = owner
}

// Getter functions for cached values

// GetValue returns a cached value by key
func (c *CMCache) GetValue(key string) *big.Int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.Values == nil {
		return big.NewInt(0)
	}

	value, exists := c.Values[key]
	if !exists || value == nil {
		return big.NewInt(0)
	}

	return new(big.Int).Set(value)
}

// GetMinSelfStake returns the cached minimum amount of stake for a validator
func (c *CMCache) GetMinSelfStake() *big.Int {
	return c.GetValue(MinSelfStakeKey)
}

// GetMaxDelegatedRatio returns the cached maximum ratio of delegations a validator can have
func (c *CMCache) GetMaxDelegatedRatio() *big.Int {
	return c.GetValue(MaxDelegatedRatioKey)
}

// GetValidatorCommission returns the cached commission fee percentage a validator gets from delegations
func (c *CMCache) GetValidatorCommission() *big.Int {
	return c.GetValue(ValidatorCommissionKey)
}

// GetBurntFeeShare returns the cached percentage of fees to burn
func (c *CMCache) GetBurntFeeShare() *big.Int {
	return c.GetValue(BurntFeeShareKey)
}

// GetTreasuryFeeShare returns the cached percentage of fees to transfer to treasury
func (c *CMCache) GetTreasuryFeeShare() *big.Int {
	return c.GetValue(TreasuryFeeShareKey)
}

// GetUnlockedRewardRatio returns the cached ratio of reward rate at base rate (no lock)
func (c *CMCache) GetUnlockedRewardRatio() *big.Int {
	return c.GetValue(UnlockedRewardRatioKey)
}

// GetMinLockupDuration returns the cached minimum duration of stake/delegation lockup
func (c *CMCache) GetMinLockupDuration() *big.Int {
	return c.GetValue(MinLockupDurationKey)
}

// GetMaxLockupDuration returns the cached maximum duration of stake/delegation lockup
func (c *CMCache) GetMaxLockupDuration() *big.Int {
	return c.GetValue(MaxLockupDurationKey)
}

// GetWithdrawalPeriodEpochs returns the cached number of epochs undelegated stake is locked for
func (c *CMCache) GetWithdrawalPeriodEpochs() *big.Int {
	return c.GetValue(WithdrawalPeriodEpochsKey)
}

// GetWithdrawalPeriodTime returns the cached number of seconds undelegated stake is locked for
func (c *CMCache) GetWithdrawalPeriodTime() *big.Int {
	return c.GetValue(WithdrawalPeriodTimeKey)
}

// GetBaseRewardPerSecond returns the cached base reward per second
func (c *CMCache) GetBaseRewardPerSecond() *big.Int {
	return c.GetValue(BaseRewardPerSecondKey)
}

// GetOfflinePenaltyThresholdBlocksNum returns the cached threshold for offline penalty in blocks
func (c *CMCache) GetOfflinePenaltyThresholdBlocksNum() *big.Int {
	return c.GetValue(OfflinePenaltyThresholdBlocksNumKey)
}

// GetOfflinePenaltyThresholdTime returns the cached threshold for offline penalty in time
func (c *CMCache) GetOfflinePenaltyThresholdTime() *big.Int {
	return c.GetValue(OfflinePenaltyThresholdTimeKey)
}

// GetTargetGasPowerPerSecond returns the cached target gas power per second
func (c *CMCache) GetTargetGasPowerPerSecond() *big.Int {
	return c.GetValue(TargetGasPowerPerSecondKey)
}

// GetGasPriceBalancingCounterweight returns the cached gas price balancing counterweight
func (c *CMCache) GetGasPriceBalancingCounterweight() *big.Int {
	return c.GetValue(GasPriceBalancingCounterweightKey)
}

// GetOwner returns the cached owner of the contract
func (c *CMCache) GetOwner() common.Address {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Owner
}
