package littleendian

import "encoding/binary"

// Uint64ToBytes converts uint64 to bytes.
func Uint64ToBytes(n uint64) []byte {
	var res [8]byte
	binary.LittleEndian.PutUint64(res[:], n)
	return res[:]
}

// BytesToUint64 converts uint64 from bytes.
func BytesToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

// Uint32ToBytes converts uint32 to bytes.
func Uint32ToBytes(n uint32) []byte {
	var res [4]byte
	binary.LittleEndian.PutUint32(res[:], n)
	return res[:]
}

// BytesToUint32 converts uint32 from bytes.
func BytesToUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}

// Uint16ToBytes converts uint16 to bytes.
func Uint16ToBytes(n uint16) []byte {
	var res [2]byte
	binary.LittleEndian.PutUint16(res[:], n)
	return res[:]
}

// BytesToUint16 converts uint16 from bytes.
func BytesToUint16(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b)
}
