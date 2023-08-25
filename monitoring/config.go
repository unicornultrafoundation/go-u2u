package monitoring

import (
	"github.com/ethereum/go-ethereum/metrics"
)

// DefaultConfig is the default config for monitorings used in U2U.
type Config struct {
	Port             int `toml:",omitempty"`
	TxCountSentMeter metrics.Counter
	TxCountGotMeter  metrics.Counter
	TxTpsMeter       metrics.Histogram
}

// DefaultConfig is the default config for monitorings used in U2U.
var DefaultConfig = Config{
	Port:             19090,
	TxCountSentMeter: metrics.NewRegisteredCounter("tx_count_sent", nil),
	TxCountGotMeter:  metrics.NewRegisteredCounter("tx_count_got", nil),
	TxTpsMeter:       metrics.NewRegisteredHistogram("tx_tps", nil, metrics.NewUniformSample(500)),
}
