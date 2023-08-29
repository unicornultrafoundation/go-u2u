package topicsdb

import (
	"context"
	"github.com/unicornultrafoundation/go-u2u/utils/dbutil/threads"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/unicornultrafoundation/go-hashgraph/native/idx"
)

// withThreadPool wraps the index and limits its threads in use
type withThreadPool struct {
	*index
}

// FindInBlocks returns all log records of block range by pattern. 1st pattern element is an address.
func (tt *withThreadPool) FindInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash) (logs []*types.Log, err error) {
	err = tt.ForEachInBlocks(
		ctx,
		from, to,
		pattern,
		func(l *types.Log) bool {
			logs = append(logs, l)
			return true
		})

	return
}

// ForEachInBlocks matches log records of block range by pattern. 1st pattern element is an address.
func (tt *withThreadPool) ForEachInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if 0 < to && to < from {
		return nil
	}

	pattern, err := limitPattern(pattern)
	if err != nil {
		return err
	}

	onMatched := func(rec *logrec) (gonext bool, err error) {
		rec.fetch(tt.table.Logrec)
		if rec.err != nil {
			err = rec.err
			return
		}
		gonext = onLog(rec.result)
		return
	}

	splitby := 0
	threads := 0
	for i := range pattern {
		threads += len(pattern[i])
		if len(pattern[splitby]) < len(pattern[i]) {
			splitby = i
		}
	}
	rest := pattern[splitby]
	threads -= len(rest)

	for len(rest) > 0 {
		got, release := globalPool.Lock(threads + len(rest))
		if got <= threads {
			release()
			select {
			case <-time.After(time.Millisecond):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		pattern[splitby] = rest[:got-threads]
		rest = rest[got-threads:]
		err = tt.searchParallel(ctx, pattern, uint64(from), uint64(to), onMatched)
		release()
		if err != nil {
			return err
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
