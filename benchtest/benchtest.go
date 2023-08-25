package benchtest

import (
	"crypto/ecdsa"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/urfave/cli.v1"
)

var (
	config *Config
)

func MakeFakenetFeatures(ctx *cli.Context) error {
	// Set up global config
	config = OpenConfig(ctx)

	// Making transfer txs
	if ctx.GlobalIsSet(GenerateTxTransferFlag.Name) {
		err := generateTransfers(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func importAcc(ctx *cli.Context) error {
	acc := accounts.Account{
		Address: common.HexToAddress(ctx.GlobalString(ImportAccAddrFlag.Name)),
	}
	other, err := openKeyStore(ctx.GlobalString(ImportAccNodeDataDirFlag.Name))
	if err != nil {
		return err
	}

	password := ctx.GlobalString(ImportAccPasswordFlag.Name)

	keyStore, err := MakeKeyStore(ctx)
	if err != nil {
		return err
	}

	decrypted, err := other.Export(acc, password, "")
	if err != nil {
		return err
	}
	_, err = keyStore.Import(decrypted, "", "")
	if err != nil {
		return err
	}

	return nil
}

func generateFakenetAccs(ctx *cli.Context) error {
	var accsCount int = 1000
	if ctx.GlobalInt(GenerateAccountFlag.Name) > 1 {
		accsCount = int(ctx.GlobalInt(GenerateAccountFlag.Name))
	}

	keyStore, err := MakeKeyStore(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < accsCount; i++ {
		reader := rand.New(rand.NewSource(int64(i)))

		key, err := ecdsa.GenerateKey(crypto.S256(), reader)
		if err != nil {
			panic(err)
		}
		_, err = keyStore.ImportECDSA(key, "")
		if err != nil {
			return err
		}
	}

	return nil
}

// initAccsBalances action.
func generateAccsBalances(ctx *cli.Context) error {
	cfg := config
	cfg.URLs = cfg.URLs[:1] // txs from single payer should be sent by single sender
	keyStore, err := MakeKeyStore(ctx)
	if err != nil {
		return err
	}

	var amount int64 = 1e18
	if ctx.GlobalInt(GenerateAccountBalanceFlag.Name) > 0 {
		amount = int64(ctx.GlobalInt(GenerateAccountBalanceFlag.Name))
	}

	maxTps := getTpsLimit(ctx)

	generator := NewBalancesGenerator(cfg, keyStore, amount)
	err = generate(generator, maxTps)

	return err
}

// generateTransfers action.
func generateTransfers(ctx *cli.Context) error {
	cfg := config
	keyStore, err := MakeKeyStore(ctx)
	if err != nil {
		return err
	}

	maxTps := getTpsLimit(ctx)

	generator := NewTransfersGenerator(cfg, keyStore)
	err = generate(generator, maxTps)
	return err
}

// generate is the main generate loop.
func generate(generator Generator, maxTps float64) error {
	cfg := config
	txs := generator.Start()
	defer generator.Stop()

	nodes := NewNodes(cfg, txs)
	go func() {
		for tps := range nodes.TPS() {
			tps += 10.0 * float64(nodes.Count())
			if maxTps > 0.0 && tps > maxTps {
				tps = maxTps
			}
			generator.SetTPS(tps)
		}
	}()

	waitForFinish(nodes.Done)
	return nil
}

func waitForFinish(done <-chan struct{}) {
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		break
	case <-done:
		break
	}
}