package integrationtests

import (
	"context"
	"math/big"
	"testing"

	"golang.org/x/net/nettest"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi/bind"
	"github.com/unicornultrafoundation/go-u2u/integrationtests/contracts/prevrandao"
)

func TestPrevRandao(t *testing.T) {
	dataDir, err := nettest.LocalPath()
	if err != nil {
		t.Fatalf(err.Error())
	}
	net, err := StartIntegrationTestNet(dataDir)
	if err != nil {
		t.Fatalf("failed to start the test network; %v", err)
	}
	defer net.Stop()

	// Deploy the contract.
	contract, _, err := DeployContract(net, prevrandao.DeployPrevrandao)
	if err != nil {
		t.Fatalf("failed to deploy contract; %v", err)
	}
	// Collect the current PrevRandao fee from the head state.
	receipt, err := net.Apply(contract.LogCurrentPrevRandao)
	if err != nil {
		t.Fatalf("failed to log current prevrandao; %v", err)
	}
	if len(receipt.Logs) != 1 {
		t.Fatalf("unexpected number of logs; expected 1, got %d", len(receipt.Logs))
	}
	entry, err := contract.ParseCurrentPrevRandao(*receipt.Logs[0])
	if err != nil {
		t.Fatalf("failed to parse log; %v", err)
	}
	fromLog := entry.Prevrandao

	client, err := net.GetClient()
	if err != nil {
		t.Fatalf("failed to get client; %v", err)
	}
	defer client.Close()
	block, err := client.BlockByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		t.Fatalf("failed to get block header; %v", err)
	}
	fromLatestBlock := block.MixDigest().Big() // MixDigest == MixHash == PrevRandao
	if block.Difficulty().Uint64() != 0 {
		t.Errorf("incorrect header difficulty got: %d, want: %d", block.Difficulty().Uint64(), 0)
	}
	// Collect the prevrandao from the archive.
	fromArchive, err := contract.GetPrevRandao(&bind.CallOpts{BlockNumber: receipt.BlockNumber})
	if err != nil {
		t.Fatalf("failed to get prevrandao from archive; %v", err)
	}
	if fromLog.Sign() < 1 {
		t.Fatalf("invalid prevrandao from log; %v", fromLog)
	}

	if fromLog.Cmp(fromLatestBlock) != 0 {
		t.Errorf("prevrandao of block %v mismatch; from log %v, from block %v", receipt.BlockNumber, fromLog, fromLatestBlock)
	}
	if fromLog.Cmp(fromArchive) != 0 {
		t.Errorf("prevrandao mismatch; from log %v, from archive %v", fromLog, fromArchive)
	}

	fromSecondLastBlock, err := contract.GetPrevRandao(&bind.CallOpts{BlockNumber: big.NewInt(receipt.BlockNumber.Int64() - 1)})
	if err != nil {
		t.Fatalf("failed to get prevrandao from archive; %v", err)
	}

	if fromSecondLastBlock.Cmp(fromLatestBlock) == 0 {
		t.Errorf("prevrandao must be different for each block, found same: %s, %s", fromSecondLastBlock, fromLatestBlock)
	}
}
