package types

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/libs/common"
)

type AccountAbstractionTx struct {
	ChainID  *big.Int
	Nonce    uint64          // nonce of sender account
	GasPrice *big.Int        // wei per gas
	Gas      uint64          // gas limit
	To       *common.Address `rlp:"nil"` // nil means contract creation
	Value    *big.Int        // wei amount
	Data     []byte          // contract invocation input data
	V, R, S  *big.Int        // signature values
}

func (tx *AccountAbstractionTx) copy() TxData {
	cpy := &AccountAbstractionTx{
		Nonce: tx.Nonce,
		To:    tx.To,
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are initialized below.
		ChainID:  new(big.Int),
		Value:    new(big.Int),
		GasPrice: new(big.Int),
		V:        new(big.Int),
		R:        new(big.Int),
		S:        new(big.Int),
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.GasPrice != nil {
		cpy.GasPrice.Set(tx.GasPrice)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

func (tx *AccountAbstractionTx) txType() byte           { return AccountAbstractionTxType }
func (tx *AccountAbstractionTx) chainID() *big.Int      { return tx.ChainID }
func (tx *AccountAbstractionTx) protected() bool        { return true }
func (tx *AccountAbstractionTx) accessList() AccessList { return nil }
func (tx *AccountAbstractionTx) data() []byte           { return tx.Data }
func (tx *AccountAbstractionTx) gas() uint64            { return tx.Gas }
func (tx *AccountAbstractionTx) gasFeeCap() *big.Int    { return tx.GasPrice }
func (tx *AccountAbstractionTx) gasTipCap() *big.Int    { return tx.GasPrice }
func (tx *AccountAbstractionTx) gasPrice() *big.Int     { return tx.GasPrice }
func (tx *AccountAbstractionTx) value() *big.Int        { return tx.Value }
func (tx *AccountAbstractionTx) nonce() uint64          { return tx.Nonce }
func (tx *AccountAbstractionTx) to() *common.Address    { return tx.To }

func (tx *AccountAbstractionTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *AccountAbstractionTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}
