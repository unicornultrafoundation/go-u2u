package benchtest

import (
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/unicornultrafoundation/go-u2u/u2u"

	"gopkg.in/urfave/cli.v1"
)

var (
	gasLimit = uint64(21000)
	gasPrice = new(big.Int).Mul(u2u.FakeEconomyRules().MinGasPrice, big.NewInt(1000))
)

func MakeKeyStore(ctx *cli.Context) (*keystore.KeyStore, error) {
	keydir := ctx.GlobalString(AccKeyStoreDirFlag.Name)
	keydir, err := filepath.Abs(keydir)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(keydir, 0700)
	if err != nil {
		return nil, err
	}

	keyStore := keystore.NewPlaintextKeyStore(keydir)

	return keyStore, nil
}

func openKeyStore(keydir string) (*keystore.KeyStore, error) {
	keydir, err := filepath.Abs(keydir)
	if err != nil {
		return nil, err
	}

	keyStore := keystore.NewKeyStore(keydir, keystore.StandardScryptN, keystore.StandardScryptP)

	return keyStore, nil
}
