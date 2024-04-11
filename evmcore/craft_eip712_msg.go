package evmcore

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/eip712"
)

func craftValidateAndPayForPaymasterTransaction(st *StateTransition) ([4]byte, []byte, *ExecutionResult, error) {
	// Apply pmValidateMethod msg
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
	)
	// Temporarily set total initial gas as transaction gas limit
	st.initialGas = st.msg.Gas()
	execRes := Apply(st, msg)
	st.initialGas = 0
	if execRes.Failed() {
		return [4]byte{}, nil, execRes, nil
	}
	// Unpack call result
	result := new(struct {
		Magic   [4]byte
		Context []byte
	})
	if err := IPaymasterABI.UnpackIntoInterface(result, pmValidateMethod, execRes.ReturnData); err != nil {
		return [4]byte{}, nil, nil, err
	}
	return result.Magic, result.Context, execRes, nil
}

func craftValidateAndPayForPaymasterPayload(st *StateTransition) ([]byte, error) {
	// Pack msg payload
	to := big.NewInt(0)
	if st.msg.To() != nil {
		to = new(big.Int).SetBytes(st.msg.To().Bytes())
	}
	tx := &eip712.Transaction{
		TxType:                 big.NewInt(types.EIP712TxType),
		From:                   new(big.Int).SetBytes(st.msg.From().Bytes()),
		To:                     to,
		GasLimit:               new(big.Int).SetUint64(st.msg.Gas()),
		GasPerPubdataByteLimit: big.NewInt(0),
		MaxFeePerGas:           st.msg.GasFeeCap(),
		MaxPriorityFeePerGas:   st.msg.GasTipCap(),
		Nonce:                  new(big.Int).SetUint64(st.msg.Nonce()),
		Value:                  st.msg.Value(),
		Reserved: [4]*big.Int{
			big.NewInt(0),
			big.NewInt(0),
			big.NewInt(0),
			big.NewInt(0),
		},
		Data:      st.msg.Data(),
		Signature: []byte{},
	}
	// Sanity checks
	if st.paymasterParams != nil {
		if st.paymasterParams.Paymaster != nil {
			tx.Paymaster = new(big.Int).SetBytes(st.paymasterParams.Paymaster.Bytes())
		}
		if st.paymasterParams.PaymasterInput != nil {
			tx.PaymasterInput = st.paymasterParams.PaymasterInput
		}
	}
	return IPaymasterABI.Pack(pmValidateMethod,
		common.Hash{1}, common.Hash{2}, common.Hash{3}.Bytes(), tx)
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
