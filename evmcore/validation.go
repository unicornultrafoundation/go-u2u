// Copyright 2023 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"fmt"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// ValidationOptions define certain differences between transaction validation
// across the different pools without having to duplicate those checks.
type ValidationOptions struct {
	Config *params.ChainConfig // Chain configuration to selectively validate based on current fork rules

	Local        bool
	Accept       uint8    // Bitmap of transaction types that should be accepted for the calling pool
	MaxSize      uint64   // Maximum size of a transaction that the caller can meaningfully handle
	MinTip       *big.Int // Minimum gas tip needed to allow a transaction into the caller pool
	MinGasPrice  *big.Int // The current hard lower bound for gas price required by chain
	PoolGasPrice *big.Int
}

// ValidateTransaction is a helper method to check whether a transaction is valid
// according to the consensus rules, but does not check state-dependent validation
// (balance, nonce, etc).
//
// This check is public to allow different transaction pools to check the basic
// rules without duplicating code and running the risk of missed updates.
func ValidateTransaction(tx *types.Transaction, head *EvmHeader, signer types.Signer, opts *ValidationOptions) error {
	// Ensure transactions not implemented by the calling pool are rejected
	if opts.Accept&(1<<tx.Type()) == 0 {
		return fmt.Errorf("%w: tx type %v not supported by this pool", ErrTxTypeNotSupported, tx.Type())
	}
	// Before performing any expensive validations, sanity check that the tx is
	// smaller than the maximum limit the pool can meaningfully handle
	if uint64(tx.Size()) > opts.MaxSize {
		return fmt.Errorf("%w: transaction size %v, limit %v", ErrOversizedData, tx.Size(), opts.MaxSize)
	}
	// Ensure only transactions that have been enabled are accepted
	if !opts.Config.IsBerlin(head.Number) && tx.Type() != types.LegacyTxType {
		return fmt.Errorf("%w: type %d rejected, pool not yet in Berlin", ErrTxTypeNotSupported, tx.Type())
	}
	if !opts.Config.IsLondon(head.Number) && tx.Type() == types.DynamicFeeTxType {
		return fmt.Errorf("%w: type %d rejected, pool not yet in London", ErrTxTypeNotSupported, tx.Type())
	}
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur for transactions created using the RPC.
	if tx.Value().Sign() < 0 {
		return ErrNegativeValue
	}
	// Ensure the transaction doesn't exceed the current block limit gas
	if head.GasLimit < tx.Gas() {
		return ErrGasLimit
	}
	// Sanity check for extremely large numbers (supported by RLP or RPC)
	if tx.GasFeeCap().BitLen() > 256 {
		return ErrFeeCapVeryHigh
	}
	if tx.GasTipCap().BitLen() > 256 {
		return ErrTipVeryHigh
	}
	// Ensure gasFeeCap is greater than or equal to gasTipCap
	if tx.GasFeeCapIntCmp(tx.GasTipCap()) < 0 {
		return ErrTipAboveFeeCap
	}
	// Ensure the gasprice is high enough to cover the requirement of the calling pool with U2U-specific hard bounds
	if !opts.Local && tx.GasTipCapIntCmp(opts.PoolGasPrice) < 0 {
		return fmt.Errorf("%w: gas tip cap %v, minimum needed %v", ErrUnderpriced, tx.GasTipCap(), opts.PoolGasPrice)
	}
	effectiveGasPrice := new(big.Int).Add(opts.MinTip, opts.MinGasPrice)
	if tx.GasFeeCapIntCmp(effectiveGasPrice) < 0 {
		return fmt.Errorf("%w: gas fee cap %v, minimum needed %v", ErrUnderpriced, tx.GasFeeCap(), effectiveGasPrice)
	}
	// Make sure the transaction is signed properly
	if _, err := types.Sender(signer, tx); err != nil {
		return ErrInvalidSender
	}
	// Ensure the transaction has more gas than the bare minimum needed to cover
	// the transaction metadata
	intrGas, err := IntrinsicGas(tx.Data(), tx.AccessList(), tx.To() == nil)
	if err != nil {
		return err
	}
	if tx.Gas() < intrGas {
		return fmt.Errorf("%w: gas %v, minimum needed %v", ErrIntrinsicGas, tx.Gas(), intrGas)
	}
	// Check for invalid AA params at tx level
	if tx.Type() == types.EIP712TxType && tx.InitiatorAddress() != nil {
		if tx.AAParams() == nil || (tx.AAParams() != nil && tx.AAParams().ValidationTransactionInput == nil) {
			invalidAAParamsTxCounter.Inc(1)
			return ErrInvalidAAParams
		}
	}
	// Check for invalid paymaster params at tx level
	if tx.Type() == types.EIP712TxType && (tx.PaymasterParams() == nil ||
		tx.PaymasterParams().Paymaster == nil || tx.PaymasterParams().PaymasterInput == nil) {
		invalidPaymasterParamsTxCounter.Inc(1)
		return ErrInvalidPaymasterParams
	}
	return nil
}

// ValidationOptionsWithState define certain differences between stateful transaction
// validation across the different pools without having to duplicate those checks.
type ValidationOptionsWithState struct {
	State *state.StateDB // State database to check nonces and balances against

	// FirstNonceGap is an optional callback to retrieve the first nonce gap in
	// the list of pooled transactions of a specific account. If this method is
	// set, nonce gaps will be checked and forbidden. If this method is not set,
	// nonce gaps will be ignored and permitted.
	FirstNonceGap func(addr common.Address) uint64

	// UsedAndLeftSlots is a mandatory callback to retrieve the number of tx slots
	// used and the number still permitted for an account. New transactions will
	// be rejected once the number of remaining slots reaches zero.
	UsedAndLeftSlots func(addr common.Address) (int, int)

	// ExistingExpenditure is a mandatory callback to retrieve the cumulative
	// cost of the already pooled transactions to check for overdrafts.
	ExistingExpenditure func(addr common.Address) *big.Int

	// ExistingCost is a mandatory callback to retrieve an already pooled
	// transaction's cost with the given nonce to check for overdrafts.
	ExistingCost func(addr common.Address, nonce uint64) *big.Int
}

// ValidateTransactionWithState is a helper method to check whether a transaction
// is valid according to the pool's internal state checks (balance, nonce, gaps).
//
// This check is public to allow different transaction pools to check the stateful
// rules without duplicating code and running the risk of missed updates.
func ValidateTransactionWithState(tx *types.Transaction, signer types.Signer, opts *ValidationOptionsWithState) error {
	// Ensure the transaction adheres to nonce ordering
	from, err := signer.Sender(tx) // already validated (and cached), but cleaner to check
	if err != nil {
		log.Error("Normal transaction sender recovery failed", "err", err)
		from, err = validateAASignature(tx, opts)
		if err != nil {
			log.Error("AA Transaction sender recovery failed", "err", err)
			return err
		}
	}
	next := opts.State.GetNonce(from)
	if next > tx.Nonce() {
		return fmt.Errorf("%w: next nonce %v, tx nonce %v", ErrNonceTooLow, next, tx.Nonce())
	}
	// Ensure the transaction doesn't produce a nonce gap in pools that do not
	// support arbitrary orderings
	if opts.FirstNonceGap != nil {
		if gap := opts.FirstNonceGap(from); gap < tx.Nonce() {
			return fmt.Errorf("%w: tx nonce %v, gapped nonce %v", ErrNonceTooHigh, tx.Nonce(), gap)
		}
	}
	// Ensure the transactor has enough funds to cover the transaction costs
	var (
		balance = opts.State.GetBalance(from)
		cost    = tx.Cost()
	)
	if balance.Cmp(cost) < 0 {
		if tx.Type() == types.EIP712TxType && tx.PaymasterParams() != nil && tx.PaymasterParams().Paymaster != nil &&
			opts.State.GetBalance(*tx.PaymasterParams().Paymaster).Cmp(tx.Cost()) < 0 {
			return fmt.Errorf("%w: paymaster balance %v, tx cost %v, overshot %v", ErrInsufficientFunds, balance, cost, new(big.Int).Sub(cost, balance))
		}
		return fmt.Errorf("%w: balance %v, tx cost %v, overshot %v", ErrInsufficientFunds, balance, cost, new(big.Int).Sub(cost, balance))
	}
	// Ensure the transactor has enough funds to cover for replacements or nonce
	// expansions without overdrafts
	spent := opts.ExistingExpenditure(from)
	if prev := opts.ExistingCost(from, tx.Nonce()); prev != nil {
		bump := new(big.Int).Sub(cost, prev)
		need := new(big.Int).Add(spent, bump)
		if balance.Cmp(need) < 0 {
			return fmt.Errorf("%w: balance %v, queued cost %v, tx bumped %v, overshot %v", ErrInsufficientFunds, balance, spent, bump, new(big.Int).Sub(need, balance))
		}
	} else {
		need := new(big.Int).Add(spent, cost)
		if balance.Cmp(need) < 0 {
			return fmt.Errorf("%w: balance %v, queued cost %v, tx cost %v, overshot %v", ErrInsufficientFunds, balance, spent, cost, new(big.Int).Sub(need, balance))
		}
		// Transaction takes a new nonce value out of the pool. Ensure it doesn't
		// overflow the number of permitted transactions from a single account
		// (i.e. max cancellable via out-of-bound transaction).
		if used, left := opts.UsedAndLeftSlots(from); left <= 0 {
			return fmt.Errorf("%w: pooled %d txs", ErrAccountLimitExceeded, used)
		}
	}
	return nil
}

func validateAASignature(tx *types.Transaction, opts *ValidationOptionsWithState) (common.Address, error) {

	return common.Address{}, nil
}
