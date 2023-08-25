package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenericNameOfTmpDB(t *testing.T) {
	require := require.New(t)

	for name, exp := range map[string]string{
		"":               "",
		"main":           "main",
		"main-single":    "main-single",
		"hashgraph-0":    "hashgraph-tmp",
		"hashgraph-0999": "hashgraph-tmp",
		"gossip-50":      "gossip-tmp",
		"epoch-1":        "epoch-tmp",
		"xxx-1a":         "xxx-1a",
		"123":            "123",
	} {
		got := genericNameOfTmpDB(name)
		require.Equal(exp, got, name)
	}
}
