package skipkeys

import "github.com/unicornultrafoundation/go-u2u/helios/u2udb"

func openDB(p u2udb.DBProducer, skipPrefix []byte, name string) (u2udb.Store, error) {
	store, err := p.OpenDB(name)
	if err != nil {
		return nil, err
	}
	return &Store{store, skipPrefix}, nil
}

type AllDBProducer struct {
	u2udb.FullDBProducer
	skipPrefix []byte
}

func WrapAllProducer(p u2udb.FullDBProducer, skipPrefix []byte) *AllDBProducer {
	return &AllDBProducer{
		FullDBProducer: p,
		skipPrefix:     skipPrefix,
	}
}

func (p *AllDBProducer) OpenDB(name string) (u2udb.Store, error) {
	return openDB(p.FullDBProducer, p.skipPrefix, name)
}

type DBProducer struct {
	u2udb.DBProducer
	skipPrefix []byte
}

func WrapProducer(p u2udb.DBProducer, skipPrefix []byte) *DBProducer {
	return &DBProducer{
		DBProducer: p,
		skipPrefix: skipPrefix,
	}
}

func (p *DBProducer) OpenDB(name string) (u2udb.Store, error) {
	return openDB(p.DBProducer, p.skipPrefix, name)
}
