package evmcore

import (
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

func craftValidateAndPayForPaymasterTransaction(st *StateTransition) ([4]byte, []byte, error) {
	// Pack msg payload and apply
	payload, err := IPaymasterABI.Pack("validateAndPayForPaymasterTransaction")
	if err != nil {
		return [4]byte{}, nil, err
	}
	if err != nil {
		return [4]byte{}, nil, err
	}
	res := ConstructAndApplySmcCallMsg(st, payload)
	if len(res.ReturnData) == 0 {
		return [4]byte{}, nil, nil
	}
	// Unpack call result
	result := new(struct {
		magic   [4]byte
		context []byte
	})
	if err := IPaymasterABI.UnpackIntoInterface(result, "mint", res.ReturnData); err != nil {
		return [4]byte{}, nil, err
	}
	return result.magic, result.context, nil
}

func ConstructAndApplySmcCallMsg(st *StateTransition, payload []byte) *ExecutionResult {
	// Apply this fake message just to get the function output
	msg := types.NewMessage(
		st.msg.From(),
		st.msg.To(),
		0,
		st.msg.Value(),
		st.msg.Gas(),
		st.msg.GasPrice(),
		st.msg.GasFeeCap(),
		st.msg.GasTipCap(),
		payload,
		nil,
		true,
		nil,
	)
	return Apply(st, msg)
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
