// SFC contract profiler
package main

import (
	"fmt"
	"math/big"
	"os"
	"runtime/pprof"
	"time"

	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfc"
)

func main() {
	// Set up profiling
	profileDir := "./profiles"
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		if err := os.MkdirAll(profileDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Could not create profiles directory: %v\n", err)
			os.Exit(1)
		}
	}

	cpuFile, err := os.Create(profileDir + "/sfc_cpu.pprof")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create CPU profile: %v\n", err)
		os.Exit(1)
	}
	defer cpuFile.Close()

	memFile, err := os.Create(profileDir + "/sfc_mem.pprof")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create memory profile: %v\n", err)
		os.Exit(1)
	}
	defer memFile.Close()

	fmt.Println("Starting profiling of SFC functions...")
	
	// Start CPU profiling
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		fmt.Fprintf(os.Stderr, "Could not start CPU profile: %v\n", err)
		os.Exit(1)
	}
	defer pprof.StopCPUProfile()

	// Profile getEpochSnapshotSlot
	fmt.Println("Profiling getEpochSnapshotSlot...")
	
	iterations := 1000
	epochs := []*big.Int{
		big.NewInt(1),
		big.NewInt(100),
		big.NewInt(1000),
		big.NewInt(10000),
		big.NewInt(100000),
	}
	
	start := time.Now()
	for _, epoch := range epochs {
		for i := 0; i < iterations; i++ {
			// This will calculate without using cache
			sfc.GetCachedEpochSnapshotSlot(epoch)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("Completed %d iterations across %d epochs in %s\n", 
		iterations*len(epochs), len(epochs), elapsed)
	fmt.Printf("Average time per call: %s\n", 
		elapsed/time.Duration(iterations*len(epochs)))
	
	// Write memory profile
	if err := pprof.WriteHeapProfile(memFile); err != nil {
		fmt.Fprintf(os.Stderr, "Could not write memory profile: %v\n", err)
	}
	
	fmt.Println("Profiling complete. Results saved to:")
	fmt.Println("  - " + profileDir + "/sfc_cpu.pprof")
	fmt.Println("  - " + profileDir + "/sfc_mem.pprof")
}
