// benchmark_sealepoch.go
package main

import (
	"fmt"
	"math/big"
	"os"
	"runtime/pprof"
	"time"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfc"
)

func main() {
	// Check if profiling is requested
	if len(os.Args) > 1 && os.Args[1] == "profile" {
		// Create CPU profile
		cpuFile, err := os.Create("cpu_profile.pprof")
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
		memFile, err := os.Create("mem_profile.pprof")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create memory profile: %v\n", err)
			os.Exit(1)
		}
		defer memFile.Close()

		// Run benchmark with profiles
		fmt.Println("Running with profiling enabled...")
		benchmarkEpochSnapshot(10)

		// Write memory profile
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			fmt.Fprintf(os.Stderr, "Could not write memory profile: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Run without profiling
		fmt.Println("Running benchmark without profiling...")
		fmt.Println("To enable profiling, use: go run benchmark_sealepoch.go profile")
		benchmarkEpochSnapshot(100)
	}
}

// benchmarkEpochSnapshot benchmarks the getEpochSnapshotSlot function
func benchmarkEpochSnapshot(iterations int) {
	// Create a range of epoch numbers to test
	epochs := []*big.Int{
		big.NewInt(1),
		big.NewInt(100),
		big.NewInt(1000),
		big.NewInt(10000),
		big.NewInt(100000),
	}

	// Create variables to store timing data
	totalTime := time.Duration(0)
	totalGas := uint64(0)
	totalDirectTime := time.Duration(0)

	fmt.Println("\nBenchmarking epoch snapshot calculations:")
	fmt.Println("======================================")
	fmt.Printf("Running %d iterations for each epoch value\n\n", iterations)
	fmt.Println("Epoch Number | Cached (μs) | Direct (μs) | Gas Used")
	fmt.Println("--------------------------------------------------")

	// Run benchmark for each epoch
	for _, epoch := range epochs {
		// Clear caches between epoch tests
		sfc.ClearCache()

		var cachedTime time.Duration
		var directTime time.Duration
		var gasUsed uint64

		// Benchmark the cached version
		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, gas := sfc.GetCachedEpochSnapshotSlot(epoch)
			elapsed := time.Since(start)
			cachedTime += elapsed
			gasUsed += gas
		}

		// Benchmark the direct calculation
		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, _ = getEpochSnapshotSlotDirect(epoch)
			elapsed := time.Since(start)
			directTime += elapsed
		}

		// Calculate averages
		avgCachedTime := cachedTime / time.Duration(iterations)
		avgDirectTime := directTime / time.Duration(iterations)
		avgGas := gasUsed / uint64(iterations)

		// Print results for this epoch
		fmt.Printf("%-12s | %-11s | %-11s | %d\n",
			epoch.String(),
			fmt.Sprintf("%.2f", float64(avgCachedTime.Nanoseconds())/1000.0),
			fmt.Sprintf("%.2f", float64(avgDirectTime.Nanoseconds())/1000.0),
			avgGas)

		// Accumulate totals
		totalTime += avgCachedTime
		totalDirectTime += avgDirectTime
		totalGas += avgGas
	}

	// Print summary
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Average:     | %-11s | %-11s | %d\n",
		fmt.Sprintf("%.2f", float64(totalTime.Nanoseconds())/(float64(len(epochs))*1000.0)),
		fmt.Sprintf("%.2f", float64(totalDirectTime.Nanoseconds())/(float64(len(epochs))*1000.0)),
		totalGas/uint64(len(epochs)))
}

// getEpochSnapshotSlotDirect is a direct implementation of getEpochSnapshotSlot without caching
func getEpochSnapshotSlotDirect(epoch *big.Int) (*big.Int, uint64) {
	// Initialize gas used counter
	gasUsed := uint64(0)

	// Calculate the slot for epoch snapshots
	// This is identical to getEpochSnapshotSlot but without caching
	slotConstant := int64(2)

	// Left-pad epoch to 32 bytes
	epochBytes := common.LeftPadBytes(epoch.Bytes(), 32)

	// Left-pad slot constant to 32 bytes
	slotBytes := common.LeftPadBytes(big.NewInt(slotConstant).Bytes(), 32)

	// Combine the bytes for the initial hash input
	initialHashInput := make([]byte, len(epochBytes)+len(slotBytes))
	copy(initialHashInput, epochBytes)
	copy(initialHashInput[len(epochBytes):], slotBytes)

	// Calculate the initial hash
	initialHash := crypto.Keccak256(initialHashInput)
	gasUsed += 30 // Approximate gas cost for keccak256

	// Create input for the final hash: epoch + mappingSlot
	mappingSlot := new(big.Int).SetBytes(initialHash)

	// Create a new byte slice for the final hash input
	mappingSlotBytes := common.LeftPadBytes(mappingSlot.Bytes(), 32)
	finalHashInput := make([]byte, len(epochBytes)+len(mappingSlotBytes))
	copy(finalHashInput, epochBytes)
	copy(finalHashInput[len(epochBytes):], mappingSlotBytes)

	// Calculate the final hash
	finalHash := crypto.Keccak256(finalHashInput)
	gasUsed += 30 // Approximate gas cost for keccak256

	return new(big.Int).SetBytes(finalHash), gasUsed
}
