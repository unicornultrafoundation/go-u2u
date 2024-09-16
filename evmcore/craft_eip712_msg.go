package evmcore

import (
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"math/big"
)

func craftValidateTxTransaction(st *StateTransition) ([4]byte, *ExecutionResult, error) {
	// Apply a fake ValidateTransaction msg just to get the result and gas used
	msg := types.NewMessage(
		st.msg.From(),
		st.paymasterParams.Paymaster,
		0,
		st.msg.Value(),
		st.msg.Gas(),
		st.msg.GasPrice(),
		st.msg.GasFeeCap(),
		st.msg.GasTipCap(),
		st.paymasterParams.PaymasterInput,
		nil,
		true,
		nil,
		nil,
		nil,
	)
	// Temporarily set total initial gas as transaction gas limit
	st.initialGas = st.msg.Gas()
	execRes := Apply(st, msg)
	st.initialGas = 0 // reset
	if execRes.Failed() {
		return [4]byte{}, execRes, nil
	}
	// Unpack call result
	result := new(struct {
		Magic   [4]byte
		Context []byte
	})
	if err := IPaymasterABI.UnpackIntoInterface(result, aaValidateTxMethod, execRes.ReturnData); err != nil {
		return [4]byte{}, execRes, err
	}
	return result.Magic, execRes, nil
}

func craftPrepareForPaymasterTransaction(st *StateTransition) (*ExecutionResult, error) {
	to := st.msg.From()
	// Apply a fake PrepareForPaymasterTransaction msg just to get the result and gas used
	msg := types.NewMessage(
		st.msg.From(),
		&to,
		0,
		st.msg.Value(),
		st.msg.Gas(),
		big.NewInt(0),
		st.msg.GasFeeCap(),
		st.msg.GasTipCap(),
		st.aaParams.PrepareForPaymasterInput,
		nil,
		true,
		nil,
		nil,
		nil,
	)
	// Temporarily set total initial gas as transaction gas limit
	st.initialGas = st.msg.Gas()
	execRes := Apply(st, msg)
	st.initialGas = 0 // reset
	return execRes, nil
}

func craftValidateAndPayForPaymasterTransaction(st *StateTransition) ([4]byte, []byte, *ExecutionResult, error) {
	// Apply a fake ValidateAndPayForPaymasterTransaction msg just to get the result and gas used
	msg := types.NewMessage(
		st.msg.From(),
		st.paymasterParams.Paymaster,
		0,
		st.msg.Value(),
		st.msg.Gas(),
		st.msg.GasPrice(),
		st.msg.GasFeeCap(),
		st.msg.GasTipCap(),
		st.paymasterParams.PaymasterInput,
		nil,
		true,
		nil,
		nil,
		nil,
	)
	// Temporarily set total initial gas as transaction gas limit
	st.initialGas = st.msg.Gas()
	execRes := Apply(st, msg)
	st.initialGas = 0 // reset
	if execRes.Failed() {
		return [4]byte{}, nil, execRes, nil
	}
	// Unpack call result
	result := new(struct {
		Magic   [4]byte
		Context []byte
	})
	if err := IPaymasterABI.UnpackIntoInterface(result, pmValidateAndPayMethod, execRes.ReturnData); err != nil {
		return [4]byte{}, nil, execRes, err
	}
	return result.Magic, result.Context, execRes, nil
}

func Apply(st *StateTransition, msg types.Message) *ExecutionResult {
	sender := vm.AccountRef(msg.From())
	var (
		ret   []byte
		vmerr error // vm errors do not effect consensus and are therefore not assigned to err
	)
	ret, st.gas, vmerr = st.evm.Call(sender, *msg.To(), msg.Data(), msg.Gas(), msg.Value())
	return &ExecutionResult{
		UsedGas:    st.gasUsed(),
		Err:        vmerr,
		ReturnData: ret,
	}
}
