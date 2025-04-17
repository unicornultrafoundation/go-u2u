package vecfc

import (
	"github.com/unicornultrafoundation/go-u2u/consensus/hash"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
)

func (vi *Index) getBytes(table u2udb.Store, id hash.Event) []byte {
	key := id.Bytes()
	b, err := table.Get(key)
	if err != nil {
		vi.crit(err)
	}
	return b
}

func (vi *Index) setBytes(table u2udb.Store, id hash.Event, b []byte) {
	key := id.Bytes()
	err := table.Put(key, b)
	if err != nil {
		vi.crit(err)
	}
}

// GetLowestAfter reads the vector from DB
func (vi *Index) GetLowestAfter(id hash.Event) *LowestAfterSeq {
	if bVal, okGet := vi.cache.LowestAfterSeq.Get(id); okGet {
		return bVal.(*LowestAfterSeq)
	}

	b := LowestAfterSeq(vi.getBytes(vi.table.LowestAfterSeq, id))
	if b == nil {
		return nil
	}
	vi.cache.LowestAfterSeq.Add(id, &b, uint(len(b)))
	return &b
}

// GetHighestBefore reads the vector from DB
func (vi *Index) GetHighestBefore(id hash.Event) *HighestBeforeSeq {
	if bVal, okGet := vi.cache.HighestBeforeSeq.Get(id); okGet {
		return bVal.(*HighestBeforeSeq)
	}

	b := HighestBeforeSeq(vi.getBytes(vi.table.HighestBeforeSeq, id))
	if b == nil {
		return nil
	}
	vi.cache.HighestBeforeSeq.Add(id, &b, uint(len(b)))
	return &b
}

// SetLowestAfter stores the vector into DB
func (vi *Index) SetLowestAfter(id hash.Event, seq *LowestAfterSeq) {
	vi.setBytes(vi.table.LowestAfterSeq, id, *seq)

	vi.cache.LowestAfterSeq.Add(id, seq, uint(len(*seq)))
}

// SetHighestBefore stores the vectors into DB
func (vi *Index) SetHighestBefore(id hash.Event, seq *HighestBeforeSeq) {
	vi.setBytes(vi.table.HighestBeforeSeq, id, *seq)

	vi.cache.HighestBeforeSeq.Add(id, seq, uint(len(*seq)))
}
