package concurrent

import (
	"sync"

	"github.com/unicornultrafoundation/go-u2u/consensus/hash"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/idx"
)

type ValidatorEventsSet struct {
	sync.RWMutex
	Val map[idx.ValidatorID]hash.Event
}

func WrapValidatorEventsSet(v map[idx.ValidatorID]hash.Event) *ValidatorEventsSet {
	return &ValidatorEventsSet{
		RWMutex: sync.RWMutex{},
		Val:     v,
	}
}
