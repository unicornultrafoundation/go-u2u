package types

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
)

type PaymasterParams struct {
	Paymaster      *common.Address
	PaymasterInput []byte
}

// EIP712Tx is the transaction data of regular Ethereum transactions.
type EIP712Tx struct {
	ChainID         *big.Int
	Nonce           uint64
	GasPrice        *big.Int
	Gas             uint64
	To              *common.Address `rlp:"nil"` // nil means contract creation
	Value           *big.Int
	Data            []byte
	PaymasterParams *PaymasterParams
	V, R, S         *big.Int // signature values
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *EIP712Tx) copy() TxData {
	cpy := &EIP712Tx{
		Nonce:           tx.Nonce,
		To:              tx.To, // TODO: copy pointed-to address
		Data:            common.CopyBytes(tx.Data),
		Gas:             tx.Gas,
		PaymasterParams: tx.PaymasterParams,
		// These are initialized below.
		Value:    new(big.Int),
		ChainID:  new(big.Int),
		GasPrice: new(big.Int),
		V:        new(big.Int),
		R:        new(big.Int),
		S:        new(big.Int),
	}
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
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

// accessors for innerTx.
func (tx *EIP712Tx) txType() byte                      { return EIP712TxType }
func (tx *EIP712Tx) chainID() *big.Int                 { return tx.ChainID }
func (tx *EIP712Tx) accessList() AccessList            { return nil }
func (tx *EIP712Tx) data() []byte                      { return tx.Data }
func (tx *EIP712Tx) gas() uint64                       { return tx.Gas }
func (tx *EIP712Tx) gasPrice() *big.Int                { return tx.GasPrice }
func (tx *EIP712Tx) gasTipCap() *big.Int               { return tx.GasPrice }
func (tx *EIP712Tx) gasFeeCap() *big.Int               { return tx.GasPrice }
func (tx *EIP712Tx) value() *big.Int                   { return tx.Value }
func (tx *EIP712Tx) nonce() uint64                     { return tx.Nonce }
func (tx *EIP712Tx) to() *common.Address               { return tx.To }
func (tx *EIP712Tx) paymasterParams() *PaymasterParams { return tx.PaymasterParams }

func (tx *EIP712Tx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *EIP712Tx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.V, tx.R, tx.S = v, r, s
}
