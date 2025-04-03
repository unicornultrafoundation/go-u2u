package vm

import "github.com/unicornultrafoundation/go-u2u/metrics"

var (
	// execution time diff metrics
	sfcDiffCallMeter         = metrics.GetOrRegisterMeter("sfc/diff/call", nil)
	sfcDiffCallCodeMeter     = metrics.GetOrRegisterMeter("sfc/diff/callcode", nil)
	sfcDiffDelegateCallMeter = metrics.GetOrRegisterMeter("sfc/diff/delegatecall", nil)
	sfcDiffStaticCallMeter   = metrics.GetOrRegisterMeter("sfc/diff/staticcall", nil)
)
