package evmcore

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

func craftValidateAndPayForPaymasterTransaction(originalMsg Message, paymasterParams *types.PaymasterParams) types.Message {
	payload, err := IPaymasterABI.Pack("validateAndPayForPaymasterTransaction")
	if err != nil {
		return err
	}
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	return types.NewMessage(
		originalMsg.From(),
		originalMsg.To(),
		originalMsg.Nonce(),
		originalMsg.Value(),
		originalMsg.Gas(),
		originalMsg.GasPrice(),
		originalMsg.GasFeeCap(),
		originalMsg.GasTipCap(),
		originalMsg.Data(),
		originalMsg.AccessList(),
		true,
		nil,
	)
}

func ConstructAndApplySmcCallMsg(statedb *state.StateDB, header *types.Header, bc vm.ChainContext, cfg vm.Config, payload []byte) ([]byte, error) {
	msg := types.NewMessage(
		s.ContractAddress,
		&s.ContractAddress,
		0,
		big.NewInt(0),
		100000000,
		big.NewInt(0),
		payload,
		false,
	)
	return Apply(s.logger, bc, statedb, header, cfg, msg)
}

// Apply ...
func Apply(logger log.Logger, bc vm.ChainContext, statedb *state.StateDB, header *types.Header, cfg kvm.Config, msg types.Message) ([]byte, error) {
	// Create a new context to be used in the KVM environment
	context := vm.NewKVMContext(msg, header, bc)
	vmenv := kvm.NewKVM(context, kvm.TxContext{}, statedb, configs.MainnetChainConfig, cfg)
	sender := kvm.AccountRef(msg.From())
	ret, _, vmerr := vmenv.Call(sender, *msg.To(), msg.Data(), msg.Gas(), msg.Value())
	if vmerr != nil {
		return nil, vmerr
	}
	// Update the state with pending changes
	statedb.Finalise(true)
	return ret, nil
}
