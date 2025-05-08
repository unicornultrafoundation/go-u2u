#!/bin/bash

# Profiling SFC contract functions directly by instrumenting the code

# Set up profiles directory
mkdir -p ./profiles

# Create or update the profiler Go file
cat > ./benchmarks/sfc_profiler.go << EOF
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
EOF

echo "Compiling and running SFC profiler..."
cd "$(dirname "$0")/.." && go run ./benchmarks/sfc_profiler.go

if [ -f "./profiles/sfc_cpu.pprof" ]; then
  echo "Analyzing CPU profile..."
  mkdir -p ./profiles/analysis
  go tool pprof -text -nodecount=20 ./profiles/sfc_cpu.pprof > ./profiles/analysis/cpu_top20.txt
  go tool pprof -text -nodecount=1000 ./profiles/sfc_cpu.pprof | grep -i "sfc\|getEpochSnapshotSlot" > ./profiles/analysis/cpu_sfc_functions.txt
  go tool pprof -svg ./profiles/sfc_cpu.pprof > ./profiles/analysis/cpu_graph.svg
else
  echo "Failed to fetch CPU profile: ./profiles/sfc_cpu.pprof does not exist"
fi

if [ -f "./profiles/sfc_mem.pprof" ]; then
  echo "Analyzing memory profile..."
  mkdir -p ./profiles/analysis
  go tool pprof -text -nodecount=20 ./profiles/sfc_mem.pprof > ./profiles/analysis/heap_top20.txt
  go tool pprof -text -nodecount=1000 ./profiles/sfc_mem.pprof | grep -i "sfc\|getEpochSnapshotSlot" > ./profiles/analysis/heap_sfc_functions.txt
  go tool pprof -svg ./profiles/sfc_mem.pprof > ./profiles/analysis/heap_graph.svg
else
  echo "Failed to fetch memory profile: ./profiles/sfc_mem.pprof does not exist"
fi

echo "Results available in ./profiles/ directory"
echo "To view interactive results, run:"
echo "  go tool pprof ./profiles/sfc_cpu.pprof"
echo "  go tool pprof ./profiles/sfc_mem.pprof" 