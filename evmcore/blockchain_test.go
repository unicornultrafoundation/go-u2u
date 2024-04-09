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
			rawTx := &types.EIP712Tx{
				ChainID:         gspec.Config.ChainID,
				Nonce:           nonce,
				To:              nil,
				Gas:             600000,
				GasPrice:        utils.ToGWEI(250),
				PaymasterParams: pmParams,
				Data:            common.Hex2Bytes("61014060405234801561001157600080fd5b5060405161096c38038061096c83398101604081905261003091610191565b815160209283012081519183019190912060e08290526101008190524660a0818152604080517f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f818801819052818301969096526060810194909452608080850193909352308483018190528151808603909301835260c09485019091528151919095012090529190915261012052600080546001600160a01b031916331790556101f4565b634e487b7160e01b600052604160045260246000fd5b600082601f8301126100fd57600080fd5b81516001600160401b0380821115610117576101176100d6565b604051601f8301601f19908116603f0116810190828211818310171561013f5761013f6100d6565b8160405283815260209250868385880101111561015b57600080fd5b600091505b8382101561017d5785820183015181830184015290820190610160565b600093810190920192909252949350505050565b600080604083850312156101a457600080fd5b82516001600160401b03808211156101bb57600080fd5b6101c7868387016100ec565b935060208501519150808211156101dd57600080fd5b506101ea858286016100ec565b9150509250929050565b60805160a05160c05160e051610100516101205161070061026c6000396000818161011001526103c7015260008181610164015261041601526000818161021d01526103f1015260008181610251015261034a01526000818160a801526103740152600081816101e9015261039e01526107006000f3fe6080604052600436106100915760003560e01c80638da5cb5b116100595780638da5cb5b146101865780639f3cb0b0146101be578063a9e91e54146101d7578063caac6c821461020b578063da28b5271461023f57600080fd5b80632b437d4814610096578063391c91d2146100dd5780635d2dab0b146100fe5780636b03e22214610132578063712ac56d14610152575b600080fd5b3480156100a257600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b6040519081526020015b60405180910390f35b6100f06100eb366004610493565b610273565b6040516100d492919061057c565b34801561010a57600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b34801561013e57600080fd5b506100ca61014d3660046105db565b61029b565b34801561015e57600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b34801561019257600080fd5b506000546101a6906001600160a01b031681565b6040516001600160a01b0390911681526020016100d4565b6101d56101cc366004610608565b50505050505050565b005b3480156101e357600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b34801561021757600080fd5b506100ca7f000000000000000000000000000000000000000000000000000000000000000081565b34801561024b57600080fd5b506101a67f000000000000000000000000000000000000000000000000000000000000000081565b600060608482036102875760009150610292565b631c8e48e960e11b91505b94509492505050565b60006102f07f2bb156903cb7269fcabf8c90e4a20f3c4eaf1dde6b407dbc50a24ac21c9fab696040805160208101929092528101849052606001604051602081830303815290604052805190602001206102f6565b92915050565b600061030061033d565b60405161190160f01b6020820152602281019190915260428101839052606201604051602081830303815290604052805190602001209050919050565b6000306001600160a01b037f00000000000000000000000000000000000000000000000000000000000000001614801561039657507f000000000000000000000000000000000000000000000000000000000000000046145b156103c057507f000000000000000000000000000000000000000000000000000000000000000090565b50604080517f00000000000000000000000000000000000000000000000000000000000000006020808301919091527f0000000000000000000000000000000000000000000000000000000000000000828401527f000000000000000000000000000000000000000000000000000000000000000060608301524660808301523060a0808401919091528351808403909101815260c0909201909252805191012090565b634e487b7160e01b600052604160045260246000fd5b6000610220828403121561048d57600080fd5b50919050565b600080600080608085870312156104a957600080fd5b8435935060208501359250604085013567ffffffffffffffff808211156104cf57600080fd5b818701915087601f8301126104e357600080fd5b8135818111156104f5576104f5610464565b604051601f8201601f19908116603f0116810190838211818310171561051d5761051d610464565b816040528281528a602084870101111561053657600080fd5b82602086016020830137600060208483010152809650505050606087013591508082111561056357600080fd5b506105708782880161047a565b91505092959194509250565b63ffffffff60e01b8316815260006020604081840152835180604085015260005b818110156105b95785810183015185820160600152820161059d565b506000606082860101526060601f19601f830116850101925050509392505050565b6000602082840312156105ed57600080fd5b5035919050565b80356002811061060357600080fd5b919050565b600080600080600080600060c0888a03121561062357600080fd5b873567ffffffffffffffff8082111561063b57600080fd5b818a0191508a601f83011261064f57600080fd5b81358181111561065e57600080fd5b8b602082850101111561067057600080fd5b60209283019950975090890135908082111561068b57600080fd5b506106988a828b0161047a565b95505060408801359350606088013592506106b5608089016105f4565b915060a088013590509295989194975092955056fea2646970667358221220cfe90f67e26a4eecde3eb5c017d883a935bf1b6444733a7a4dd8e53e3805c9f364736f6c634300081400330000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000095061796d6173746572000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000013100000000000000000000000000000000000000000000000000000000000000"),
			}
			switch i {
			case 0:
				rawTx.PaymasterParams = nil
				break
			case 1:
				rawTx.PaymasterParams.PaymasterInput, _ = craftTestValidateAndPayForPaymasterPayload(true)
				break
			case 2:
				rawTx.PaymasterParams.PaymasterInput, _ = craftTestValidateAndPayForPaymasterPayload(false)
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
			if i == 1 {
				assert.Equal(t, true, postAddrBalance.Cmp(preAddrBalance) == 0)
			} else {
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
	GenerateChain(params.TestChainConfig, genesis, db, 3, gen)
}

func craftTestValidateAndPayForPaymasterPayload(valid bool) ([]byte, error) {
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
	if !valid {
		fakeBytes = []byte{}
	}
	return IPaymasterABI.Pack("validateAndPayForPaymasterTransaction",
		common.Hash{1}, common.BytesToHash(fakeBytes), fakeBigInt.Bytes(), tx)
}
