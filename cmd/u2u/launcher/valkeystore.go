package launcher

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/node"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/native/validatorpk"
	"github.com/unicornultrafoundation/go-u2u/valkeystore"
)

func addFakeValidatorKey(ctx *cli.Context, key *ecdsa.PrivateKey, pubkey validatorpk.PubKey, valKeystore valkeystore.RawKeystoreI) {
	// add fake validator key
	if key != nil && !valKeystore.Has(pubkey) {
		err := valKeystore.Add(pubkey, crypto.FromECDSA(key), validatorpk.FakePassword)
		if err != nil {
			utils.Fatalf("Failed to add fake validator key: %v", err)
		}
	}
}

func getValKeystoreDir(cfg node.Config) string {
	_, _, keydir, err := cfg.AccountConfig()
	if err != nil {
		utils.Fatalf("Failed to setup account config: %v", err)
	}
	return keydir
}

// makeValidatorPasswordList reads password lines from the file specified by the global --validator.password flag.
func makeValidatorPasswordList(ctx *cli.Context) []string {
	if path := ctx.GlobalString(validatorPasswordFlag.Name); path != "" {
		text, err := ioutil.ReadFile(path)
		if err != nil {
			utils.Fatalf("Failed to read password file: %v", err)
		}
		lines := strings.Split(string(text), "\n")
		// Sanitise DOS line endings.
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], "\r")
		}
		return lines
	}
	if ctx.GlobalIsSet(FakeNetFlag.Name) {
		return []string{validatorpk.FakePassword}
	}
	return nil
}

func unlockValidatorKey(ctx *cli.Context, pubKey validatorpk.PubKey, valKeystore valkeystore.KeystoreI) error {
	if !valKeystore.Has(pubKey) {
		return valkeystore.ErrNotFound
	}
	var err error
	for trials := 0; trials < 3; trials++ {
		prompt := fmt.Sprintf("Unlocking validator key %s | Attempt %d/%d", pubKey.String(), trials+1, 3)
		password := getPassPhrase(prompt, false, 0, makeValidatorPasswordList(ctx))
		err = valKeystore.Unlock(pubKey, password)
		if err == nil {
			log.Info("Unlocked validator key", "pubkey", pubKey.String())
			return nil
		}
		if err.Error() != "could not decrypt key with given password" {
			return err
		}
	}
	// All trials expended to unlock account, bail out
	return err
}
