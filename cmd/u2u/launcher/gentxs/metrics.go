package gentxs

import (
	"fmt"

	"github.com/ethereum/go-ethereum/metrics"
	
	"github.com/unicornultrafoundation/go-u2u/monitoring/prometheus"
	
	cli "gopkg.in/urfave/cli.v1"
)

var PrometheusMonitoringPortFlag = cli.IntFlag{
	Name:  "monitor.prometheus.port",
	Usage: "Opens Prometheus API port to mornitor metrics",
	Value: 19090,
}

var (
	reg = metrics.NewRegistry()

	txCountSentMeter = metrics.NewRegisteredCounter("tx_count_sent", reg)
	txCountGotMeter  = metrics.NewRegisteredCounter("tx_count_got", reg)
	txTpsMeter       = metrics.NewRegisteredHistogram("tx_tps", reg, metrics.NewUniformSample(500))
)

func SetupPrometheus(ctx *cli.Context) {
	prometheus.SetNamespace("gentxs")
	endpoint := ctx.GlobalInt(PrometheusMonitoringPortFlag.Name)
	prometheus.PrometheusListener(fmt.Sprintf(":%d",endpoint), reg)
}
