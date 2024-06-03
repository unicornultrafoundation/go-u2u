package evmcore

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/rawdb"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"
)

var (
	// A sender who makes transactions, has some funds
	key, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	address = crypto.PubkeyToAddress(key.PublicKey)
	funds   = big.NewInt(1_000_000_000_000_000_000)
	code = append([]byte{
		byte(vm.PUSH1), 0x1,
		byte(vm.TLOAD),
		byte(vm.PUSH1), 0x1,
		byte(vm.SSTORE),
	}, make([]byte, 32-6)...)
)

func GenerateChainWithGenesis(gen func(int, *BlockGen)) ([]*EvmBlock, []types.Receipts, DummyChain) {
	var (
		n          = 5
		config     = params.TestChainConfig
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db), nil)
	)
	config.EIP712Block = big.NewInt(1)
	return GenerateChain(config, MustApplyFakeGenesis(statedb, FakeGenesisTime, map[common.Address]*big.Int{
		address: funds,
		dest: types.GenesisAlloc{
			Balance: 0,
			Code: code,
		}
	}), db, n, gen)
}

func TestEVMTransition(t *testing.T) {
	dest := crypto.CreateAddress(address, 0)
	// Transient Storage Test
	code = append(
		append(
			[]byte{
				byte(vm.PUSH1), 0x1,
				byte(vm.PUSH1), 0x1,
				byte(vm.TSTORE),
				byte(vm.PUSH32),
			}, code...),
		[]byte{
			byte(vm.PUSH1), 0x0,
			byte(vm.MSTORE),
			byte(vm.PUSH1), 0x6,
			byte(vm.PUSH1), 0x0,
			byte(vm.RETURN),
		}...,
	)
	signer := types.HomesteadSigner{}
	nonce := uint64(0)
	_, receipts, _ := GenerateChainWithGenesis(func(i int, b *BlockGen) {
		fee := big.NewInt(1)
		if b.header.BaseFee != nil {
			fee = b.header.BaseFee
		}
		b.SetCoinbase(common.Address{1})

		switch i {
		case 0:
			tx, _ := types.SignNewTx(key, signer, &types.LegacyTx{
				Nonce:    nonce,
				GasPrice: fee,
				Gas:      100000,
				Data:     code,
			})
			nonce++
			b.AddTx(tx)

		case 1:
			tx, _ := types.SignNewTx(key, signer, &types.LegacyTx{
				Nonce:    nonce,
				GasPrice: fee,
				Gas:      100000,
				To:       &dest,
			})
			nonce++
			b.AddTx(tx)
		case 2:
		}
	})
	print("xxx:", receipts[0][0].Status)
	print("xxx:", receipts[1][0].Status)
}
