package discfilter

import (
	lru "github.com/hashicorp/golang-lru"

	"github.com/unicornultrafoundation/go-u2u/libs/p2p/enode"
	"github.com/unicornultrafoundation/go-u2u/libs/p2p/enr"
)

var (
	enabled    = false
	dynamic, _ = lru.New(50000)
)

func Enable() {
	enabled = true
}

func Ban(id enode.ID) {
	if enabled {
		dynamic.Add(id, struct{}{})
	}
}

func BannedDynamic(id enode.ID) bool {
	if !enabled {
		return false
	}
	return dynamic.Contains(id)
}

func BannedStatic(rec *enr.Record) bool {
	if !enabled {
		return false
	}
	return rec.Has("eth") || rec.Has("eth2")
}

func Banned(id enode.ID, rec *enr.Record) bool {
	if !enabled {
		return false
	}
	return BannedStatic(rec) || BannedDynamic(id)
}
