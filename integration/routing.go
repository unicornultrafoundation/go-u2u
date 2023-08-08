package integration

import (
	"fmt"

	"github.com/unicornultrafoundation/go-hashgraph/u2udb"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb/cachedproducer"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb/multidb"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb/skipkeys"
)

type RoutingConfig struct {
	Table map[string]multidb.Route
}

func (a RoutingConfig) Equal(b RoutingConfig) bool {
	if len(a.Table) != len(b.Table) {
		return false
	}
	for k, v := range a.Table {
		if b.Table[k] != v {
			return false
		}
	}
	return true
}

func MakeMultiProducer(rawProducers map[multidb.TypeName]u2udb.IterableDBProducer, scopedProducers map[multidb.TypeName]u2udb.FullDBProducer, cfg RoutingConfig) (u2udb.FullDBProducer, error) {
	cachedProducers := make(map[multidb.TypeName]u2udb.FullDBProducer)
	var flushID []byte
	var err error
	for typ, producer := range scopedProducers {
		flushID, err = producer.Initialize(rawProducers[typ].Names(), flushID)
		if err != nil {
			return nil, fmt.Errorf("failed to open existing databases: %v. Try to use 'db heal' to recover", err)
		}
		cachedProducers[typ] = cachedproducer.WrapAll(producer)
	}

	p, err := makeMultiProducer(cachedProducers, cfg)
	return p, err
}

func MakeDirectMultiProducer(rawProducers map[multidb.TypeName]u2udb.IterableDBProducer, cfg RoutingConfig) (u2udb.FullDBProducer, error) {
	dproducers := map[multidb.TypeName]u2udb.FullDBProducer{}
	for typ, producer := range rawProducers {
		dproducers[typ] = &DummyScopedProducer{producer}
	}
	return MakeMultiProducer(rawProducers, dproducers, cfg)
}

func makeMultiProducer(scopedProducers map[multidb.TypeName]u2udb.FullDBProducer, cfg RoutingConfig) (u2udb.FullDBProducer, error) {
	multi, err := multidb.NewProducer(scopedProducers, cfg.Table, TablesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to construct multidb: %v", err)
	}

	err = multi.Verify()
	if err != nil {
		return nil, fmt.Errorf("incompatible chainstore DB layout: %v. Try to use 'db transform' to recover", err)
	}
	return skipkeys.WrapAllProducer(multi, MetadataPrefix), nil
}
