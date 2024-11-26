package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenericNameOfTmpDB(t *testing.T) {
	require := require.New(t)

	for name, exp := range map[string]string{
		"":            "",
		"main":        "main",
		"main-single": "main-single",
		"helios-0":    "helios-tmp",
		"helios-0999": "helios-tmp",
		"gossip-50":   "gossip-tmp",
		"epoch-1":     "epoch-tmp",
		"xxx-1a":      "xxx-1a",
		"123":         "123",
	} {
		got := genericNameOfTmpDB(name)
		require.Equal(exp, got, name)
	}
}
