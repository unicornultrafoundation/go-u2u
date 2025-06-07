package integrationtests

import (
	"context"
	"fmt"
	"math/big"

	go_u2u "github.com/unicornultrafoundation/go-u2u"
	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
)

// CraftSFCTx crafts a transaction to the SFC contract then sign and send it
func (n *IntegrationTestNet) CraftSFCTx(account *Account, abi abi.ABI, to *common.Address, value *big.Int, name string, args ...interface{}) (*types.Receipt, error) {
	client, err := n.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the network: %w", err)
	}
	defer client.Close()
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}
	nonce, err := client.NonceAt(context.Background(), account.Address(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}
	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	data, err := abi.Pack(name, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transaction data: %w, name: %s, args: %v", err, name, args)
	}
	gas, err := client.EstimateGas(context.Background(), go_u2u.CallMsg{
		From:  account.Address(),
		To:    to,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}
	transaction, err := types.SignTx(types.NewTx(&types.AccessListTx{
		ChainID:  chainId,
		Nonce:    nonce,
		Gas:      gas,
		GasPrice: price,
		To:       to,
		Value:    value,
		Data:     data,
	}), types.NewLondonSigner(chainId), account.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}
	return n.Run(transaction)
}
