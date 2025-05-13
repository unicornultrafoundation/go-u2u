package ethapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/params"
)

func newTestBackend(t *testing.T, n int, gspec *core.Genesis, generator func(i int, b *evmcore.BlockGen)) *Backend {
	return nil
}

func setupReceiptBackend(t *testing.T, genBlocks int) (*Backend, []common.Hash) {
	// Initialize test accounts
	var (
		acc1Key, _ = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
		acc2Key, _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
		acc1Addr   = crypto.PubkeyToAddress(acc1Key.PublicKey)
		acc2Addr   = crypto.PubkeyToAddress(acc2Key.PublicKey)
		contract   = common.HexToAddress("0000000000000000000000000000000000031ec7")
		genesis    = &core.Genesis{
			Config: params.TestChainConfig,
			Alloc: core.GenesisAlloc{
				acc1Addr: {Balance: big.NewInt(params.Ether)},
				acc2Addr: {Balance: big.NewInt(params.Ether)},
				// Token with transfer that emit Transfer
				contract: {Balance: big.NewInt(params.Ether), Code: common.FromHex("0x608060405234801561001057600080fd5b506004361061002b5760003560e01c8063a9059cbb14610030575b600080fd5b61004a6004803603810190610045919061016a565b610060565b60405161005791906101c5565b60405180910390f35b60008273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040516100bf91906101ef565b60405180910390a36001905092915050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610101826100d6565b9050919050565b610111816100f6565b811461011c57600080fd5b50565b60008135905061012e81610108565b92915050565b6000819050919050565b61014781610134565b811461015257600080fd5b50565b6000813590506101648161013e565b92915050565b60008060408385031215610181576101806100d1565b5b600061018f8582860161011f565b92505060206101a085828601610155565b9150509250929050565b60008115159050919050565b6101bf816101aa565b82525050565b60006020820190506101da60008301846101b6565b92915050565b6101e981610134565b82525050565b600060208201905061020460008301846101e0565b9291505056fea2646970667358221220b469033f4b77b9565ee84e0a2f04d496b18160d26034d54f9487e57788fd36d564736f6c63430008120033")},
			},
		}
		signer   = types.LatestSignerForChainID(params.TestChainConfig.ChainID)
		txHashes = make([]common.Hash, genBlocks)
	)
	backend := newTestBackend(t, genBlocks, genesis, func(i int, b *evmcore.BlockGen) {
		var (
			tx  *types.Transaction
			err error
		)
		switch i {
		case 0:
			// transfer 1000wei
			tx, err = types.SignTx(types.NewTx(&types.LegacyTx{Nonce: uint64(i), To: &acc2Addr, Value: big.NewInt(1000), Gas: params.TxGas, GasPrice: b.BaseFee(), Data: nil}), types.HomesteadSigner{}, acc1Key)
		case 1:
			// create contract
			tx, err = types.SignTx(types.NewTx(&types.LegacyTx{Nonce: uint64(i), To: nil, Gas: 53100, GasPrice: b.BaseFee(), Data: common.FromHex("0x60806040")}), signer, acc1Key)
		case 2:
			// with logs
			// transfer(address to, uint256 value)
			data := fmt.Sprintf("0xa9059cbb%s%s", common.HexToHash(common.BigToAddress(big.NewInt(int64(i + 1))).Hex()).String()[2:], common.BytesToHash([]byte{byte(i + 11)}).String()[2:])
			tx, err = types.SignTx(types.NewTx(&types.LegacyTx{Nonce: uint64(i), To: &contract, Gas: 60000, GasPrice: b.BaseFee(), Data: common.FromHex(data)}), signer, acc1Key)
		case 3:
			// dynamic fee with logs
			// transfer(address to, uint256 value)
			data := fmt.Sprintf("0xa9059cbb%s%s", common.HexToHash(common.BigToAddress(big.NewInt(int64(i + 1))).Hex()).String()[2:], common.BytesToHash([]byte{byte(i + 11)}).String()[2:])
			fee := big.NewInt(500)
			fee.Add(fee, b.BaseFee())
			tx, err = types.SignTx(types.NewTx(&types.DynamicFeeTx{Nonce: uint64(i), To: &contract, Gas: 60000, Value: big.NewInt(1), GasTipCap: big.NewInt(500), GasFeeCap: fee, Data: common.FromHex(data)}), signer, acc1Key)
		case 4:
			// access list with contract create
			accessList := types.AccessList{{
				Address:     contract,
				StorageKeys: []common.Hash{{0}},
			}}
			tx, err = types.SignTx(types.NewTx(&types.AccessListTx{Nonce: uint64(i), To: nil, Gas: 58100, GasPrice: b.BaseFee(), Data: common.FromHex("0x60806040"), AccessList: accessList}), signer, acc1Key)
		}
		if err != nil {
			t.Errorf("failed to sign tx: %v", err)
		}
		if tx != nil {
			b.AddTx(tx)
			txHashes[i] = tx.Hash()
		}
	})
	return backend, txHashes
}

func TestRPCGetTransactionReceipt(t *testing.T) {
	t.Parallel()

	var (
		backend, txHashes = setupReceiptBackend(t, 5)
		api               = NewPublicTransactionPoolAPI(*backend, new(AddrLocker))
	)

	var testSuite = []struct {
		txHash common.Hash
		want   string
	}{
		// 0. normal success
		{
			txHash: txHashes[0],
			want: `{
				"blockHash": "0x1356e49a24d4504e450b303aa770f4ae13c29b9ffacaea1d7dd4043396229dd9",
				"blockNumber": "0x1",
				"contractAddress": null,
				"cumulativeGasUsed": "0x5208",
				"effectiveGasPrice": "0x342770c0",
				"from": "0x703c4b2bd70c169f5717101caee543299fc946c7",
				"gasUsed": "0x5208",
				"logs": [],
				"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"status": "0x1",
				"to": "0x0d3ab14bbad3d99f4203bd7a11acb94882050e7e",
				"transactionHash": "0x644a31c354391520d00e95b9affbbb010fc79ac268144ab8e28207f4cf51097e",
				"transactionIndex": "0x0",
				"type": "0x0"
			}`,
		},
		// 1. create contract
		{
			txHash: txHashes[1],
			want: `{
				"blockHash": "0x4fc27a4efa7fb8faa04b12b53ec8c8424ab4c21aab1323846365f000e8b4a594",
				"blockNumber": "0x2",
				"contractAddress": "0xae9bea628c4ce503dcfd7e305cab4e29e7476592",
				"cumulativeGasUsed": "0xcf4e",
				"effectiveGasPrice": "0x2db16291",
				"from": "0x703c4b2bd70c169f5717101caee543299fc946c7",
				"gasUsed": "0xcf4e",
				"logs": [],
				"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"status": "0x1",
				"to": null,
				"transactionHash": "0x340e58cda5086495010b571fe25067fecc9954dc4ee3cedece00691fa3f5904a",
				"transactionIndex": "0x0",
				"type": "0x0"
			}`,
		},
		// 2. with logs success
		{
			txHash: txHashes[2],
			want: `{
				"blockHash": "0x73385c190219326907524b0020ef453ebc450eaa971ebce16f79e2d23e7e8d4d",
				"blockNumber": "0x3",
				"contractAddress": null,
				"cumulativeGasUsed": "0x5e28",
				"effectiveGasPrice": "0x281c2534",
				"from": "0x703c4b2bd70c169f5717101caee543299fc946c7",
				"gasUsed": "0x5e28",
				"logs": [
					{
						"address": "0x0000000000000000000000000000000000031ec7",
						"topics": [
							"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
							"0x000000000000000000000000703c4b2bd70c169f5717101caee543299fc946c7",
							"0x0000000000000000000000000000000000000000000000000000000000000003"
						],
						"data": "0x000000000000000000000000000000000000000000000000000000000000000d",
						"blockNumber": "0x3",
						"transactionHash": "0x9dbf43ec9afc8d711932618616471088f66ba4f25fd5c672d97473d02dae967f",
						"transactionIndex": "0x0",
						"blockHash": "0x73385c190219326907524b0020ef453ebc450eaa971ebce16f79e2d23e7e8d4d",
						"logIndex": "0x0",
						"removed": false
					}
				],
				"logsBloom": "0x00000000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000800000000000000008000000000000000000000000000000000020000000080000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000800000000000000000400000000002000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000",
				"status": "0x1",
				"to": "0x0000000000000000000000000000000000031ec7",
				"transactionHash": "0x9dbf43ec9afc8d711932618616471088f66ba4f25fd5c672d97473d02dae967f",
				"transactionIndex": "0x0",
				"type": "0x0"
			}`,
		},
		// 3. dynamic tx with logs success
		{
			txHash: txHashes[3],
			want: `{
				"blockHash": "0x77c3f8919590e0e68db4ce74a3da3140ac3e96dd3d078a48db1da4c08b07503d",
				"blockNumber": "0x4",
				"contractAddress": null,
				"cumulativeGasUsed": "0x538d",
				"effectiveGasPrice": "0x2325c3e8",
				"from": "0x703c4b2bd70c169f5717101caee543299fc946c7",
				"gasUsed": "0x538d",
				"logs": [],
				"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"status": "0x0",
				"to": "0x0000000000000000000000000000000000031ec7",
				"transactionHash": "0x672e3e39adf23b5656989b7a36e54d54004b1866f53871113bc52e137edb9faf",
				"transactionIndex": "0x0",
				"type": "0x2"
			}`,
		},
		// 4. access list tx with create contract
		{
			txHash: txHashes[4],
			want: `{
				"blockHash": "0x08e23d8e3711a21fbb8becd7de22fda8fb0a49fba14e1be763d00f99063627e1",
				"blockNumber": "0x5",
				"contractAddress": "0xfdaa97661a584d977b4d3abb5370766ff5b86a18",
				"cumulativeGasUsed": "0xe01a",
				"effectiveGasPrice": "0x1ecb3f75",
				"from": "0x703c4b2bd70c169f5717101caee543299fc946c7",
				"gasUsed": "0xe01a",
				"logs": [],
				"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"status": "0x1",
				"to": null,
				"transactionHash": "0x8f3c4e2663af0312d508ebd8587f0c88dccbbc8a9bcc322421ff4bc28c456a92",
				"transactionIndex": "0x0",
				"type": "0x1"
			}`,
		},
		// 5. txhash empty
		{
			txHash: common.Hash{},
			want:   `null`,
		},
		// 6. txhash not found
		{
			txHash: common.HexToHash("deadbeef"),
			want:   `null`,
		},
	}

	for i, tt := range testSuite {
		var (
			result interface{}
			err    error
		)
		result, err = api.GetTransactionReceipt(context.Background(), tt.txHash)
		if err != nil {
			t.Errorf("test %d: want no error, have %v", i, err)
			continue
		}
		data, err := json.Marshal(result)
		if err != nil {
			t.Errorf("test %d: json marshal error", i)
			continue
		}
		want, have := tt.want, string(data)
		require.JSONEqf(t, want, have, "test %d: json not match, want: %s, have: %s", i, want, have)
	}
}
