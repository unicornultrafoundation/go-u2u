package evmcore

import (
	"github.com/unicornultrafoundation/go-u2u/metrics"
)

var (
	validPaymasterCounter   = metrics.NewRegisteredCounter("evmcore/paymaster/valid", nil)
	invalidPaymasterCounter = metrics.NewRegisteredCounter("evmcore/paymaster/invalid", nil)
)
