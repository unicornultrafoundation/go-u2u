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
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/eip712"
	"github.com/unicornultrafoundation/go-u2u/utils"
)

var (
	// A sender who makes transactions, has some funds
	key, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	address = crypto.PubkeyToAddress(key.PublicKey)
	funds   = big.NewInt(1_000_000_000_000_000_000)
)

func TestEIP712Transition(t *testing.T) {
	chainCfg := params.TestChainConfig
	chainCfg.EIP712Block = big.NewInt(1)
	var (
		gspec = &core.Genesis{Config: chainCfg}
		nonce = uint64(0)
		gen   = func(i int, b *BlockGen) {
			defer func() {
				// Need to assert addTx panics of chain maker here
				// because this function panics on ApplyMessage errors
				if err := recover(); err != nil {
					assert.Equal(t, ErrInvalidPaymasterParams.Error(), errors.New(fmt.Sprint(err)).Error())
					assert.Equal(t, true, i == 1 || i == 2)
				}
			}()
			b.SetCoinbase(common.Address{1})
			// Default paymaster params
			pmParams := &types.PaymasterParams{
				Paymaster:      &common.Address{},
				PaymasterInput: []byte{},
			}
			switch i {
			case 1: // invalid paymaster input
				pmParams.PaymasterInput = nil
				break
			case 2: // invalid paymaster contract address
				pmParams.Paymaster = nil
				break
			case 4: // happy case, nil params field
				pmParams = nil
				break
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
		genesis    = MustApplyFakeGenesis(statedb, FakeGenesisTime, core.GenesisAlloc{
			address: core.GenesisAccount{Balance: funds},
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

func TestSimpleValidateAndPayForPaymaster(t *testing.T) {
	var (
		paymasterSmcAddress = common.HexToAddress("0x3A220f351252089D385b29beca14e27F204c296A")
		pmParams            = &types.PaymasterParams{
			Paymaster:      &paymasterSmcAddress,
			PaymasterInput: []byte{},
		}
		gspec = &core.Genesis{Config: params.TestChainConfig}
		nonce = uint64(0)
		gen   = func(i int, b *BlockGen) {
			b.SetCoinbase(common.Address{1})
			signer := types.LatestSigner(params.TestChainConfig)
			// Construct 1 normal EIP-712 tx for each block
			rawTx := &types.EIP712Tx{
				ChainID:         gspec.Config.ChainID,
				Nonce:           nonce,
				To:              nil,
				Gas:             600000,
				GasPrice:        utils.ToGWEI(250),
				PaymasterParams: pmParams,
				Data:            common.Hex2Bytes("61014060405234801561001157600080fd5b506040516105e13803806105e183398101604081905261003091610191565b815160209283012081519183019190912060e08290526101008190524660a0818152604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f818801819052818301969096526060810194909452608080850193909352308483018190528151808603909301835260c09485019091528151919095012090529190915261012052600080546001600160a01b031916331790556101f4565b634e487b7160e01b600052604160045260246000fd5b600082601f8301126100fd57600080fd5b81516001600160401b0380821115610117576101176100d6565b604051601f8301601f19908116603f0116810190828211818310171561013f5761013f6100d6565b8160405283815260209250868385880101111561015b57600080fd5b600091505b8382101561017d5785820183015181830184015290820190610160565b600093810190920192909252949350505050565b600080604083850312156101a457600080fd5b82516001600160401b03808211156101bb57600080fd5b6101c7868387016100ec565b935060208501519150808211156101dd57600080fd5b506101ea858286016100ec565b9150509250929050565b60805160a05160c05160e05161010051610120516103b06102316000396000505060005050600050506000505060005050600050506103b06000f3fe6080604052600436106100345760003560e01c8063391c91d2146100395780638da5cb5b146100635780639f3cb0b01461009b575b600080fd5b61004c61004736600461015c565b6100b4565b60405161005a929190610245565b60405180910390f35b34801561006f57600080fd5b50600054610083906001600160a01b031681565b6040516001600160a01b03909116815260200161005a565b6100b26100a93660046102b8565b50505050505050565b005b600060608482036100cb57600160e01b9150610124565b60001985016101195760405162461bcd60e51b815260206004820152601660248201527514185e5b585cdd195c8e8815195cdd081c995d995c9d60521b604482015260640160405180910390fd5b631c8e48e960e11b91505b94509492505050565b634e487b7160e01b600052604160045260246000fd5b6000610220828403121561015657600080fd5b50919050565b6000806000806080858703121561017257600080fd5b8435935060208501359250604085013567ffffffffffffffff8082111561019857600080fd5b818701915087601f8301126101ac57600080fd5b8135818111156101be576101be61012d565b604051601f8201601f19908116603f011681019083821181831017156101e6576101e661012d565b816040528281528a60208487010111156101ff57600080fd5b82602086016020830137600060208483010152809650505050606087013591508082111561022c57600080fd5b5061023987828801610143565b91505092959194509250565b63ffffffff60e01b8316815260006020604081840152835180604085015260005b8181101561028257858101830151858201606001528201610266565b506000606082860101526060601f19601f830116850101925050509392505050565b8035600281106102b357600080fd5b919050565b600080600080600080600060c0888a0312156102d357600080fd5b873567ffffffffffffffff808211156102eb57600080fd5b818a0191508a601f8301126102ff57600080fd5b81358181111561030e57600080fd5b8b602082850101111561032057600080fd5b60209283019950975090890135908082111561033b57600080fd5b506103488a828b01610143565b9550506040880135935060608801359250610365608089016102a4565b915060a088013590509295989194975092955056fea264697066735822122031601f3bec0cba129e7c3e06d952abc02aadfeb534b5add1cd6be66a18b84bcf64736f6c634300081400330000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000095061796d6173746572000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000013100000000000000000000000000000000000000000000000000000000000000"),
			}
			switch i {
			case 0:
				rawTx.PaymasterParams = nil
				break
			case 1:
				rawTx.PaymasterParams.PaymasterInput, _ = craftTestValidateAndPayForPaymasterPayload([]byte{0})
				break
			case 2:
				rawTx.PaymasterParams.PaymasterInput, _ = craftTestValidateAndPayForPaymasterPayload([]byte{1})
				break
			case 3:
				rawTx.PaymasterParams.PaymasterInput, _ = craftTestValidateAndPayForPaymasterPayload([]byte{2})
				break
			}
			tx, _ := types.SignNewTx(key, signer, rawTx)
			nonce++
			// Balance tracking
			preAddrBalance := b.statedb.GetBalance(address)
			prePMBalance := b.statedb.GetBalance(paymasterSmcAddress)
			b.AddTx(tx)
			postAddrBalance := b.statedb.GetBalance(address)
			postPMBalance := b.statedb.GetBalance(paymasterSmcAddress)
			if i == 3 {
				// paymaster eligible cases, balance of the sender address stays the same
				// while the balance of the paymaster decreases
				assert.Equal(t, true, postAddrBalance.Cmp(preAddrBalance) == 0)
				assert.Equal(t, true, postPMBalance.Cmp(prePMBalance) < 0)
			} else {
				// paymaster illegible cases, balance of the paymaster stays the same
				// while the balance of the sender address decreases
				assert.Equal(t, true, postAddrBalance.Cmp(preAddrBalance) < 0)
				assert.Equal(t, true, postPMBalance.Cmp(prePMBalance) == 0)
			}
		}
		db         = rawdb.NewMemoryDatabase()
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(db), nil)
	)
	genesis := MustApplyFakeGenesis(statedb, FakeGenesisTime, core.GenesisAlloc{
		address:             core.GenesisAccount{Balance: funds},
		paymasterSmcAddress: core.GenesisAccount{Balance: funds},
	})
	GenerateChain(params.TestChainConfig, genesis, db, 4, gen)
}

func craftTestValidateAndPayForPaymasterPayload(payload []byte) ([]byte, error) {
	var (
		fakeBigInt = big.NewInt(1)
		fakeBytes  = []byte{1}
	)
	// Pack msg payload
	tx := &eip712.Transaction{
		TxType:                 big.NewInt(types.EIP712TxType),
		From:                   fakeBigInt,
		To:                     fakeBigInt,
		GasLimit:               fakeBigInt,
		GasPerPubdataByteLimit: fakeBigInt,
		MaxFeePerGas:           fakeBigInt,
		MaxPriorityFeePerGas:   fakeBigInt,
		Nonce:                  fakeBigInt,
		Value:                  fakeBigInt,
		Reserved: [4]*big.Int{
			fakeBigInt,
			fakeBigInt,
			fakeBigInt,
			fakeBigInt,
		},
		Data:           fakeBytes,
		Signature:      fakeBytes,
		Paymaster:      fakeBigInt,
		PaymasterInput: fakeBytes,
	}
	return IPaymasterABI.Pack(pmValidateAndPayMethod,
		common.Hash{1}, common.BytesToHash(payload), fakeBigInt.Bytes(), tx)
}
