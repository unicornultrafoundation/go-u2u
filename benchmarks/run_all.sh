#!/bin/bash

# Make all scripts executable
chmod +x benchmarks/*.sh

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}U2U SFC Contract Memory Profiling Workflow${NC}"
echo -e "${GREEN}=========================================${NC}"

# Step 1: Start the test node
echo -e "\n${YELLOW}Step 1: Starting test node with profiling enabled...${NC}"
./benchmarks/start_test_node.sh &
NODE_PID=$!

# Wait for the node to start up
echo "Waiting 10 seconds for the node to initialize..."
sleep 10

# Step 2: Run the profiling
echo -e "\n${YELLOW}Step 2: Running profiling of SealEpoch function...${NC}"
./benchmarks/profile_sealepoch.sh
PROFILE_RESULT=$?

if [ $PROFILE_RESULT -ne 0 ]; then
    echo -e "\n${RED}Profiling failed! Please check the error messages above.${NC}"
    
    # Stop the node even if profiling failed
    echo -e "\n${YELLOW}Stopping test node...${NC}"
    ./benchmarks/stop_test_node.sh
    
    exit 1
fi

# Step 3: Analyze the profiles
echo -e "\n${YELLOW}Step 3: Analyzing collected profiles...${NC}"
./benchmarks/analyze_profiles.sh

# Step 4: Stop the test node
echo -e "\n${YELLOW}Step 4: Stopping test node...${NC}"
./benchmarks/stop_test_node.sh

echo -e "\n${GREEN}=========================================${NC}"
echo -e "${GREEN}Profiling workflow completed!${NC}"
echo -e "${GREEN}Check the profiles/analysis directory for results.${NC}"
echo -e "${GREEN}=========================================${NC}" 