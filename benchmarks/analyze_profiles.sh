#!/bin/bash

# Directory containing profiles
PROFILE_DIR="./profiles"

# Check if profile directory exists
if [ ! -d "$PROFILE_DIR" ]; then
    echo "Profile directory $PROFILE_DIR does not exist."
    echo "Please run profile_sealepoch.sh first."
    exit 1
fi

# Create analysis directory for reports
ANALYSIS_DIR="$PROFILE_DIR/analysis"
mkdir -p $ANALYSIS_DIR

# Function to check if a profile is valid
is_valid_profile() {
    local profile_file=$1
    
    # Check if file exists and has content
    if [ ! -f "$profile_file" ] || [ ! -s "$profile_file" ]; then
        echo "Profile $profile_file does not exist or is empty."
        return 1
    fi
    
    # Try to run pprof on the file to check if it's valid
    if go tool pprof -raw "$profile_file" &> /dev/null; then
        return 0
    else
        echo "Profile $profile_file appears to be invalid or corrupt."
        return 1
    fi
}

# Check if any profiles were actually created - this indicates the node was running
if [ ! -s "$PROFILE_DIR/cpu.pprof" ] && [ ! -s "$PROFILE_DIR/heap_after.pprof" ]; then
    echo "ERROR: No valid profiles found. The node may not have been running during profiling."
    echo "Please ensure the node is running with --pprof flag before running profile_sealepoch.sh."
    echo
    echo "To start a node with profiling enabled, use:"
    echo "  ./build/demo_u2u --pprof --pprof.addr=\"127.0.0.1\" --pprof.port=6060 [other flags]"
    echo
    echo "Then run profile_sealepoch.sh again."
    exit 1
fi

echo "Generating profile reports in $ANALYSIS_DIR..."

# Analyze CPU profile
CPU_PROFILE="$PROFILE_DIR/cpu.pprof"
if is_valid_profile $CPU_PROFILE; then
    echo "Analyzing CPU profile..."
    go tool pprof -text -nodecount=20 $CPU_PROFILE > $ANALYSIS_DIR/cpu_top20.txt 2>/dev/null
    if [ $? -eq 0 ]; then
        go tool pprof -svg $CPU_PROFILE > $ANALYSIS_DIR/cpu_graph.svg 2>/dev/null
        
        # Extract specific SFC contract functions
        echo "Extracting SFC contract function details..."
        go tool pprof -text -nodecount=100 $CPU_PROFILE 2>/dev/null | grep -E "handleSealEpoch|_sealEpoch_|getEpochSnapshotSlot" > $ANALYSIS_DIR/cpu_sfc_functions.txt
    else
        echo "WARNING: Could not analyze CPU profile. It may be empty or in an incorrect format."
        echo "Profile content type: $(file $CPU_PROFILE)"
    fi
else
    echo "WARNING: CPU profile appears to be invalid. Skipping CPU profile analysis."
fi

# Analyze heap profiles
HEAP_BEFORE="$PROFILE_DIR/heap_before.pprof"
HEAP_AFTER="$PROFILE_DIR/heap_after.pprof"
if is_valid_profile $HEAP_BEFORE && is_valid_profile $HEAP_AFTER; then
    echo "Analyzing heap profiles..."
    go tool pprof -text -nodecount=20 $HEAP_AFTER > $ANALYSIS_DIR/heap_top20.txt 2>/dev/null
    if [ $? -eq 0 ]; then
        go tool pprof -svg $HEAP_AFTER > $ANALYSIS_DIR/heap_graph.svg 2>/dev/null
        
        # Analyze heap growth
        echo "Analyzing heap growth..."
        go tool pprof -text -nodecount=20 -base=$HEAP_BEFORE $HEAP_AFTER > $ANALYSIS_DIR/heap_growth_top20.txt 2>/dev/null
        go tool pprof -svg -base=$HEAP_BEFORE $HEAP_AFTER > $ANALYSIS_DIR/heap_growth_graph.svg 2>/dev/null
        
        # Extract specific SFC contract functions
        echo "Extracting SFC contract heap allocation details..."
        go tool pprof -text -nodecount=100 $HEAP_AFTER 2>/dev/null | grep -E "handleSealEpoch|_sealEpoch_|getEpochSnapshotSlot" > $ANALYSIS_DIR/heap_sfc_functions.txt
        go tool pprof -text -nodecount=100 -base=$HEAP_BEFORE $HEAP_AFTER 2>/dev/null | grep -E "handleSealEpoch|_sealEpoch_|getEpochSnapshotSlot" > $ANALYSIS_DIR/heap_growth_sfc_functions.txt
    else
        echo "WARNING: Could not analyze heap profiles. They may be empty or in an incorrect format."
    fi
else
    echo "WARNING: Heap profiles appear to be invalid. Skipping heap profile analysis."
fi

# Analyze allocs profile (cumulative allocations)
ALLOCS_PROFILE="$PROFILE_DIR/allocs.pprof"
if is_valid_profile $ALLOCS_PROFILE; then
    echo "Analyzing memory allocations profile..."
    go tool pprof -text -nodecount=20 $ALLOCS_PROFILE > $ANALYSIS_DIR/allocs_top20.txt 2>/dev/null
    if [ $? -eq 0 ]; then
        go tool pprof -svg $ALLOCS_PROFILE > $ANALYSIS_DIR/allocs_graph.svg 2>/dev/null
        
        # Extract specific SFC contract functions
        echo "Extracting SFC contract allocation details..."
        go tool pprof -text -nodecount=100 $ALLOCS_PROFILE 2>/dev/null | grep -E "handleSealEpoch|_sealEpoch_|getEpochSnapshotSlot" > $ANALYSIS_DIR/allocs_sfc_functions.txt
    else
        echo "WARNING: Could not analyze allocs profile. It may be empty or in an incorrect format."
    fi
else
    echo "WARNING: Allocs profile appears to be invalid. Skipping allocs profile analysis."
fi

# Check if any reports were generated
if [ -f "$ANALYSIS_DIR/cpu_top20.txt" ] || [ -f "$ANALYSIS_DIR/heap_top20.txt" ] || [ -f "$ANALYSIS_DIR/allocs_top20.txt" ]; then
    echo "Analysis complete. Reports are available in $ANALYSIS_DIR/"
    echo 
    echo "To start an interactive web UI for further analysis, run:"
    echo "  go tool pprof -http=:8080 $PROFILE_DIR/cpu.pprof"
    echo "  go tool pprof -http=:8080 $PROFILE_DIR/allocs.pprof"
    echo "  go tool pprof -http=:8080 -base=$PROFILE_DIR/heap_before.pprof $PROFILE_DIR/heap_after.pprof"
else
    echo "ERROR: No reports were generated because all profiles were invalid."
    echo "This typically happens when:"
    echo "  1. The node's pprof endpoint wasn't running (localhost:6060)"
    echo "  2. The SealEpoch operation wasn't actually executed"
    echo "  3. The profiling duration was too short"
    echo
    echo "Please check the profile_sealepoch.sh script and ensure a node is running with profiling enabled."
fi 