package prometheus

import (
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/metrics/prometheus"
)

var PrometheusEndpointFlag = cli.StringFlag{
	Name:  "monitoring.prometheus.endpoint",
	Usage: "Prometheus API endpoint to report metrics to",
	Value: ":19090",
}

func SetupPrometheus(ctx *cli.Context) {
	if !metrics.Enabled {
		return
	}
	prometheus.SetNamespace("u2u")
	var endpoint = ctx.GlobalString(PrometheusEndpointFlag.Name)
	prometheus.ListenTo(endpoint, nil)
}

var (
	// TODO: refactor it
	dbDirMonitor        atomic.Value
	dbSizeMetricMonitor = metrics.NewRegisteredFunctionalGauge("db_size", nil, measureDbDirMonitor)
)

func SetDataDirMonitor(datadir string) {
	dbDirMonitor.Store(datadir)
}

func measureDbDirMonitor() (size int64) {
	datadir, ok := dbDirMonitor.Load().(string)
	if !ok || datadir == "" || datadir == "inmemory" {
		return
	}

	err := filepath.Walk(datadir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		log.Error("filepath.Walk", "path", datadir, "err", err)
		return 0
	}

	return
}
