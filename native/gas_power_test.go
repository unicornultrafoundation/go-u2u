package native

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/go-u2u/libs/common"
	"github.com/unicornultrafoundation/go-u2u/libs/rlp"
)

func TestRlp(t *testing.T) {
	require := require.New(t)
	v := GasPowerLeft{
		Gas: [2]uint64{0xBAADCAFE, 0xDEADBEEF},
	}
	b, err := rlp.EncodeToBytes(v)
	require.NoError(err)
	require.Equal("cbca84baadcafe84deadbeef", common.Bytes2Hex(b))
}
