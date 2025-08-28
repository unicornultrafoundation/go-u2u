// Copyright 2024 The go-u2u Authors
// This file is part of the go-u2u library.
//
// The go-u2u library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-u2u library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-u2u library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/hexutil"
)

//go:generate gencodec -type BlobTx -field-override blobTxMarshaling -out gen_blob_tx.go

// BlobTxSidecar contains the actual blob data and commitments.
// This is transmitted alongside blob transactions but not stored on-chain.
type BlobTxSidecar struct {
	Blobs       []Blob       // The actual blob data
	Commitments []Commitment // KZG commitments for each blob
	Proofs      []Proof      // KZG proofs for each blob
}

// Blob represents a 4096 * 32 byte blob of data.
type Blob [131072]byte

// Commitment represents a KZG commitment.
type Commitment [48]byte

// Proof represents a KZG proof.
type Proof [48]byte

// BlobTx represents an EIP-4844 blob transaction.
// It carries blob data for L2 rollup data availability.
type BlobTx struct {
	ChainID    *big.Int           // destination chain ID
	Nonce      uint64             // nonce of sender account  
	GasTipCap  *big.Int           // a.k.a. maxPriorityFeePerGas
	GasFeeCap  *big.Int           // a.k.a. maxFeePerGas
	Gas        uint64             // gas limit
	To         *common.Address    `rlp:"nil"` // nil means contract creation
	Value      *big.Int           // wei amount
	Data       []byte             // contract invocation input data
	AccessList AccessList         // EIP-2930 access list
	BlobFeeCap *big.Int           // a.k.a. maxFeePerBlobGas
	BlobHashes []common.Hash      // versioned hashes of blob commitments

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// Sidecar data (not stored on-chain, transmitted separately)
	Sidecar *BlobTxSidecar `json:"sidecar" rlp:"-"`
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *BlobTx) copy() TxData {
	cpy := &BlobTx{
		Nonce:      tx.Nonce,
		To:         tx.To, // TODO: copy pointed-to address
		Data:       common.CopyBytes(tx.Data),
		Gas:        tx.Gas,
		// These are copied below.
		AccessList: make(AccessList, len(tx.AccessList)),
		BlobHashes: make([]common.Hash, len(tx.BlobHashes)),
		Value:      new(big.Int),
		ChainID:    new(big.Int),
		GasTipCap:  new(big.Int),
		GasFeeCap:  new(big.Int),
		BlobFeeCap: new(big.Int),
		V:          new(big.Int),
		R:          new(big.Int),
		S:          new(big.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	copy(cpy.BlobHashes, tx.BlobHashes)
	
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.BlobFeeCap != nil {
		cpy.BlobFeeCap.Set(tx.BlobFeeCap)
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
	// Note: Sidecar is not copied as it's not part of consensus data
	return cpy
}

// accessList returns the access list of the transaction.
func (tx *BlobTx) accessList() AccessList { return tx.AccessList }

// txType returns the transaction type.
func (tx *BlobTx) txType() byte { return BlobTxType }

// chainID returns the destination chain ID.
func (tx *BlobTx) chainID() *big.Int { return tx.ChainID }

// data returns the transaction data.
func (tx *BlobTx) data() []byte { return tx.Data }

// gas returns the gas limit.
func (tx *BlobTx) gas() uint64 { return tx.Gas }

// gasPrice returns the gas fee cap.
func (tx *BlobTx) gasPrice() *big.Int { return tx.GasFeeCap }

// gasTipCap returns the gas tip cap.
func (tx *BlobTx) gasTipCap() *big.Int { return tx.GasTipCap }

// gasFeeCap returns the gas fee cap.
func (tx *BlobTx) gasFeeCap() *big.Int { return tx.GasFeeCap }

// value returns the transaction value.
func (tx *BlobTx) value() *big.Int { return tx.Value }

// nonce returns the sender account nonce.
func (tx *BlobTx) nonce() uint64 { return tx.Nonce }

// to returns the recipient address.
func (tx *BlobTx) to() *common.Address { return tx.To }

// blobFeeCap returns the blob gas fee cap.
func (tx *BlobTx) blobFeeCap() *big.Int { return tx.BlobFeeCap }

// blobHashes returns the blob versioned hashes.
func (tx *BlobTx) blobHashes() []common.Hash { return tx.BlobHashes }

// rawSignatureValues returns the V, R, S signature values of the transaction.
func (tx *BlobTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

// setSignatureValues sets the signature values of the transaction.
func (tx *BlobTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}

// blobTxMarshaling is used for JSON marshalling of BlobTx.
type blobTxMarshaling struct {
	ChainID    *hexutil.Big
	Nonce      hexutil.Uint64
	GasTipCap  *hexutil.Big
	GasFeeCap  *hexutil.Big
	Gas        hexutil.Uint64
	Value      *hexutil.Big
	Data       hexutil.Bytes
	BlobFeeCap *hexutil.Big
	BlobHashes []common.Hash
	V          *hexutil.Big
	R          *hexutil.Big
	S          *hexutil.Big
}
