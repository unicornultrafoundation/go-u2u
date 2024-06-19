package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/metrics"
)

var logger = log.New("module", "prometheus")

// PrometheusListener serves prometheus connections.
func PrometheusListener(endpoint string, reg metrics.Registry) {
	if reg == nil {
		reg = metrics.DefaultRegistry
	}
	reg.Each(collect)

	go func() {
		logger.Info("Metrics server starts", "endpoint", endpoint)
		defer logger.Info("Metrics server is stopped")

		http.HandleFunc(
			"/metrics", promhttp.Handler().ServeHTTP)
		err := http.ListenAndServe(endpoint, nil)
		if err != nil {
			logger.Info("metrics server", "err", err)
		}

		// TODO: wait for exit signal?
	}()
}

func collect(name string, metric interface{}) {
	collector, ok := convertToPrometheusMetric(name, metric)
	if !ok {
		return
	}

	err := prometheus.Register(collector)
	if err != nil {
		switch err.(type) {
		case prometheus.AlreadyRegisteredError:
			return
		default:
			logger.Warn(err.Error())
		}
	}
}
