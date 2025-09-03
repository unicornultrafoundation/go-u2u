package driver

// Storage slots for NodeDriver contract variables
const (
	isInitialized int64 = 0x0
	offset        int64 = 0x33 // Base offset for storage slots of NodeDriver contract when implement Initializable contract
	// uint256 private erased0 - slot 0
	backendSlot   = offset + 1
	evmWriterSlot = offset + 2
)
