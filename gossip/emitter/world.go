package emitter

import (
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-hashgraph/native/idx"
	"github.com/unicornultrafoundation/go-hashgraph/native/pos"

	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/valkeystore"
	"github.com/unicornultrafoundation/go-u2u/vecmt"
)

var (
	ErrNotEnoughGasPower = errors.New("not enough gas power")
)

type (
	// External world
	External interface {
		sync.Locker
		Reader

		Check(e *native.EventPayload, parents native.Events) error
		Process(*native.EventPayload) error
		Broadcast(*native.EventPayload)
		Build(*native.MutableEventPayload, func()) error
		DagIndex() *vecmt.Index

		IsBusy() bool
		IsSynced() bool
		PeersNum() int
	}

	// aliases for mock generator
	Signer   valkeystore.SignerI
	TxSigner types.Signer

	// World is an emitter's environment
	World struct {
		External
		TxPool   TxPool
		Signer   valkeystore.SignerI
		TxSigner types.Signer
	}
)

type LlrReader interface {
	GetLowestBlockToDecide() idx.Block
	GetLastBV(id idx.ValidatorID) *idx.Block
	GetBlockRecordHash(idx.Block) *hash.Hash
	GetBlockEpoch(idx.Block) idx.Epoch

	GetLowestEpochToDecide() idx.Epoch
	GetLastEV(id idx.ValidatorID) *idx.Epoch
	GetEpochRecordHash(epoch idx.Epoch) *hash.Hash
}

// Reader is a callback for getting events from an external storage.
type Reader interface {
	LlrReader
	GetLatestBlockIndex() idx.Block
	GetEpochValidators() (*pos.Validators, idx.Epoch)
	GetEvent(hash.Event) *native.Event
	GetEventPayload(hash.Event) *native.EventPayload
	GetLastEvent(epoch idx.Epoch, from idx.ValidatorID) *hash.Event
	GetHeads(idx.Epoch) hash.Events
	GetGenesisTime() native.Timestamp
	GetRules() u2u.Rules
}

type TxPool interface {
	// Has returns an indicator whether txpool has a transaction cached with the
	// given hash.
	Has(hash common.Hash) bool
	// Pending should return pending transactions.
	// The slice should be modifiable by the caller.
	Pending(enforceTips bool) (map[common.Address]types.Transactions, error)

	// Count returns the total number of transactions
	Count() int
}
