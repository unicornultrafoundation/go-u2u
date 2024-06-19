package epochcheck

import (
	"errors"
	"fmt"
	
	base "github.com/unicornultrafoundation/go-helios/eventcheck/epochcheck"
	"github.com/unicornultrafoundation/go-helios/native/idx"

	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

var (
	ErrTooManyParents    = errors.New("event has too many parents")
	ErrTooBigGasUsed     = errors.New("event uses too much gas power")
	ErrWrongGasUsed      = errors.New("event has incorrect gas power")
	ErrUnderpriced       = errors.New("event transaction underpriced")
	ErrTooBigExtra       = errors.New("event extra data is too large")
	ErrWrongVersion      = errors.New("event has wrong version")
	ErrUnsupportedTxType = errors.New("unsupported tx type")
	ErrNotRelevant       = base.ErrNotRelevant
	ErrAuth              = base.ErrAuth
)

// Reader returns currents epoch and its validators group.
type Reader interface {
	base.Reader
	GetEpochRules() (u2u.Rules, idx.Epoch)
}

// Checker which require only current epoch info
type Checker struct {
	Base   *base.Checker
	reader Reader
}

func New(reader Reader) *Checker {
	return &Checker{
		Base:   base.New(reader),
		reader: reader,
	}
}

func CalcGasPowerUsed(e native.EventPayloadI, rules u2u.Rules) uint64 {
	txsGas := uint64(0)
	for _, tx := range e.Txs() {
		txsGas += tx.Gas()
	}

	gasCfg := rules.Economy.Gas

	parentsGas := uint64(0)
	if idx.Event(len(e.Parents())) > rules.Dag.MaxFreeParents {
		parentsGas = uint64(idx.Event(len(e.Parents()))-rules.Dag.MaxFreeParents) * gasCfg.ParentGas
	}
	extraGas := uint64(len(e.Extra())) * gasCfg.ExtraDataGas

	mpsGas := uint64(len(e.MisbehaviourProofs())) * gasCfg.MisbehaviourProofGas

	bvsGas := uint64(0)
	if e.BlockVotes().Start != 0 {
		bvsGas = gasCfg.BlockVotesBaseGas + uint64(len(e.BlockVotes().Votes))*gasCfg.BlockVoteGas
	}

	ersGas := uint64(0)
	if e.EpochVote().Epoch != 0 {
		ersGas = gasCfg.EpochVoteGas
	}

	return txsGas + parentsGas + extraGas + gasCfg.EventGas + mpsGas + bvsGas + ersGas
}

func (v *Checker) checkGas(e native.EventPayloadI, rules u2u.Rules) error {
	if e.GasPowerUsed() > rules.Economy.Gas.MaxEventGas {
		return ErrTooBigGasUsed
	}
	if e.GasPowerUsed() != CalcGasPowerUsed(e, rules) {
		return ErrWrongGasUsed
	}
	return nil
}

func CheckTxs(txs types.Transactions, rules u2u.Rules) error {
	var accept uint8 = 0 | 1<<types.LegacyTxType
	if rules.Upgrades.Berlin {
		accept = accept | 1<<types.AccessListTxType
	}
	if rules.Upgrades.London {
		accept = accept | 1<<types.DynamicFeeTxType
	}
	if rules.Upgrades.EIP712 {
		accept = accept | 1<<types.EIP712TxType
	}
	for _, tx := range txs {
		if accept&(1<<tx.Type()) == 0 {
			return fmt.Errorf("%w: tx type %v not supported by this pool", ErrUnsupportedTxType, tx.Type())
		}
		if tx.GasFeeCapIntCmp(rules.Economy.MinGasPrice) < 0 {
			return ErrUnderpriced
		}
	}
	return nil
}

// Validate event
func (v *Checker) Validate(e native.EventPayloadI) error {
	if err := v.Base.Validate(e); err != nil {
		return err
	}
	rules, epoch := v.reader.GetEpochRules()
	// Check epoch of the rules to prevent a race condition
	if e.Epoch() != epoch {
		return base.ErrNotRelevant
	}
	if idx.Event(len(e.Parents())) > rules.Dag.MaxParents {
		return ErrTooManyParents
	}
	if uint32(len(e.Extra())) > rules.Dag.MaxExtraData {
		return ErrTooBigExtra
	}
	if err := v.checkGas(e, rules); err != nil {
		return err
	}
	if err := CheckTxs(e.Txs(), rules); err != nil {
		return err
	}
	version := uint8(0)
	if rules.Upgrades.Llr {
		version = 1
	}
	if e.Version() != version {
		return ErrWrongVersion
	}
	return nil
}
