package migration

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb"
)

// U2UdbIDStore stores id
type U2UdbIDStore struct {
	table u2udb.Store
	key   []byte
}

// NewU2UdbIDStore constructor
func NewU2UdbIDStore(table u2udb.Store) *U2UdbIDStore {
	return &U2UdbIDStore{
		table: table,
		key:   []byte("id"),
	}
}

// GetID is a getter
func (p *U2UdbIDStore) GetID() string {
	id, err := p.table.Get(p.key)
	if err != nil {
		log.Crit("Failed to get key-value", "err", err)
	}

	if id == nil {
		return ""
	}
	return string(id)
}

// SetID is a setter
func (p *U2UdbIDStore) SetID(id string) {
	err := p.table.Put(p.key, []byte(id))
	if err != nil {
		log.Crit("Failed to put key-value", "err", err)
	}
}
