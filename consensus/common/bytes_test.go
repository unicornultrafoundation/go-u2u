package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/go-u2u/consensus/common/bigendian"
	"github.com/unicornultrafoundation/go-u2u/consensus/common/littleendian"
)

func Test_IntToBytes(t *testing.T) {
	assertar := assert.New(t)

	for _, n1 := range []uint64{
		0,
		9,
		0xF000000000000000,
		0x000000000000000F,
		0xFFFFFFFFFFFFFFFF,
		47528346792,
	} {
		{
			b := bigendian.Uint64ToBytes(n1)
			assertar.Equal(8, len(b))
			n2 := bigendian.BytesToUint64(b)
			assertar.Equal(n1, n2)
		}
		{
			b := littleendian.Uint64ToBytes(n1)
			assertar.Equal(8, len(b))
			n2 := littleendian.BytesToUint64(b)
			assertar.Equal(n1, n2)
		}
	}
	for _, n1 := range []uint32{
		0,
		9,
		0xFFFFFFFF,
		475283467,
	} {
		{
			b := bigendian.Uint32ToBytes(n1)
			assertar.Equal(4, len(b))
			n2 := bigendian.BytesToUint32(b)
			assertar.Equal(n1, n2)
		}
		{
			b := littleendian.Uint32ToBytes(n1)
			assertar.Equal(4, len(b))
			n2 := littleendian.BytesToUint32(b)
			assertar.Equal(n1, n2)
		}
	}
	for _, n1 := range []uint16{
		0,
		9,
		0xFFFF,
		47528,
	} {
		{
			b := bigendian.Uint16ToBytes(n1)
			assertar.Equal(2, len(b))
			n2 := bigendian.BytesToUint16(b)
			assertar.Equal(n1, n2)
		}
		{
			b := littleendian.Uint16ToBytes(n1)
			assertar.Equal(2, len(b))
			n2 := littleendian.BytesToUint16(b)
			assertar.Equal(n1, n2)
		}
	}
}
