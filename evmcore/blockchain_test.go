package evmcore

import (
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core"
	"github.com/unicornultrafoundation/go-u2u/core/rawdb"
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
		gen     = func(i int, b *BlockGen) {
			b.SetCoinbase(common.Address{1})
			pmParams := &types.PaymasterParams{
				Paymaster:      nil,
				PaymasterInput: nil,
			}
			if i == 2 {
				pmParams = nil
			}
			// Construct 1 normal EIP-712 tx for each block
			signer := types.LatestSigner(gspec.Config)
			tx, _ := types.SignNewTx(key, signer, &types.EIP712Tx{
				ChainID:         gspec.Config.ChainID,
				Nonce:           0,
				To:              &common.Address{},
				Gas:             30000,
				GasPrice:        b.header.BaseFee,
				PaymasterParams: pmParams,
			})
			b.AddTx(tx)
		}
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db), nil)
		genesis    = MustApplyFakeGenesis(statedb, FakeGenesisTime, map[common.Address]*big.Int{
			address: funds,
		})
	)
	blocks, _, chain := GenerateChain(chainCfg, genesis, db, 3, gen)

}
