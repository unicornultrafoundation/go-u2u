package cachedproducer

import (
	"errors"
	"sync"

	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
)

type cacheState struct {
	opened     map[string]u2udb.Store
	refCounter map[string]int
	notDropped map[string]bool
	mu         sync.Mutex
}

func openDB(p u2udb.DBProducer, c *cacheState, name string) (u2udb.Store, error) {
	// Validate name parameter
	if !u2udb.IsValidDatabaseName(name) {
		return nil, errors.New("invalid database name")
	}
	{ // protected by mutex
		c.mu.Lock()
		c.notDropped[name] = true
		if store, ok := c.opened[name]; ok {
			c.refCounter[name]++
			c.mu.Unlock()
			return store, nil
		}
		c.mu.Unlock()
	}
	store, err := p.OpenDB(name)
	if err != nil {
		return nil, err
	}
	realClose := store.Close
	realDrop := store.Drop
	store = &StoreWithFn{
		Store: store,
		CloseFn: func() error {
			// call realClose only after every instance was closed
			toClose := false
			{ // protected by mutex
				c.mu.Lock()
				counter := c.refCounter[name]
				if counter <= 0 {
					c.mu.Unlock()
					return errors.New("called Close more times than OpenDB")
				} else if counter == 1 {
					delete(c.refCounter, name)
					delete(c.opened, name)
					toClose = true
				} else {
					counter--
					c.refCounter[name] = counter
				}
				c.mu.Unlock()
			}
			if toClose {
				return realClose()
			}
			return nil
		},
		DropFn: func() {
			// don't allow to call realDrop twice
			toDrop := false
			{ // protected by mutex
				c.mu.Lock()
				toDrop = c.notDropped[name]
				delete(c.notDropped, name)
				c.mu.Unlock()
			}
			if toDrop {
				realDrop()
			}
		},
	}

	{ // protected by mutex
		c.mu.Lock()
		c.opened[name] = store
		c.refCounter[name]++
		c.mu.Unlock()
	}
	return store, nil
}

type AllDBProducer struct {
	u2udb.FullDBProducer
	cacheState
}

func WrapAll(p u2udb.FullDBProducer) *AllDBProducer {
	return &AllDBProducer{
		FullDBProducer: p,
		cacheState: cacheState{
			opened:     make(map[string]u2udb.Store),
			refCounter: make(map[string]int),
			notDropped: make(map[string]bool),
		},
	}
}

func (p *AllDBProducer) OpenDB(name string) (u2udb.Store, error) {
	return openDB(p.FullDBProducer, &p.cacheState, name)
}

type DBProducer struct {
	u2udb.DBProducer
	cacheState
}

func Wrap(p u2udb.DBProducer) *DBProducer {
	return &DBProducer{
		DBProducer: p,
		cacheState: cacheState{
			opened:     make(map[string]u2udb.Store),
			notDropped: make(map[string]bool),
		},
	}
}

func (p *DBProducer) OpenDB(name string) (u2udb.Store, error) {
	return openDB(p.DBProducer, &p.cacheState, name)
}
