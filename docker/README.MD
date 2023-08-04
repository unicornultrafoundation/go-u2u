# Docker

Contains the scripts to do u2u benchmarking (only for fakenet now) with Docker from [`docker/`] dir

## Prometheus metrics collection
  From [`docker/monitoring`] (./monitoring) dir
  - `./local-prometheus-on.sh` start prometheus container in local
  - `./prometheus-on.sh` start prometheus container collects metrics from running nodes (so run it after);
  - `docker-compose.yaml` docker compose for starting multiple service configuration(Prometheus, Grafana, Cadvisor, NodeReporter) run with command `docker compose up`
  From [`docker/monitoring/prometheus`] (./monitoring/prometheus) dir
  - `prometheus.yml` configuration for prometheus monitoring defined enpoint
  From [`docker/monitoring/grafana`] (./monitoring/grafana) dir
  - `datasource.yml` data source for grafana analysising and storing
  - see webUI at `http://localhost:9090`;
  - stop: `./prometheus-off.sh`;

See results at:

 - client side: [tx latency](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_ttf&g0.tab=0)
 - client side: [count of sent txs](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_count_sent&g0.tab=0)
 - client side: [count of confirmed txs](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_count_got&g0.tab=0)
 - node side: [tx time2finish](http://localhost:9090/graph?g0.range_input=5m&g0.expr=u2u_tx_ttf&g0.tab=0)
 - node side: [data dir size](http://localhost:9090/graph?g0.range_input=5m&g0.expr=u2u_db_size&g0.tab=0)