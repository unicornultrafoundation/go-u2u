package bvprocessor

import (
	"time"

	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/unicornultrafoundation/go-hashgraph/native/dag"
	"github.com/unicornultrafoundation/go-hashgraph/utils/cachescale"
)

type Config struct {
	BufferLimit dag.Metric

	SemaphoreTimeout time.Duration

	MaxTasks int
}

func DefaultConfig(scale cachescale.Func) Config {
	return Config{
		BufferLimit: dag.Metric{
			Num:  3000,
			Size: scale.U64(15 * opt.MiB),
		},
		SemaphoreTimeout: 10 * time.Second,
		MaxTasks:         512,
	}
}
