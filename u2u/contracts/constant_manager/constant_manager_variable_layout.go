package constant_manager

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
