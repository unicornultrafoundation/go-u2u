package constant_manager

import "math/big"

// Storage slots for ConstantManager contract variables
const (
	isInitialized                        int64 = 0x0
	ownerSlot                            int64 = 0x33
	offset                               int64 = 0x66 // Base offset for storage slots of ConstantsManager contract when implement Ownable contract
	minSelfStakeSlot                           = offset + 0
	maxDelegatedRatioSlot                      = offset + 1
	validatorCommissionSlot                    = offset + 2
	burntFeeShareSlot                          = offset + 3
	treasuryFeeShareSlot                       = offset + 4
	unlockedRewardRatioSlot                    = offset + 5
	minLockupDurationSlot                      = offset + 6
	maxLockupDurationSlot                      = offset + 7
	withdrawalPeriodEpochsSlot                 = offset + 8
	withdrawalPeriodTimeSlot                   = offset + 9
	baseRewardPerSecondSlot                    = offset + 10
	offlinePenaltyThresholdBlocksNumSlot       = offset + 11
	offlinePenaltyThresholdTimeSlot            = offset + 12
	targetGasPowerPerSecondSlot                = offset + 13
	gasPriceBalancingCounterweightSlot         = offset + 14
	// address private secondaryOwner_erased		- slot 15
)

// Storage slots for ConstantManager contract variables
var ConstantManagerSlots = map[string]*big.Int{
	"isInitialized":                  big.NewInt(isInitialized),
	"owner":                          big.NewInt(ownerSlot),
	"minSelfStake":                   big.NewInt(minSelfStakeSlot),
	"maxDelegatedRatio":              big.NewInt(maxDelegatedRatioSlot),
	"validatorCommission":            big.NewInt(validatorCommissionSlot),
	"burntFeeShare":                  big.NewInt(burntFeeShareSlot),
	"treasuryFeeShare":               big.NewInt(treasuryFeeShareSlot),
	"unlockedRewardRatio":            big.NewInt(unlockedRewardRatioSlot),
	"minLockupDuration":              big.NewInt(minLockupDurationSlot),
	"maxLockupDuration":              big.NewInt(maxLockupDurationSlot),
	"withdrawalPeriodEpochs":         big.NewInt(withdrawalPeriodEpochsSlot),
	"withdrawalPeriodTime":           big.NewInt(withdrawalPeriodTimeSlot),
	"baseRewardPerSecond":            big.NewInt(baseRewardPerSecondSlot),
	"offlinePenaltyThresholdBlocksNum": big.NewInt(offlinePenaltyThresholdBlocksNumSlot),
	"offlinePenaltyThresholdTime":    big.NewInt(offlinePenaltyThresholdTimeSlot),
	"targetGasPowerPerSecond":        big.NewInt(targetGasPowerPerSecondSlot),
	"gasPriceBalancingCounterweight": big.NewInt(gasPriceBalancingCounterweightSlot),
}

// Variable name constants for cache keys
const (
	MinSelfStakeKey                   = "minSelfStake"
	MaxDelegatedRatioKey              = "maxDelegatedRatio"
	ValidatorCommissionKey            = "validatorCommission"
	BurntFeeShareKey                  = "burntFeeShare"
	TreasuryFeeShareKey               = "treasuryFeeShare"
	UnlockedRewardRatioKey            = "unlockedRewardRatio"
	MinLockupDurationKey              = "minLockupDuration"
	MaxLockupDurationKey              = "maxLockupDuration"
	WithdrawalPeriodEpochsKey         = "withdrawalPeriodEpochs"
	WithdrawalPeriodTimeKey           = "withdrawalPeriodTime"
	BaseRewardPerSecondKey            = "baseRewardPerSecond"
	OfflinePenaltyThresholdBlocksNumKey = "offlinePenaltyThresholdBlocksNum"
	OfflinePenaltyThresholdTimeKey    = "offlinePenaltyThresholdTime"
	TargetGasPowerPerSecondKey        = "targetGasPowerPerSecond"
	GasPriceBalancingCounterweightKey = "gasPriceBalancingCounterweight"
)
