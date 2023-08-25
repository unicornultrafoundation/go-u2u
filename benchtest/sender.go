package benchtest

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/logger"
)

type Sender struct {
	url       string
	input     chan *Transaction
	callbacks map[common.Hash]TxCallback

	work sync.WaitGroup

	logger.Instance
}

func NewSender(url string) *Sender {
	s := &Sender{
		url:       url,
		input:     make(chan *Transaction, 10),
		callbacks: make(map[common.Hash]TxCallback),

		Instance: logger.New("sender"),
	}

	s.work.Add(1)
	go s.background(s.input)

	return s
}

func (s *Sender) Close() {
	if s.input == nil {
		return
	}
	close(s.input)
	s.input = nil

	s.work.Wait()
}

func (s *Sender) Send(tx *Transaction) {
	s.input <- tx
}

func (s *Sender) background(input <-chan *Transaction) {
	defer s.work.Done()
	s.Log.Info("started")
	defer s.Log.Info("stopped")

	var (
		client   *ethclient.Client
		err      error
		ok       bool
		tx       *Transaction
		curBlock = big.NewInt(1)
		maxBlock = big.NewInt(0)
		sbscr    ethereum.Subscription
		headers  = make(chan *types.Header, 1)
	)

	disconnect := func() {
		if sbscr != nil {
			sbscr.Unsubscribe()
			sbscr = nil
		}
		if client != nil {
			client.Close()
			client = nil
			s.Log.Error("disonnect from", "url", s.url)
		}
	}
	defer disconnect()

	for {
		// client connect
		for client == nil {
			client, err = s.connect()
			if err != nil {
				disconnect()
				delay()
				continue
			}
			sbscr, err = s.subscribe(client, headers)
			if err != nil {
				disconnect()
				delay()
				continue
			}
		}

		if curBlock.Cmp(maxBlock) <= 0 {
			err = s.readReceipts(curBlock, client)
			if err != nil {
				disconnect()
				delay()
				continue
			}
			curBlock.Add(curBlock, big.NewInt(1))
		}

		if tx != nil {
			err := s.sendTx(tx, client)
			if err != nil {
				disconnect()
				delay()
				continue
			}
			tx = nil
		}

		// wait for nex task
		select {
		case b := <-headers:
			if maxBlock.Cmp(b.Number) < 0 {
				maxBlock.Set(b.Number)
				if curBlock.Cmp(big.NewInt(1)) == 0 {
					curBlock.Set(maxBlock)
				}
			}
		case tx, ok = <-input:
			if !ok {
				return
			}
		}
	}
}

func (s *Sender) sendTx(tx *Transaction, client *ethclient.Client) (err error) {
	var (
		t      *types.Transaction
		txHash common.Hash
	)
	err = try(func() error {
		t, err = tx.Make(client)
		return err
	})
	if t != nil {
		txHash = t.Hash()
	}

	if err == nil {
		if tx.Callback != nil {
			s.callbacks[txHash] = tx.Callback
		}
		s.Log.Info("tx sending ok", "hash", txHash, "dsc", tx.Dsc)
		return
	}

	switch err.Error() {
	case "already known",
		fmt.Sprintf("known transaction: %x", txHash),
		evmcore.ErrNonceTooLow.Error(),
		evmcore.ErrReplaceUnderpriced.Error():
		s.Log.Warn("tx sending skip", "hash", txHash, "dsc", tx.Dsc, "cause", err)
		err = nil
	default:
		s.Log.Error("tx sending err", "hash", txHash, "dsc", tx.Dsc, "err", err)
	}
	if tx.Callback != nil {
		tx.Callback(nil, err)
	}

	return
}

func (s *Sender) connect() (*ethclient.Client, error) {
	client, err := ethclient.Dial(s.url)
	if err != nil {
		s.Log.Error("connect to", "url", s.url, "err", err)
		return nil, err
	}
	s.Log.Info("connect to", "url", s.url)
	return client, nil
}

func (s *Sender) subscribe(client *ethclient.Client, headers chan *types.Header) (sbscr ethereum.Subscription, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	try(func() error {
		sbscr, err = client.SubscribeNewHead(ctx, headers)
		return err
	})
	if err != nil {
		s.Log.Error("subscribe to", "url", s.url, "err", err)
		return
	}
	s.Log.Info("subscribe to", "url", s.url)
	return
}

func (s *Sender) readReceipts(n *big.Int, client *ethclient.Client) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	blk, err := client.BlockByNumber(ctx, n)
	if err != nil {
		s.Log.Error("new block", "block", n, "err", err)
		return
	}
	s.Log.Info("new block", "block", n)

	for index, tx := range blk.Transactions() {
		txHash := tx.Hash()

		callback := s.callbacks[txHash]
		if callback == nil {
			continue
		}

		var r *types.Receipt
		err = try(func() error {
			r, err = client.TransactionReceipt(ctx, txHash)
			return err
		})
		if err != nil {
			s.Log.Error("new receipt", "block", n, "index", index, "tx", txHash, "err", err)
			switch err.Error() {
			case "not found":
				err = nil // ignore
			default:
				return
			}
		}

		callback(r, err)
		delete(s.callbacks, txHash)
		s.Log.Info("new receipt", "block", n, "index", index, "tx", txHash)
	}

	return
}

func delay() {
	<-time.After(2 * time.Second)
}

func try(f func() error) (err error) {
	defer func() {
		if catch := recover(); catch != nil {
			err = fmt.Errorf("client panic: %v", catch)
		}
	}()

	err = f()
	return
}
