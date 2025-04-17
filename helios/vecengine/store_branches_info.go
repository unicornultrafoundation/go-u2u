package vecengine

import (
	"errors"

	"github.com/unicornultrafoundation/go-u2u/rlp"

	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
)

func (vi *Engine) setRlp(table u2udb.Store, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		vi.crit(err)
	}

	if err := table.Put(key, buf); err != nil {
		vi.crit(err)
	}
}

func (vi *Engine) getRlp(table u2udb.Store, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		vi.crit(err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		vi.crit(err)
	}
	return to
}

func (vi *Engine) getBytes(table u2udb.Store, id hash.Event) []byte {
	key := id.Bytes()
	b, err := table.Get(key)
	if err != nil {
		vi.crit(err)
	}
	return b
}

func (vi *Engine) setBytes(table u2udb.Store, id hash.Event, b []byte) {
	key := id.Bytes()
	err := table.Put(key, b)
	if err != nil {
		vi.crit(err)
	}
}

func (vi *Engine) setBranchesInfo(info *BranchesInfo) {
	key := []byte("c")

	vi.setRlp(vi.table.BranchesInfo, key, info)
}

func (vi *Engine) getBranchesInfo() *BranchesInfo {
	key := []byte("c")

	w, exists := vi.getRlp(vi.table.BranchesInfo, key, &BranchesInfo{}).(*BranchesInfo)
	if !exists {
		return nil
	}

	return w
}

// SetEventBranchID stores the event's global branch ID
func (vi *Engine) SetEventBranchID(id hash.Event, branchID idx.Validator) {
	vi.setBytes(vi.table.EventBranch, id, branchID.Bytes())
}

// GetEventBranchID reads the event's global branch ID
func (vi *Engine) GetEventBranchID(id hash.Event) idx.Validator {
	b := vi.getBytes(vi.table.EventBranch, id)
	if b == nil {
		vi.crit(errors.New("failed to read event's branch ID (inconsistent DB)"))
		return 0
	}
	branchID := idx.BytesToValidator(b)
	return branchID
}
