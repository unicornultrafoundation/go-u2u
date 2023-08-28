package benchtest

import (
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/metrics"

	"github.com/unicornultrafoundation/go-u2u/logger"
	"github.com/unicornultrafoundation/go-u2u/utils"
)

// Nodes pool.
type Nodes struct {
	tps      chan float64
	sents    chan int64
	receipts chan int64
	conns    []*Sender
	Done     chan struct{}
	cfg      *Config
	logger.Instance
}

var (
	txCountSentMeter = metrics.GetOrRegisterCounter("tx_count_sent", nil)
	txCountGotMeter  = metrics.GetOrRegisterCounter("tx_count_got", nil)
	txLatencyMeter   = metrics.GetOrRegisterHistogram("tx_latency", nil, metrics.NewUniformSample(10000))
)

func NewNodes(cfg *Config, input <-chan *Transaction) *Nodes {
	n := &Nodes{
		tps:      make(chan float64, 1),
		sents:    make(chan int64, 10),
		receipts: make(chan int64, 10),
		Done:     make(chan struct{}),
		cfg:      cfg,
		Instance: logger.New("nodes"),
	}

	for _, url := range cfg.URLs {
		n.add(url)
	}

	n.notifyTPS(0)
	go n.background(input)
	go n.measureGeneratorTPS()
	go n.measureNetworkTPS()
	return n
}

func (n *Nodes) Count() int {
	return len(n.conns)
}

func (n *Nodes) TPS() <-chan float64 {
	return n.tps
}

func (n *Nodes) notifyTPS(tps float64) {
	select {
	case n.tps <- tps:
		break
	default:
	}
}

func (n *Nodes) measureGeneratorTPS() {
	var (
		avgbuff       = utils.NewAvgBuff(50)
		txCount int64 = 0
		start         = time.Now()
	)
	for sent := range n.sents {
		txCountSentMeter.Inc(sent)
		txCount += sent

		dur := time.Since(start).Seconds()
		if dur < 5.0 && txCount < 100 {
			continue
		}

		tps := float64(txCount) / dur
		avgbuff.Push(float64(txCount), dur)

		avg := avgbuff.Avg()
		n.Log.Info("generator TPS", "current", tps, "avg", avg)

		start = time.Now()
		txCount = 0
	}
}

func (n *Nodes) measureNetworkTPS() {
	var (
		avgbuff       = utils.NewAvgBuff(50)
		txCount int64 = 0
		start         = time.Now()
	)
	for got := range n.receipts {
		txCountGotMeter.Inc(got)
		n.Log.Info("txCountGotMeter", "txCountGotMeter", txCountGotMeter.Snapshot())
		txCount += got

		dur := time.Since(start).Seconds()
		if dur < 5.0 && txCount < 100 {
			continue
		}

		tps := float64(txCount) / dur
		avgbuff.Push(float64(txCount), dur)

		avg := avgbuff.Avg()
		n.Log.Info("network TPS", "current", tps, "avg", avg)

		start = time.Now()
		txCount = 0

		n.notifyTPS(avg)
		txLatencyMeter.Update(int64(tps))
		n.Log.Info("TxTpsMeter", "TxTpsMeter", txLatencyMeter.Count())
	}
}

func (n *Nodes) add(url string) {
	c := NewSender(url)
	n.conns = append(n.conns, c)
}

func (n *Nodes) background(input <-chan *Transaction) {
	defer close(n.Done)

	if len(n.conns) < 1 {
		panic("no connections")
	}

	i := 0
	for tx := range input {
		if tx == nil {
			continue
		}
		c := n.conns[i]
		c.Send(n.wrapWithCounter(tx))
		i = (i + 1) % len(n.conns)
	}

	for _, c := range n.conns {
		c.Close()
	}
}

func (n *Nodes) wrapWithCounter(tx *Transaction) *Transaction {
	callback := tx.Callback
	tx.Callback = func(r *types.Receipt, e error) {
		if r != nil {
			n.receipts <- 1
		}
		if callback != nil {
			callback(r, e)
		}
	}

	maker := tx.Make
	tx.Make = func(client *ethclient.Client) (*types.Transaction, error) {
		t, e := maker(client)
		if e == nil {
			n.sents <- 1
		}
		return t, e
	}

	return tx
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
