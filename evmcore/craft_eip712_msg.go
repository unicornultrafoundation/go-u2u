package evmcore

import "github.com/unicornultrafoundation/go-u2u/core/types"

func craftValidateAndPayForPaymasterTransactionMsg(originalMsg Message) types.Message {
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
