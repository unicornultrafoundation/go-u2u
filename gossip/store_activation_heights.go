package gossip

import (
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

func (s *Store) AddUpgradeHeight(h u2u.UpgradeHeight) {
	orig := s.GetUpgradeHeights()
	// allocate new memory to avoid race condition in cache
	cp := make([]u2u.UpgradeHeight, 0, len(orig)+1)
	cp = append(append(cp, orig...), h)

	s.rlp.Set(s.table.UpgradeHeights, []byte{}, cp)
	s.cache.UpgradeHeights.Store(cp)
}

func (s *Store) GetUpgradeHeights() []u2u.UpgradeHeight {
	if v := s.cache.UpgradeHeights.Load(); v != nil {
		return v.([]u2u.UpgradeHeight)
	}
	hh, ok := s.rlp.Get(s.table.UpgradeHeights, []byte{}, &[]u2u.UpgradeHeight{}).(*[]u2u.UpgradeHeight)
	if !ok {
		return []u2u.UpgradeHeight{}
	}
	s.cache.UpgradeHeights.Store(*hh)
	return *hh
}
