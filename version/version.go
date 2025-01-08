package version

import (
	"fmt"

	"github.com/unicornultrafoundation/go-u2u/params"
)

func AsString() string {
	return ToString(uint16(params.VersionMajor), uint16(params.VersionMinor), uint16(params.VersionPatch))
}

func AsU64() uint64 {
	return ToU64(uint16(params.VersionMajor), uint16(params.VersionMinor), uint16(params.VersionPatch))
}

func ToU64(vMajor, vMinor, vPatch uint16) uint64 {
	return uint64(vMajor)*1e12 + uint64(vMinor)*1e6 + uint64(vPatch)
}

func ToString(major, minor, patch uint16) string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func U64ToString(v uint64) string {
	return ToString(uint16((v/1e12)%1e6), uint16((v/1e6)%1e6), uint16(v%1e6))
}
