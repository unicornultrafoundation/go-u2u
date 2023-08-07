package utils

import (
	"github.com/unicornultrafoundation/go-hashgraph/u2udb"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb/table"
)

func NewTableOrSelf(db u2udb.Store, prefix []byte) u2udb.Store {
	if len(prefix) == 0 {
		return db
	}
	return table.New(db, prefix)
}
