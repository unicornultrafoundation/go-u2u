package evmcore

import (
	"github.com/unicornultrafoundation/go-u2u/metrics"
)

// Paymaster metrics
var (
	invalidAAParamsTxCounter          = metrics.NewRegisteredCounter("evmcore/aa/params/tx/invalid", nil)
	invalidPaymasterParamsTxCounter   = metrics.NewRegisteredCounter("evmcore/paymaster/params/tx/invalid", nil)
	invalidPaymasterParamsMsgCounter  = metrics.NewRegisteredCounter("evmcore/paymaster/params/msg/invalid", nil)
	validPaymasterCounter             = metrics.NewRegisteredCounter("evmcore/paymaster/valid", nil)
	invalidPaymasterCounter           = metrics.NewRegisteredCounter("evmcore/paymaster/invalid", nil)
	paymasterDepletedCounter          = metrics.NewRegisteredCounter("evmcore/paymaster/depleted", nil)
	invalidPrepareForPaymasterCounter = metrics.NewRegisteredCounter("evmcore/aa/prepare/invalid", nil)
)