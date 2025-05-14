package vm

import "github.com/unicornultrafoundation/go-u2u/metrics"

var (
	// execution time diff metrics per tx
	sfcDiffCallMeter         = metrics.GetOrRegisterMeter("sfc/diff/call", nil)
	sfcDiffCallCodeMeter     = metrics.GetOrRegisterMeter("sfc/diff/callcode", nil)
	sfcDiffDelegateCallMeter = metrics.GetOrRegisterMeter("sfc/diff/delegatecall", nil)
	sfcDiffStaticCallMeter   = metrics.GetOrRegisterMeter("sfc/diff/staticcall", nil)

	// count metrics
	sfcCallGauge         = metrics.GetOrRegisterGauge("sfc/call", nil)
	sfcCallCodeGauge     = metrics.GetOrRegisterGauge("sfc/callcode", nil)
	sfcDelegateCallGauge = metrics.GetOrRegisterGauge("sfc/delegatecall", nil)
	sfcStaticCallGauge   = metrics.GetOrRegisterGauge("sfc/staticcall", nil)
)
