package evmcore

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core"
	"github.com/unicornultrafoundation/go-u2u/core/rawdb"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"
)

func TestEIP712Transition(t *testing.T) {
	var (
		// A sender who makes transactions, has some funds
		key, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		address = crypto.PubkeyToAddress(key.PublicKey)
		funds   = big.NewInt(1000000000000000)
		gspec   = &core.Genesis{Config: params.TestChainConfig}
		gen     = func(i int, b *BlockGen) {
			b.SetCoinbase(common.Address{1})
			pmParams := &types.PaymasterParams{
				Paymaster:      &common.Address{},
				PaymasterInput: []byte{},
			}
			//if i == 2 {
			//	pmParams.Paymaster = nil
			//}
			//if i == 3 {
			//	pmParams = nil
			//}
			// Construct 1 normal EIP-712 tx for each block
			signer := types.LatestSigner(gspec.Config)
			tx, err := types.SignNewTx(key, signer, &types.EIP712Tx{
				ChainID:         gspec.Config.ChainID,
				Nonce:           uint64(i),
				To:              &common.Address{},
				Gas:             30000,
				GasPrice:        b.header.BaseFee,
				PaymasterParams: pmParams,
			})
			if err == nil {
				b.AddTx(tx)
			} else {
				//t.Logf("block %v, tx: %+v, err: %v\n", i, tx, err)
			}
		}
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db), nil)
		genesis    = MustApplyFakeGenesis(statedb, FakeGenesisTime, map[common.Address]*big.Int{
			address: funds,
		})
	)
	blocks, _, _ := GenerateChain(params.TestChainConfig, genesis, db, 4, gen)
	for i := range blocks {
		fmt.Printf("block %v, txs: %+v\n", i, blocks[i].Transactions)
	}
}
