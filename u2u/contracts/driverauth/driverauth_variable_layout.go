package driverauth

// Storage slots for NodeDriverAuth contract variables
const (
	isInitialized int64 = 0x0
	ownerSlot     int64 = 0x33
	offset        int64 = 0x66 // Base offset for storage slots of NodeDriveAuth contract
	sfcSlot             = offset + 1
	driverSlot          = offset + 2
)
