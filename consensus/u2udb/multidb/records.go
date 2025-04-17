package multidb

import (
	"github.com/unicornultrafoundation/go-u2u/rlp"

	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
)

type TableRecord struct {
	Req   string
	Table string
}

func ReadTablesList(store u2udb.Store, tableRecordsKey []byte) (res []TableRecord, err error) {
	b, err := store.Get(tableRecordsKey)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return []TableRecord{}, nil
	}
	err = rlp.DecodeBytes(b, &res)
	return
}

func WriteTablesList(store u2udb.Store, tableRecordsKey []byte, records []TableRecord) error {
	b, err := rlp.EncodeToBytes(records)
	if err != nil {
		return err
	}
	return store.Put(tableRecordsKey, b)
}
