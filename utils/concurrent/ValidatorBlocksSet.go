package concurrent

import (
	"sync"

	"github.com/unicornultrafoundation/go-hashgraph/inter/idx"
)

type ValidatorBlocksSet struct {
	sync.RWMutex
	Val map[idx.ValidatorID]idx.Block
}

func WrapValidatorBlocksSet(v map[idx.ValidatorID]idx.Block) *ValidatorBlocksSet {
	return &ValidatorBlocksSet{
		RWMutex: sync.RWMutex{},
		Val:     v,
	}
}
