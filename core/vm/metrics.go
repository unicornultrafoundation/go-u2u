package vm

import "github.com/unicornultrafoundation/go-u2u/metrics"

var (
	// execution time diff metrics per tx
	sfcDiffCallMeter = metrics.GetOrRegisterMeter("sfc/diff/call", nil)
)
