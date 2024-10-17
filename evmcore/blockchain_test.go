package evmcore

import (
	"bytes"
	"fmt"
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

	signer = types.HomesteadSigner{}
)

func GetTx(nonce *uint64, to *common.Address, data []byte) *types.Transaction {
	tx, _ := types.SignNewTx(key, signer, &types.LegacyTx{
		Nonce:    *nonce,
		GasPrice: big.NewInt(1),
		Gas:      100_000,
		To:       to,
		Data:     data,
	})
	*nonce++
	return tx
}

func GenerateChainWithGenesis(gen func(int, *BlockGen)) ([]*EvmBlock, []types.Receipts, DummyChain) {
	var (
		n          = 4
		config     = params.AllProtocolChanges
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db), nil)
	)
	config.EIP712Block = big.NewInt(2)
	return GenerateChain(config, MustApplyFakeGenesis(statedb, FakeGenesisTime, map[common.Address]*big.Int{
		address: funds,
		//dest: types.GenesisAlloc{ Balance: 0, Code: code }
	}), db, n, gen)
}

func TestEVMTransition(t *testing.T) {
	// Transient Storage Test
	code0 := bytes.Join([][]byte{{
		byte(vm.PUSH1), 0x1,
		byte(vm.PUSH1), 0x1,
		byte(vm.TSTORE),
		byte(vm.PUSH32),
	}, make([]byte, 32), {
		byte(vm.PUSH1), 0x1,
		byte(vm.TLOAD),
		byte(vm.PUSH1), 0x1,
		byte(vm.SSTORE),
	}, {
		byte(vm.PUSH1), 0x0,
		byte(vm.MSTORE),
		byte(vm.PUSH1), 0x6,
		byte(vm.PUSH1), 0x0,
		byte(vm.RETURN),
	}}, nil)

	// PUSH0, MCOPY
	code1 := bytes.Join([][]byte{{
		byte(vm.PUSH1), 0x20,
		byte(vm.PUSH0),
		byte(vm.PUSH1), 0x20,
		byte(vm.MCOPY),
	}, {
		byte(vm.PUSH1), 0x6,
		byte(vm.PUSH0),
		byte(vm.RETURN),
	}}, nil)

	code2 := bytes.Join([][]byte{{
		byte(vm.PREVRANDAO),
		byte(vm.INVALID),
	}, {
		byte(vm.PUSH1), 0x2,
		byte(vm.PUSH0),
		byte(vm.RETURN),
	}}, nil)

	expects := []uint64{0, 0, 0, 1, 1, 1, 1, 1, 1}

	nonce := uint64(0)
	_, receipts, _ := GenerateChainWithGenesis(func(i int, b *BlockGen) {
		b.SetCoinbase(common.Address{1})
		switch i {
		case 0:
			b.AddTx(GetTx(&nonce, nil, code0))
			b.AddTx(GetTx(&nonce, nil, code1))
			b.AddTx(GetTx(&nonce, nil, code2))
		case 1:
			b.AddTx(GetTx(&nonce, nil, code0))
			b.AddTx(GetTx(&nonce, nil, code1))
			b.AddTx(GetTx(&nonce, nil, code1))
		case 2:
			end := nonce
			dest := crypto.CreateAddress(address, end)
			b.AddTx(GetTx(&nonce, &dest, nil))
			end--
			dest = crypto.CreateAddress(address, end)
			b.AddTx(GetTx(&nonce, &dest, nil))
			end--
			dest = crypto.CreateAddress(address, end)
			b.AddTx(GetTx(&nonce, &dest, nil))
		}
	})

	i := 0
	for _, block := range receipts {
		for _, receipt := range block {
			fmt.Println("_tx:", i, "=", receipt.Status, receipt.GasUsed)
			if i == len(expects) || receipt.Status != expects[i] {
				t.Fatal()
			}
			i++
		}
	}
	if i != len(expects) {
		t.Fatal("receipts count mismatch")
	}
}
