package benchtest

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/unicornultrafoundation/go-u2u/logger"
)

type TransfersGenerator struct {
	tps     uint32
	chainId *big.Int
	ks      *keystore.KeyStore
	accs    []accounts.Account
	nonces  []uint64

	position uint
	state    genState

	work sync.WaitGroup
	done chan struct{}
	sync.Mutex

	logger.Instance
}

func NewTransfersGenerator(cfg *Config, ks *keystore.KeyStore) *TransfersGenerator {
	g := &TransfersGenerator{
		chainId: big.NewInt(cfg.ChainId),
		ks:      ks,

		Instance: logger.New("gentxs_transfer"),
	}
	g.state.Log = g.Log

	for _, acc := range ks.Accounts() {
		if acc.Address == cfg.Payer {
			continue
		}
		if err := ks.Unlock(acc, ""); err != nil {
			panic(err)
		}
		g.accs = append(g.accs, acc)
	}

	g.nonces = make([]uint64, len(g.accs))

	return g
}

func (g *TransfersGenerator) Start() <-chan *Transaction {
	g.Lock()
	defer g.Unlock()

	if g.done != nil {
		return nil
	}
	g.done = make(chan struct{})

	output := make(chan *Transaction, 10)
	g.work.Add(1)
	go g.background(output)

	return output
}

func (g *TransfersGenerator) Stop() {
	g.Lock()
	defer g.Unlock()

	if g.done == nil {
		return
	}

	close(g.done)
	g.work.Wait()
	g.done = nil
}

func (g *TransfersGenerator) getTPS() float64 {
	tps := atomic.LoadUint32(&g.tps)
	return float64(tps)
}

func (g *TransfersGenerator) SetTPS(tps float64) {
	x := uint32(math.Ceil(tps))
	atomic.StoreUint32(&g.tps, x)
}

func (g *TransfersGenerator) background(output chan<- *Transaction) {
	defer g.work.Done()
	defer close(output)

	g.Log.Info("started")
	defer g.Log.Info("stopped")

	for {
		begin := time.Now()
		var (
			generating time.Duration
			sending    time.Duration
		)

		tps := g.getTPS()
		for count := tps; count > 0; count-- {
			begin := time.Now()
			tx := g.Yield()
			generating += time.Since(begin)

			begin = time.Now()
			select {
			case output <- tx:
				sending += time.Since(begin)
				continue
			case <-g.done:
				return
			}
		}

		spent := time.Since(begin)
		if spent >= time.Second {
			g.Log.Warn("exceeded performance", "tps", tps, "generating", generating, "sending", sending)
			continue
		}

		select {
		case <-time.After(time.Second - spent):
			continue
		case <-g.done:
			return
		}
	}
}

func (g *TransfersGenerator) Yield() *Transaction {
	if !g.state.IsReady(g.done) {
		return nil
	}
	tx := g.generate(g.position, &g.state)
	g.Log.Info("generated tx", "position", g.position, "dsc", tx.Dsc)
	g.position++

	return tx
}

func (g *TransfersGenerator) generate(position uint, state *genState) *Transaction {
	count := uint(len(g.accs))

	var (
		from     accounts.Account
		to       accounts.Account
		amount   *big.Int
		callback TxCallback
	)

	from = g.accs[position%count]
	to = g.accs[(position+1)%count]
	amount = big.NewInt(1e5)

	// wait every cicle
	if position%count == 0 {
		state.NotReady("transfer cicle")
		callback = func(r *types.Receipt, e error) {
			state.Ready()
		}
	}

	askNonce := (position%count == 0)

	return &Transaction{
		Make:     g.transferTx(from, to, amount, askNonce, g.nonces[position%count:]),
		Dsc:      fmt.Sprintf("%s --> %s", from.Address.String(), to.Address.String()),
		Callback: callback,
	}
}

func (g *TransfersGenerator) transferTx(from, to accounts.Account, amount *big.Int, askNonce bool, cache []uint64) TxMaker {
	return func(client *ethclient.Client) (tx *types.Transaction, err error) {
		nonce := cache[0]
		if askNonce {
			nonce, err = client.PendingNonceAt(context.Background(), from.Address)
			if err != nil {
				return
			}
		}
		cache[0] = nonce + 1

		tx = types.NewTransaction(
			nonce,
			to.Address,
			amount,
			gasLimit,
			gasPrice,
			[]byte{},
		)

		tx, err = g.ks.SignTx(from, tx, g.chainId)
		if err != nil {
			panic(err)
		}

		err = client.SendTransaction(context.Background(), tx)
		return
	}
}
