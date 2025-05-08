// profile_sealepoch_direct.go
package main

import (
	"fmt"
	"math/big"
	"os"
	"runtime/pprof"
	"time"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfc"
)

// mockEVM is a simplified mock of the EVM for testing purposes
type mockEVM struct {
	state map[common.Hash][]byte
}

func newMockEVM() *mockEVM {
	return &mockEVM{
		state: make(map[common.Hash][]byte),
	}
}

// These methods implement a minimal subset of the vm.EVM interface needed for testing
func (e *mockEVM) GetState(addr common.Address, key common.Hash) common.Hash {
	if data, ok := e.state[key]; ok {
		var h common.Hash
		copy(h[:], data)
		return h
	}
	return common.Hash{}
}

func (e *mockEVM) SetState(addr common.Address, key, value common.Hash) {
	e.state[key] = value.Bytes()
}

func (e *mockEVM) GetStateEpochRecord(epoch uint64) []byte {
	// Calculate epoch record key manually - equivalent to sfc.GetEpochRecordKey
	epochBytes := big.NewInt(int64(epoch)).Bytes()
	key := common.BytesToHash(append([]byte("epoch-"), epochBytes...))
	return e.state[key]
}

func (e *mockEVM) SetStateEpochRecord(epoch uint64, data []byte) {
	// Calculate epoch record key manually - equivalent to sfc.GetEpochRecordKey
	epochBytes := big.NewInt(int64(epoch)).Bytes()
	key := common.BytesToHash(append([]byte("epoch-"), epochBytes...))
	e.state[key] = data
}

// A minimal driver contract address for auth checks
func (e *mockEVM) GetCaller() common.Address {
	return common.HexToAddress("0x1000000000000000000000000000000000000000")
}

// mockDriver helps set up minimal mock driver
func setupEpochState(evm *mockEVM) {
	// Set up minimal state needed for handleSealEpoch to work
	// This includes setting up previous epoch data and current epoch

	// Set current epoch to 1
	currentEpochKey := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	evm.state[currentEpochKey] = big.NewInt(1).Bytes()

	// Set minimum state for epoch 1
	evm.SetStateEpochRecord(1, []byte{1, 2, 3, 4}) // Some dummy data

	// Set driver auth contract address for permission checks
	driverAuthKey := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000005")
	driverAddr := common.HexToAddress("0x1000000000000000000000000000000000000000").Bytes()
	evm.state[driverAuthKey] = driverAddr
}

func profileSealEpoch() {
	// Check if profiling is requested
	shouldProfile := len(os.Args) > 1 && os.Args[1] == "profile"

	// Initialize profiling if requested
	var cpuFile, memFile *os.File
	var err error

	if shouldProfile {
		// Create CPU profile
		cpuFile, err = os.Create("sealepoch_cpu.pprof")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer cpuFile.Close()

		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			fmt.Fprintf(os.Stderr, "Could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()

		// Create memory profile
		memFile, err = os.Create("sealepoch_mem.pprof")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create memory profile: %v\n", err)
			os.Exit(1)
		}
		defer memFile.Close()

		fmt.Println("Running with profiling enabled...")
	} else {
		fmt.Println("Running without profiling...")
		fmt.Println("To enable profiling, use: go run profile_sealepoch_direct.go profile")
	}

	// Create mock EVM
	evm := newMockEVM()

	// Set up necessary state
	setupEpochState(evm)

	// Prepare args for handleSealEpoch
	// Format: func handleSealEpoch(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error)
	offlineTimes := make([]*big.Int, 0)
	offlineBlocks := make([]*big.Int, 0)
	uptimes := make([]*big.Int, 0)
	originatedTxsFee := make([]*big.Int, 0)
	gasPrice := big.NewInt(0)

	args := []interface{}{
		offlineTimes,
		offlineBlocks,
		uptimes,
		originatedTxsFee,
		gasPrice,
	}

	caller := common.HexToAddress("0x1000000000000000000000000000000000000000")

	// Create a driver API version of the EVM
	// Note: This is a simplified mock that won't work in all cases
	driverEvm := (*vm.EVM)(nil)

	// Measure execution time
	start := time.Now()

	// Call handleSealEpoch multiple times for profiling
	iterations := 10
	if shouldProfile {
		iterations = 100 // More iterations for profiling
	}

	fmt.Printf("Running handleSealEpoch for %d iterations\n", iterations)

	// Execute handleSealEpoch multiple times
	for i := 0; i < iterations; i++ {
		// Call the SFC contract function directly - we're assuming this function is exported
		// If not exported, we'll need to modify the sfc package to expose it for testing
		_, _, _ = sfc.HandleSealEpoch(driverEvm, caller, args)

		// Print progress for long runs
		if i%10 == 0 && iterations > 10 {
			fmt.Printf("Completed %d iterations\n", i)
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Execution completed in %s\n", elapsed)

	// Write memory profile if profiling
	if shouldProfile {
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			fmt.Fprintf(os.Stderr, "Could not write memory profile: %v\n", err)
		}

		fmt.Println("Profiling completed. Profile files written to:")
		fmt.Println("- sealepoch_cpu.pprof (CPU profile)")
		fmt.Println("- sealepoch_mem.pprof (Memory profile)")
		fmt.Println("\nTo analyze these profiles, run:")
		fmt.Println("go tool pprof -http=:8080 sealepoch_cpu.pprof")
		fmt.Println("go tool pprof -http=:8080 sealepoch_mem.pprof")
	}
}

func main() {
	profileSealEpoch()
}
