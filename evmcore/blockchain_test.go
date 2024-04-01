package evmcore

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core"
	"github.com/unicornultrafoundation/go-u2u/core/rawdb"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"
)

func TestEIP712Transition(t *testing.T) {
	chainCfg := params.TestChainConfig
	chainCfg.EIP712Block = big.NewInt(1)
	var (
		// A sender who makes transactions, has some funds
		key, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		address = crypto.PubkeyToAddress(key.PublicKey)
		funds   = big.NewInt(1000000000000000)
		gspec   = &core.Genesis{Config: chainCfg}
		nonce   = uint64(0)
		gen     = func(i int, b *BlockGen) {
			defer func() {
				// Need to assert addTx panics of chain maker here
				// because this function panics on ApplyMessage errors
				if err := recover(); err != nil {
					assert.Equal(t, ErrInvalidPaymasterParams.Error(), errors.New(fmt.Sprint(err)).Error())
					assert.Equal(t, i == 1 || i == 2, true)
				}
			}()
			b.SetCoinbase(common.Address{1})
			// Happy cases when an EIP-712 tx can have the paymaster params field or not
			// i == 3
			pmParams := &types.PaymasterParams{
				Paymaster:      &common.Address{},
				PaymasterInput: []byte{},
			}
			if i == 4 {
				pmParams = nil
			}
			// Error cases with invalid paymaster params
			if i == 1 {
				pmParams.PaymasterInput = nil
			}
			if i == 2 {
				pmParams.Paymaster = nil
			}

			// Construct 1 normal EIP-712 tx for each block
			signer := types.MakeSigner(gspec.Config, big.NewInt(int64(i)))
			tx, err := types.SignNewTx(key, signer, &types.EIP712Tx{
				ChainID:         gspec.Config.ChainID,
				Nonce:           nonce,
				To:              &common.Address{},
				Gas:             30000,
				GasPrice:        b.header.BaseFee,
				PaymasterParams: pmParams,
			})
			if i == 0 {
				// EIP712TxType txs haven't been supported yet at block 0
				// Assert the error then skip adding this tx to block
				assert.Equal(t, types.ErrTxTypeNotSupported, err)
				return
			}
			if i >= 3 {
				nonce++
			}
			b.AddTx(tx)
		}
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db), nil)
		genesis    = MustApplyFakeGenesis(statedb, FakeGenesisTime, map[common.Address]*big.Int{
			address: funds,
		})
	)
	blocks, _, _ := GenerateChain(chainCfg, genesis, db, 5, gen)
	for i := range blocks {
		switch i {
		case 0: // EIP-712 HF hasn't been applied
		case 1: // EIP-712 HF has been applied, nil PaymasterInput field
		case 2: // EIP-712 HF has been applied, nil Paymaster field
			assert.Equal(t, blocks[i].Transactions.Len(), 0)
		case 3: // Happy cases
		case 4:
			assert.Equal(t, blocks[i].Transactions.Len(), 1)
		}
	}
}
