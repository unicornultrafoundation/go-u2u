package sfc

import (
	"math/big"
	"strings"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/sfc100"
)

var (
	SfcAbi             abi.ABI
	sfcContractAddress = ContractAddress
)

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

func init() {
	SfcAbi, _ = abi.JSON(strings.NewReader(sfc100.ContractMetaData.ABI))
}

// SfcPrecompile implements PrecompiledStateContract interface
type SfcPrecompile struct{}

// Run runs the precompiled contract
func (p *SfcPrecompile) Run(stateDB vm.StateDB, blockCtx vm.BlockContext, txCtx vm.TxContext, caller common.Address,
	input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	// TODO(trinhdn97): dynamic gas needed for each calls with different methods
	var gasUsed = uint64(3000) // Example fixed gas cost
	// Need at least 4 bytes for function signature
	if len(input) < 4 {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get function signature from first 4 bytes
	methodID := input[:4]
	method, err := SfcAbi.MethodById(methodID)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Parse input arguments
	args, err := method.Inputs.Unpack(input[4:])
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	var result []byte
	switch method.Name {
	case "owner":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(owner)))
		result, err = SfcAbi.Methods["owner"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
		// TODO(trinhdn97): calculate gas used after each method

	case "currentSealedEpoch":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
		result, err = SfcAbi.Methods["currentSealedEpoch"].Outputs.Pack(val.Big())

	case "lastValidatorID":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(lastValidatorIDSlot)))
		result, err = SfcAbi.Methods["lastValidatorID"].Outputs.Pack(val.Big())

	case "totalStake":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
		result, err = SfcAbi.Methods["totalStake"].Outputs.Pack(val.Big())

	case "totalActiveStake":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		result, err = SfcAbi.Methods["totalActiveStake"].Outputs.Pack(val.Big())

	case "totalSlashedStake":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(totalSlashedStakeSlot)))
		result, err = SfcAbi.Methods["totalSlashedStake"].Outputs.Pack(val.Big())

	case "totalSupply":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)))
		result, err = SfcAbi.Methods["totalSupply"].Outputs.Pack(val.Big())

	case "stakeTokenizerAddress":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(stakeTokenizerAddressSlot)))
		result, err = SfcAbi.Methods["stakeTokenizerAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))

	case "minGasPrice":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
		result, err = SfcAbi.Methods["minGasPrice"].Outputs.Pack(val.Big())

	case "treasuryAddress":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(treasuryAddressSlot)))
		result, err = SfcAbi.Methods["treasuryAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))

	case "voteBookAddress":
		val := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(voteBookAddressSlot)))
		result, err = SfcAbi.Methods["voteBookAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))

	case "getValidator":
		if len(args) != 1 {
			return nil, 0, vm.ErrExecutionReverted
		}
		validatorID := args[0].(*big.Int)
		key := common.BigToHash(validatorID)
		slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(validatorSlot)).Bytes())
		val := stateDB.GetState(sfcContractAddress, slot)
		result, err = SfcAbi.Methods["getValidator"].Outputs.Pack(val.Big())

	case "getValidatorID":
		if len(args) != 1 {
			return nil, 0, vm.ErrExecutionReverted
		}
		addr := args[0].(common.Address)
		key := addr.Hash()
		slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(validatorIDSlot)).Bytes())
		val := stateDB.GetState(sfcContractAddress, slot)
		result, err = SfcAbi.Methods["getValidatorID"].Outputs.Pack(val.Big())

	case "getValidatorPubkey":
		if len(args) != 1 {
			return nil, 0, vm.ErrExecutionReverted
		}
		validatorID := args[0].(*big.Int)
		key := common.BigToHash(validatorID)
		slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(validatorPubkeySlot)).Bytes())
		val := stateDB.GetState(sfcContractAddress, slot)
		result, err = SfcAbi.Methods["getValidatorPubkey"].Outputs.Pack(val.Bytes())

	case "stashedRewardsUntilEpoch":
		if len(args) != 2 {
			return nil, 0, vm.ErrExecutionReverted
		}
		addr := args[0].(common.Address)
		validatorID := args[1].(*big.Int)
		key1 := addr.Hash()
		slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(stashedRewardsUntilEpochSlot)).Bytes())
		key2 := common.BigToHash(validatorID)
		slot := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())
		val := stateDB.GetState(sfcContractAddress, slot)
		result, err = SfcAbi.Methods["stashedRewardsUntilEpoch"].Outputs.Pack(val.Big())

	case "getWithdrawalRequest":
		if len(args) != 3 {
			return nil, 0, vm.ErrExecutionReverted
		}
		addr := args[0].(common.Address)
		validatorID := args[1].(*big.Int)
		wrID := args[2].(*big.Int)
		key1 := addr.Hash()
		slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(withdrawalRequestSlot)).Bytes())
		key2 := common.BigToHash(validatorID)
		slot2 := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())
		key3 := common.BigToHash(wrID)
		slot := crypto.Keccak256Hash(key3.Bytes(), slot2.Bytes())

		epoch := stateDB.GetState(sfcContractAddress, slot)
		time := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(1))))
		amount := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(2))))

		result, err = SfcAbi.Methods["getWithdrawalRequest"].Outputs.Pack(
			epoch.Big(),
			time.Big(),
			amount.Big(),
		)

	case "getStake":
		if len(args) != 2 {
			return nil, 0, vm.ErrExecutionReverted
		}
		addr := args[0].(common.Address)
		validatorID := args[1].(*big.Int)
		key1 := addr.Hash()
		slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(stakeSlot)).Bytes())
		key2 := common.BigToHash(validatorID)
		slot := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())
		val := stateDB.GetState(sfcContractAddress, slot)
		result, err = SfcAbi.Methods["getStake"].Outputs.Pack(val.Big())

	case "getLockupInfo":
		if len(args) != 2 {
			return nil, 0, vm.ErrExecutionReverted
		}
		addr := args[0].(common.Address)
		validatorID := args[1].(*big.Int)
		key1 := addr.Hash()
		slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(lockupInfoSlot)).Bytes())
		key2 := common.BigToHash(validatorID)
		slot := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())

		lockedStake := stateDB.GetState(sfcContractAddress, slot)
		fromEpoch := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(1))))
		endTime := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(2))))
		duration := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(3))))

		result, err = SfcAbi.Methods["getLockupInfo"].Outputs.Pack(
			lockedStake.Big(),
			fromEpoch.Big(),
			endTime.Big(),
			duration.Big(),
		)

	case "getStashedLockupRewards":
		if len(args) != 2 {
			return nil, 0, vm.ErrExecutionReverted
		}
		addr := args[0].(common.Address)
		validatorID := args[1].(*big.Int)
		key1 := addr.Hash()
		slot1 := crypto.Keccak256Hash(key1.Bytes(), common.BigToHash(big.NewInt(stashedLockupRewardsSlot)).Bytes())
		key2 := common.BigToHash(validatorID)
		slot := crypto.Keccak256Hash(key2.Bytes(), slot1.Bytes())

		lockupBaseReward := stateDB.GetState(sfcContractAddress, slot)
		lockupExtraReward := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(1))))
		unlockedReward := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(2))))

		result, err = SfcAbi.Methods["getStashedLockupRewards"].Outputs.Pack(
			lockupBaseReward.Big(),
			lockupExtraReward.Big(),
			unlockedReward.Big(),
		)

	case "slashingRefundRatio":
		if len(args) != 1 {
			return nil, 0, vm.ErrExecutionReverted
		}
		validatorID := args[0].(*big.Int)
		key := common.BigToHash(validatorID)
		slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(slashingRefundRatioSlot)).Bytes())
		val := stateDB.GetState(sfcContractAddress, slot)
		result, err = SfcAbi.Methods["slashingRefundRatio"].Outputs.Pack(val.Big())

	case "getEpochSnapshot":
		if len(args) != 1 {
			return nil, 0, vm.ErrExecutionReverted
		}
		epoch := args[0].(*big.Int)
		key := common.BigToHash(epoch)
		slot := crypto.Keccak256Hash(key.Bytes(), common.BigToHash(big.NewInt(epochSnapshotSlot)).Bytes())

		endTime := stateDB.GetState(sfcContractAddress, slot)
		epochFee := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(1))))
		totalBaseRewardWeight := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(2))))
		totalTxRewardWeight := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(3))))
		baseRewardPerSecond := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(4))))
		totalStake := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(5))))
		totalSupply := stateDB.GetState(sfcContractAddress, common.BigToHash(big.NewInt(1).Add(slot.Big(), big.NewInt(6))))

		result, err = SfcAbi.Methods["getEpochSnapshot"].Outputs.Pack(
			endTime.Big(),
			epochFee.Big(),
			totalBaseRewardWeight.Big(),
			totalTxRewardWeight.Big(),
			baseRewardPerSecond.Big(),
			totalStake.Big(),
			totalSupply.Big(),
		)

	default:
		return nil, 0, vm.ErrExecutionReverted
	}
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	if suppliedGas < gasUsed {
		return nil, 0, vm.ErrOutOfGas
	}
	return result, gasUsed, nil
}
