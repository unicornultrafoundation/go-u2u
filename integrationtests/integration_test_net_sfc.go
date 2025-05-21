package integrationtests

import (
	"context"
	"fmt"
	"math/big"

	go_u2u "github.com/unicornultrafoundation/go-u2u"
	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// CraftSFCTx crafts a transaction to the SFC contract then sign and send it
func (n *IntegrationTestNet) CraftSFCTx(account *Account, abi abi.ABI, value *big.Int, name string, args ...interface{}) error {
	client, err := n.GetClient()
	if err != nil {
		return fmt.Errorf("failed to connect to the network: %w", err)
	}
	defer client.Close()
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}
	nonce, err := client.NonceAt(context.Background(), account.Address(), nil)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}
	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}
	data, err := abi.Pack(name, args...)
	if err != nil {
		return fmt.Errorf("failed to pack transaction data: %w, name: %s, args: %v", err, name, args)
	}
	gas, err := client.EstimateGas(context.Background(), go_u2u.CallMsg{
		From:  account.Address(),
		To:    &SfcAddress,
		Value: value,
		Data:  data,
	})
	if err != nil {
		log.Info("failed to estimate gas: %w", "err", err)
		gas = 1000000
	}
	transaction, err := types.SignTx(types.NewTx(&types.AccessListTx{
		ChainID:  chainId,
		Nonce:    nonce,
		Gas:      gas,
		GasPrice: price,
		To:       &SfcAddress,
		Value:    value,
		Data:     data,
	}), types.NewLondonSigner(chainId), account.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}
	_, err = n.Run(transaction)
	return err
}

// SfcGetStorageAt returns the storage from the SFC state at the given address, key and
// block number.
// The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta-block numbers are also allowed.
func (n *IntegrationTestNet) SfcGetStorageAt(account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	client, err := n.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the U2U client: %w", err)
	}
	defer client.Close()
	return client.SfcStorageAt(context.Background(), account, key, blockNumber)
}

// CheckIntegrity returns integrity of the SFC state, compared with the world state.
func (n *IntegrationTestNet) CheckIntegrity(blockNumber *big.Int) (bool, error) {
	client, err := n.GetClient()
	if err != nil {
		return false, fmt.Errorf("failed to connect to the U2U client: %w", err)
	}
	defer client.Close()
	return client.CheckIntegrity(context.Background(), blockNumber)
}
