package sfc

// Storage layout of the SFC contract, declared as constants
const (
	isInitialized                int64 = 0x0
	owner                        int64 = 0x33
	offset                       int64 = 0x66        // Base offset for storage slots of SFC contract when implement SFCBase contract
	nodeDriverAuthSlot                 = 0 + offset  // NodeDriverAuth internal node
	currentSealedEpochSlot             = 1 + offset  // uint256 public currentSealedEpoch
	validatorSlot                      = 2 + offset  // mapping(uint256 => Validator) public getValidator
	validatorIDSlot                    = 3 + offset  // mapping(address => uint256) public getValidatorID
	validatorPubkeySlot                = 4 + offset  // mapping(uint256 => bytes) public getValidatorPubkey
	lastValidatorIDSlot                = 5 + offset  // uint256 public lastValidatorID
	totalStakeSlot                     = 6 + offset  // uint256 public totalStake
	totalActiveStakeSlot               = 7 + offset  // uint256 public totalActiveStake
	totalSlashedStakeSlot              = 8 + offset  // uint256 public totalSlashedStake
	rewardsStashSlot                   = 9 + offset  // mapping(address => mapping(uint256 => Rewards)) internal _rewardsStash
	stashedRewardsUntilEpochSlot       = 10 + offset // mapping(address => mapping(uint256 => uint256)) public stashedRewardsUntilEpoch
	withdrawalRequestSlot              = 11 + offset // mapping(address => mapping(uint256 => mapping(uint256 => WithdrawalRequest))) public getWithdrawalRequest
	stakeSlot                          = 12 + offset // mapping(address => mapping(uint256 => uint256)) public getStake
	lockupInfoSlot                     = 13 + offset // mapping(address => mapping(uint256 => LockedDelegation)) public getLockupInfo
	stashedLockupRewardsSlot           = 14 + offset // mapping(address => mapping(uint256 => Rewards)) public getStashedLockupRewards
	// uint256 private erased0                      - slot 15
	totalSupplySlot   = 16 + offset // uint256 public totalSupply
	epochSnapshotSlot = 17 + offset // mapping(uint256 => EpochSnapshot) public getEpochSnapshot
	// uint256 private erased1                      - slot 18
	// uint256 private erased2                      - slot 19
	slashingRefundRatioSlot   = 20 + offset // mapping(uint256 => uint256) public slashingRefundRatio
	stakeTokenizerAddressSlot = 21 + offset // address public stakeTokenizerAddress
	// uint256 private erased3                      - slot 22
	// uint256 private erased4                      - slot 23
	minGasPriceSlot      = 24 + offset // uint256 public minGasPrice
	treasuryAddressSlot  = 25 + offset // address public treasuryAddress
	libAddressSlot       = 26 + offset // address internal libAddress
	constantsManagerSlot = 27 + offset // ConstantsManager internal c
	voteBookAddressSlot  = 28 + offset // address public voteBookAddress
)
