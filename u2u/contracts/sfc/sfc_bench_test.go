package sfc

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// BenchmarkHandleSealEpoch benchmarks the handleSealEpoch function
func BenchmarkHandleSealEpoch(b *testing.B) {
	// Create a minimal EVM context for testing
	stateDB, _ := newTestState()
	evm := newTestEVM(stateDB)

	// Create test validator IDs and initialize values
	numValidators := 10
	validatorIDs := make([]*big.Int, numValidators)
	offlineTimes := make([]*big.Int, numValidators)
	offlineBlocks := make([]*big.Int, numValidators)
	uptimes := make([]*big.Int, numValidators)
	originatedTxsFee := make([]*big.Int, numValidators)

	for i := 0; i < numValidators; i++ {
		validatorIDs[i] = big.NewInt(int64(i + 1))
		offlineTimes[i] = big.NewInt(0)
		offlineBlocks[i] = big.NewInt(0)
		uptimes[i] = big.NewInt(100)
		originatedTxsFee[i] = big.NewInt(1000000)
	}

	// Set up initial state
	setupInitialState(evm, validatorIDs)

	// Create args for handleSealEpoch
	args := []interface{}{
		offlineTimes,
		offlineBlocks,
		uptimes,
		originatedTxsFee,
		big.NewInt(100000), // epochGas
	}

	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Create a fresh EVM for each iteration
		stateDB, _ := newTestState()
		evm := newTestEVM(stateDB)
		setupInitialState(evm, validatorIDs)

		// Run handleSealEpoch
		_, _, _ = handleSealEpoch(evm, common.Address{}, args)
	}
}

// BenchmarkGetEpochSnapshotSlot benchmarks the getEpochSnapshotSlot function
func BenchmarkGetEpochSnapshotSlot(b *testing.B) {
	// Create some epoch values to test
	epochs := []*big.Int{
		big.NewInt(1),
		big.NewInt(10),
		big.NewInt(100),
		big.NewInt(1000),
	}

	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		epoch := epochs[i%len(epochs)]
		_, _ = getEpochSnapshotSlot(epoch)
	}
}

// BenchmarkSealEpochRewards benchmarks the _sealEpoch_rewards function
func BenchmarkSealEpochRewards(b *testing.B) {
	// Create a minimal EVM context for testing
	stateDB, _ := newTestState()
	evm := newTestEVM(stateDB)

	// Create test validator IDs and initialize values
	numValidators := 10
	validatorIDs := make([]*big.Int, numValidators)
	uptimes := make([]*big.Int, numValidators)
	accumulatedOriginatedTxsFee := make([]*big.Int, numValidators)

	for i := 0; i < numValidators; i++ {
		validatorIDs[i] = big.NewInt(int64(i + 1))
		uptimes[i] = big.NewInt(100)
		accumulatedOriginatedTxsFee[i] = big.NewInt(1000000)
	}

	// Set up initial state
	setupInitialState(evm, validatorIDs)

	// Set up args
	epochDuration := big.NewInt(100)
	currentEpoch := big.NewInt(10)
	prevEpoch := big.NewInt(9)

	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Create a fresh EVM for each iteration
		stateDB, _ := newTestState()
		evm := newTestEVM(stateDB)
		setupInitialState(evm, validatorIDs)

		// Run _sealEpoch_rewards
		_, _ = _sealEpoch_rewards(evm, epochDuration, currentEpoch, prevEpoch, validatorIDs, uptimes, accumulatedOriginatedTxsFee)
	}
}

// newTestState creates a new stateDB for testing
func newTestState() (*vm.U2UTestState, error) {
	stateDB := vm.NewU2UTestState()
	return stateDB, nil
}

// newTestEVM creates a new EVM for testing
func newTestEVM(stateDB vm.StateDB) *vm.EVM {
	context := vm.Context{
		CanTransfer: func(vm.StateDB, common.Address, *big.Int) bool { return true },
		Transfer:    func(vm.StateDB, common.Address, common.Address, *big.Int) {},
		Time:        big.NewInt(1000),
	}

	return vm.NewEVM(context, stateDB, nil, nil)
}

// setupInitialState sets up the initial state for testing
func setupInitialState(evm *vm.EVM, validatorIDs []*big.Int) {
	// Set current epoch
	currentEpoch := big.NewInt(10)
	currentEpochBytes := common.BigToHash(currentEpoch)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(currentEpochSlot)), currentEpochBytes)

	// Set previous sealed epoch
	prevEpoch := big.NewInt(9)
	prevEpochBytes := common.BigToHash(prevEpoch)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)), prevEpochBytes)

	// Set up epoch snapshot for previous epoch
	prevEpochSnapshotSlot, _ := getEpochSnapshotSlot(prevEpoch)

	// Set end time for previous epoch
	endTimeSlot := new(big.Int).Add(prevEpochSnapshotSlot, big.NewInt(endTimeOffset))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(endTimeSlot), common.BigToHash(big.NewInt(900)))

	// Set up validator ID list for current epoch
	currentEpochSnapshotSlot, _ := getEpochSnapshotSlot(currentEpoch)

	// Store validator IDs length
	validatorIDsSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(validatorIDsOffset))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorIDsSlot), common.BigToHash(big.NewInt(int64(len(validatorIDs)))))

	// Calculate the base slot for validator IDs array
	validatorIDsBaseSlotBytes := common.Keccak256(common.BigToHash(validatorIDsSlot).Bytes())
	validatorIDsBaseSlot := new(big.Int).SetBytes(validatorIDsBaseSlotBytes)

	// Store each validator ID
	for i, validatorID := range validatorIDs {
		elementSlot := new(big.Int).Add(validatorIDsBaseSlot, big.NewInt(int64(i)))
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(elementSlot), common.BigToHash(validatorID))

		// Store validator auth - we'll just use a dummy address for testing
		validatorAuthSlot, _ := getValidatorAuthSlot(validatorID)
		dummyAuth := common.HexToAddress("0x1234567890123456789012345678901234567890")
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorAuthSlot), common.BytesToHash(dummyAuth.Bytes()))

		// Store validator received stake
		receivedStakeSlot, _ := getEpochValidatorReceivedStakeSlot(currentEpoch, validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(receivedStakeSlot), common.BigToHash(big.NewInt(1000000)))

		// Store previous accumulated originated txs fee
		innerHash := CreateAndHashOffsetSlot(accumulatedOriginatedTxsFeeOffset, prevEpochSnapshotSlot)
		outerHashInput := CreateNestedHashInput(validatorID, innerHash)
		outerHash := CachedKeccak256(outerHashInput)
		prevAccumulatedTxsFeeSlot := new(big.Int).SetBytes(outerHash)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(prevAccumulatedTxsFeeSlot), common.BigToHash(big.NewInt(500000)))

		// Store self stake for validator auth
		selfStakeSlot, _ := getStakeSlot(dummyAuth, validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(selfStakeSlot), common.BigToHash(big.NewInt(500000)))

		// Store locked stake for validator auth
		lockedStakeSlot, _ := getLockedStakeSlot(dummyAuth, validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockedStakeSlot), common.BigToHash(big.NewInt(250000)))

		// Store lockup duration for validator auth
		lockupDurationSlot, _ := getLockupDurationSlot(dummyAuth, validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupDurationSlot), common.BigToHash(big.NewInt(86400)))
	}

	// Set up total supply
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)), common.BigToHash(big.NewInt(1000000000)))

	// Set constants manager method outputs
	// These won't actually be called in tests since we're not mocking the callConstantManagerMethod
	// but we'll set them up for completeness
}
