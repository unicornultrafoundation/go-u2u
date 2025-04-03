package vm

import "github.com/unicornultrafoundation/go-u2u/metrics"

var (
	// execution time diff metrics
	sfcDiffCallTimer         = metrics.NewRegisteredResettingTimer("sfc/diff/call", nil)
	sfcDiffCallCodeTimer     = metrics.NewRegisteredResettingTimer("sfc/diff/callcode", nil)
	sfcDiffDelegateCallTimer = metrics.NewRegisteredResettingTimer("sfc/diff/delegatecall", nil)
	sfcDiffStaticCallTimer   = metrics.NewRegisteredResettingTimer("sfc/diff/staticcall", nil)
)
