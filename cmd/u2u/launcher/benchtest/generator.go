package benchtest

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/unicornultrafoundation/go-u2u/logger"
)

type TxMaker func(*ethclient.Client) (*types.Transaction, error)
type TxCallback func(*types.Receipt, error)

type Transaction struct {
	Make     TxMaker
	Callback TxCallback
	Dsc      string
}

type Generator interface {
	Start() (output <-chan *Transaction)
	Stop()
	SetTPS(tps float64)
}

type genState struct {
	ready          chan struct{}
	notReadyReason string
	Session        interface{}

	logger.Instance
}

func (s *genState) NotReady(reason string) {
	s.notReadyReason = reason
	s.ready = make(chan struct{})
}

func (s *genState) IsReady(done <-chan struct{}) bool {
	if s.ready == nil {
		return true
	}

	s.Log.Warn("waiting", "reason", s.notReadyReason)

	select {
	case <-done:
		return false
	case <-s.ready:
		s.ready = nil
		return true
	}
}

func (s *genState) Ready() {
	close(s.ready)
}
