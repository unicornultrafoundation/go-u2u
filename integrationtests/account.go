package integrationtests

import (
	"crypto/ecdsa"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/crypto"
)

type Account struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewAccount() *Account {
	key, _ := crypto.GenerateKey()
	return &Account{
		PrivateKey: key,
	}
}
func (a *Account) Address() common.Address {
	return crypto.PubkeyToAddress(a.PrivateKey.PublicKey)
}
