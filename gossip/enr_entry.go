package gossip

import (
	"github.com/unicornultrafoundation/go-u2u/libs/core/forkid"
	"github.com/unicornultrafoundation/go-u2u/rlp"
)

// Enr is ENR entry which advertises eth protocol
// on the discovery network.
type Enr struct {
	ForkID forkid.ID
	// Ignore additional fields (for forward compatibility).
	Rest []rlp.RawValue `rlp:"tail"`
}

// ENRKey implements enr.Entry.
func (e Enr) ENRKey() string {
	return "u2u"
}

func (s *Service) currentEnr() *Enr {
	return &Enr{}
}
