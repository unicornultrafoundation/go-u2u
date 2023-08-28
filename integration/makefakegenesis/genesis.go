package makefakegenesis

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-hashgraph/native/idx"
	"github.com/unicornultrafoundation/go-hashgraph/native/pos"
	utypes "github.com/unicornultrafoundation/go-hashgraph/types"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb/memorydb"

	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/integration/makegenesis"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/native/drivertype"
	"github.com/unicornultrafoundation/go-u2u/native/iblockproc"
	"github.com/unicornultrafoundation/go-u2u/native/ier"
	"github.com/unicornultrafoundation/go-u2u/native/validatorpk"
	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driver"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driver/drivercall"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driverauth"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/evmwriter"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/netinit"
	netinitcall "github.com/unicornultrafoundation/go-u2u/u2u/contracts/netinit/netinitcalls"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfc"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfclib"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesis"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesis/gpos"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesisstore"
)

var (
	FakeGenesisTime = native.Timestamp(1608600000 * time.Second)
)

// FakeKey gets n-th fake private key.
func FakeKey(n idx.ValidatorID) *ecdsa.PrivateKey {
	return evmcore.FakeKey(uint32(n))
}

func FakeGenesisStore(num idx.Validator, balance, stake *big.Int) *genesisstore.Store {
	return FakeGenesisStoreWithRules(num, balance, stake, u2u.FakeNetRules())
}

func FakeGenesisStoreWithRules(num idx.Validator, balance, stake *big.Int, rules u2u.Rules) *genesisstore.Store {
	return FakeGenesisStoreWithRulesAndStart(num, balance, stake, rules, 2, 1)
}

func FakeGenesisStoreWithRulesAndStart(num idx.Validator, balance, stake *big.Int, rules u2u.Rules, epoch idx.Epoch, block idx.Block) *genesisstore.Store {
	builder := makegenesis.NewGenesisBuilder(memorydb.NewProducer(""))

	validators := GetFakeValidators(num)

	// add balances to validators
	var delegations []drivercall.Delegation
	for _, val := range validators {
		builder.AddBalance(val.Address, balance)
		delegations = append(delegations, drivercall.Delegation{
			Address:            val.Address,
			ValidatorID:        val.ID,
			Stake:              stake,
			LockedStake:        new(big.Int),
			LockupFromEpoch:    0,
			LockupEndTime:      0,
			LockupDuration:     0,
			EarlyUnlockPenalty: new(big.Int),
			Rewards:            new(big.Int),
		})
	}

	// deploy essential contracts
	// pre deploy NetworkInitializer
	builder.SetCode(netinit.ContractAddress, netinit.GetContractBin())
	// pre deploy NodeDriver
	builder.SetCode(driver.ContractAddress, driver.GetContractBin())
	// pre deploy NodeDriverAuth
	builder.SetCode(driverauth.ContractAddress, driverauth.GetContractBin())
	// pre deploy SFC
	builder.SetCode(sfc.ContractAddress, sfc.GetContractBin())
	// pre deploy SFCLib
	builder.SetCode(sfclib.ContractAddress, sfclib.GetContractBin())
	// set non-zero code for pre-compiled contracts
	builder.SetCode(evmwriter.ContractAddress, []byte{0})

	builder.SetCurrentEpoch(ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{
			BlockState: iblockproc.BlockState{
				LastBlock: iblockproc.BlockCtx{
					Idx:     block - 1,
					Time:    FakeGenesisTime,
					Atropos: hash.Event{},
				},
				FinalizedStateRoot:    hash.Hash{},
				EpochGas:              0,
				EpochCheaters:         utypes.Cheaters{},
				CheatersWritten:       0,
				ValidatorStates:       make([]iblockproc.ValidatorBlockState, 0),
				NextValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
				DirtyRules:            nil,
				AdvanceEpochs:         0,
			},
			EpochState: iblockproc.EpochState{
				Epoch:             epoch - 1,
				EpochStart:        FakeGenesisTime,
				PrevEpochStart:    FakeGenesisTime - 1,
				EpochStateRoot:    hash.Zero,
				Validators:        pos.NewBuilder().Build(),
				ValidatorStates:   make([]iblockproc.ValidatorEpochState, 0),
				ValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
				Rules:             rules,
			},
		},
		Idx: epoch - 1,
	})

	var owner common.Address
	if num != 0 {
		owner = validators[0].Address
	}

	blockProc := makegenesis.DefaultBlockProc()
	genesisTxs := GetGenesisTxs(epoch-2, validators, builder.TotalSupply(), delegations, owner)
	err := builder.ExecuteGenesisTxs(blockProc, genesisTxs)
	if err != nil {
		panic(err)
	}

	return builder.Build(genesis.Header{
		GenesisID:   builder.CurrentHash(),
		NetworkID:   rules.NetworkID,
		NetworkName: rules.Name,
	})
}

func txBuilder() func(calldata []byte, addr common.Address) *types.Transaction {
	nonce := uint64(0)
	return func(calldata []byte, addr common.Address) *types.Transaction {
		tx := types.NewTransaction(nonce, addr, common.Big0, 1e10, common.Big0, calldata)
		nonce++
		return tx
	}
}

func GetGenesisTxs(sealedEpoch idx.Epoch, validators gpos.Validators, totalSupply *big.Int, delegations []drivercall.Delegation, driverOwner common.Address) types.Transactions {
	buildTx := txBuilder()
	internalTxs := make(types.Transactions, 0, 15)
	// initialization
	calldata := netinitcall.InitializeAll(sealedEpoch, totalSupply, sfc.ContractAddress, sfclib.ContractAddress, driverauth.ContractAddress, driver.ContractAddress, evmwriter.ContractAddress, driverOwner)
	internalTxs = append(internalTxs, buildTx(calldata, netinit.ContractAddress))
	// push genesis validators
	for _, v := range validators {
		calldata := drivercall.SetGenesisValidator(v)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}
	// push genesis delegations
	for _, delegation := range delegations {
		calldata := drivercall.SetGenesisDelegation(delegation)
		internalTxs = append(internalTxs, buildTx(calldata, driver.ContractAddress))
	}
	return internalTxs
}

func GetFakeValidators(num idx.Validator) gpos.Validators {
	validators := make(gpos.Validators, 0, num)

	for i := idx.ValidatorID(1); i <= idx.ValidatorID(num); i++ {
		key := FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		pubkeyraw := crypto.FromECDSAPub(&key.PublicKey)
		validators = append(validators, gpos.Validator{
			ID:      i,
			Address: addr,
			PubKey: validatorpk.PubKey{
				Raw:  pubkeyraw,
				Type: validatorpk.Types.Secp256k1,
			},
			CreationTime:     FakeGenesisTime,
			CreationEpoch:    0,
			DeactivatedTime:  0,
			DeactivatedEpoch: 0,
			Status:           0,
		})
	}

	return validators
}

// Only for benchmark testing purpose
func FakeGenesisStoreForBenchTest(num idx.Validator, ks *keystore.KeyStore, balance, initBalance, stake *big.Int) *genesisstore.Store {
	return FakeGenesisStoreWithRulesAndStartForBenchTest(num, ks, balance, initBalance, stake, u2u.FakeNetRules(), 2, 1)
}

func FakeGenesisStoreWithRulesAndStartForBenchTest(num idx.Validator, ks *keystore.KeyStore, balance, initBalance, stake *big.Int, rules u2u.Rules, epoch idx.Epoch, block idx.Block) *genesisstore.Store {
	builder := makegenesis.NewGenesisBuilder(memorydb.NewProducer(""))
	
	validators := GetFakeValidators(num)

	// add balances to validators
	var delegations []drivercall.Delegation
	for _, val := range validators {
		builder.AddBalance(val.Address, balance)
		delegations = append(delegations, drivercall.Delegation{
			Address:            val.Address,
			ValidatorID:        val.ID,
			Stake:              stake,
			LockedStake:        new(big.Int),
			LockupFromEpoch:    0,
			LockupEndTime:      0,
			LockupDuration:     0,
			EarlyUnlockPenalty: new(big.Int),
			Rewards:            new(big.Int),
		})
	}

	// add balances to fake accounts
	for _, acc := range ks.Accounts() {
		if err := ks.Unlock(acc, ""); err != nil {
			panic(err)
		}
		builder.AddBalance(acc.Address, initBalance)
	}

	// deploy essential contracts
	// pre deploy NetworkInitializer
	builder.SetCode(netinit.ContractAddress, netinit.GetContractBin())
	// pre deploy NodeDriver
	builder.SetCode(driver.ContractAddress, driver.GetContractBin())
	// pre deploy NodeDriverAuth
	builder.SetCode(driverauth.ContractAddress, driverauth.GetContractBin())
	// pre deploy SFC
	builder.SetCode(sfc.ContractAddress, sfc.GetContractBin())
	// set non-zero code for pre-compiled contracts
	builder.SetCode(evmwriter.ContractAddress, []byte{0})

	builder.SetCurrentEpoch(ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{
			BlockState: iblockproc.BlockState{
				LastBlock: iblockproc.BlockCtx{
					Idx:     block - 1,
					Time:    FakeGenesisTime,
					Atropos: hash.Event{},
				},
				FinalizedStateRoot:    hash.Hash{},
				EpochGas:              0,
				EpochCheaters:         utypes.Cheaters{},
				CheatersWritten:       0,
				ValidatorStates:       make([]iblockproc.ValidatorBlockState, 0),
				NextValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
				DirtyRules:            nil,
				AdvanceEpochs:         0,
			},
			EpochState: iblockproc.EpochState{
				Epoch:             epoch - 1,
				EpochStart:        FakeGenesisTime,
				PrevEpochStart:    FakeGenesisTime - 1,
				EpochStateRoot:    hash.Zero,
				Validators:        pos.NewBuilder().Build(),
				ValidatorStates:   make([]iblockproc.ValidatorEpochState, 0),
				ValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
				Rules:             rules,
			},
		},
		Idx: epoch - 1,
	})

	var owner common.Address
	if num != 0 {
		owner = validators[0].Address
	}

	blockProc := makegenesis.DefaultBlockProc()
	genesisTxs := GetGenesisTxs(epoch-2, validators, builder.TotalSupply(), delegations, owner)
	err := builder.ExecuteGenesisTxs(blockProc, genesisTxs)
	if err != nil {
		panic(err)
	}

	return builder.Build(genesis.Header{
		GenesisID:   builder.CurrentHash(),
		NetworkID:   rules.NetworkID,
		NetworkName: rules.Name,
	})
}
