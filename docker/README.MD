# Docker

Contains the scripts to do u2u benchmarking (only for fakenet now) with Docker from [`docker/`] dir

## Image
- Dockerfile.u2u is used to run a docker image containing u2u build binary

## Prometheus metrics collection
  From [`docker/monitoring`] (./monitoring) dir

  1. For starting development monitoring service
  - `cd development` 
  - `./prometheus-on.sh` start prometheus container in development mode

  2. For starting Devnet monitoring service
  - `cd devnet` 
  - `./prometheus-on.sh` start prometheus container collects metrics from running devnet nodes (so run it after starting node instance);
  - stop: `./prometheus-off.sh`

  3. For starting Testnet monitoring service
  - `cd testnet`
  - `./prometheus-on.sh` start prometheus container collects metrics from running testnet nodes (so run it after starting node instance);
  - stop: `./prometheus-off.sh`

  4. For starting Grafana, Cadvisor, NodeReporter for tracking and reporting analytics
  - `docker-compose.yaml` docker compose for starting multiple service configuration run with command `docker compose up`

  5. From [`docker/monitoring/prometheus`] (./monitoring/prometheus) dir
  - `prometheus.yml` configuration for prometheus monitoring defined enpoint

  6. From [`docker/monitoring/grafana`] (./monitoring/grafana) dir
  - `datasource.yml` data source for grafana analysising and storing

  7. See webUI at `http://localhost:9090`;

## See results at:

 - client side: [tx latency](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_ttf&g0.tab=0)
 - client side: [count of sent txs](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_count_sent&g0.tab=0)
 - client side: [count of confirmed txs](http://localhost:9090/graph?g0.range_input=5m&g0.expr=txstorm_tx_count_got&g0.tab=0)
 - node side: [tx time2finish](http://localhost:9090/graph?g0.range_input=5m&g0.expr=u2u_tx_ttf&g0.tab=0)
 - node side: [data dir size](http://localhost:9090/graph?g0.range_input=5m&g0.expr=u2u_db_size&g0.tab=0)