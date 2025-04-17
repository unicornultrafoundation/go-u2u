package threads

import (
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"

	"github.com/unicornultrafoundation/go-u2u/logger"
)

type (
	countedDbProducer struct {
		u2udb.DBProducer
	}

	countedFullDbProducer struct {
		u2udb.FullDBProducer
	}

	countedStore struct {
		u2udb.Store
	}

	countedIterator struct {
		u2udb.Iterator
		release func(count int)
	}
)

func CountedDBProducer(dbs u2udb.DBProducer) u2udb.DBProducer {
	return &countedDbProducer{dbs}
}

func CountedFullDBProducer(dbs u2udb.FullDBProducer) u2udb.FullDBProducer {
	return &countedFullDbProducer{dbs}
}

func (p *countedDbProducer) OpenDB(name string) (u2udb.Store, error) {
	s, err := p.DBProducer.OpenDB(name)
	return &countedStore{s}, err
}

func (p *countedFullDbProducer) OpenDB(name string) (u2udb.Store, error) {
	s, err := p.FullDBProducer.OpenDB(name)
	return &countedStore{s}, err
}

var notifier = logger.New("threads-pool")

func (s *countedStore) NewIterator(prefix []byte, start []byte) u2udb.Iterator {
	got, release := GlobalPool.Lock(1)
	if got < 1 {
		notifier.Log.Warn("Too many DB iterators")
	}

	return &countedIterator{
		Iterator: s.Store.NewIterator(prefix, start),
		release:  release,
	}
}

func (it *countedIterator) Release() {
	it.Iterator.Release()
	it.release(1)
}
