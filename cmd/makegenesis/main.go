package main

import (
	"compress/gzip"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"path"

	"github.com/unicornultrafoundation/go-helios/common/bigendian"
	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/native/ibr"
	"github.com/unicornultrafoundation/go-u2u/native/ier"
	"github.com/unicornultrafoundation/go-u2u/rlp"
	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesis"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesisstore"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesisstore/fileshash"
)

var (
	outputFile    = flag.String("output", "genesis.g", "Output genesis file path")
	numValidators = flag.Int("validators", 3, "Number of validators")
	networkName   = flag.String("network", "testnet", "Network name")
	networkID     = flag.Uint64("networkid", u2u.FakeNetworkID, "Network ID")
	balance       = flag.String("balance", "1000000000000000000000000", "Initial balance per validator (in wei)")
	stake         = flag.String("stake", "1000000000000000000", "Initial stake per validator (in wei)")
	startEpoch    = flag.Uint64("epoch", 2, "Starting epoch number")
	startBlock    = flag.Uint64("block", 1, "Starting block number")
	mnemonic      = flag.String("mnemonic", "", "BIP39 mnemonic for deterministic validator key generation (optional, uses hardcoded keys if empty)")
)

func main() {
	flag.Parse()

	// Parse balance and stake
	balanceBig, ok := new(big.Int).SetString(*balance, 10)
	if !ok {
		log.Crit("Invalid balance value", "balance", *balance)
	}
	stakeBig, ok := new(big.Int).SetString(*stake, 10)
	if !ok {
		log.Crit("Invalid stake value", "stake", *stake)
	}

	mnemonicStr := *mnemonic
	if mnemonicStr != "" {
		log.Info("Using mnemonic for validator key generation")
	} else {
		log.Info("Using deterministic hardcoded keys (no mnemonic provided)")
	}

	log.Info("Creating genesis file",
		"validators", *numValidators,
		"network", *networkName,
		"networkID", *networkID,
		"balance", balanceBig.String(),
		"stake", stakeBig.String(),
		"epoch", *startEpoch,
		"block", *startBlock,
	)

	// Create network rules
	rules := u2u.FakeNetRules(u2u.Upgrades{
		Berlin:  true,
		London:  true,
		Llr:     true,
		Clymene: true,
	})
	rules.Name = *networkName
	rules.NetworkID = *networkID

	// Generate genesis store
	store := makefakegenesis.FakeGenesisStoreWithRulesAndStartAndMnemonic(
		idx.Validator(*numValidators),
		balanceBig,
		stakeBig,
		rules,
		idx.Epoch(*startEpoch),
		idx.Block(*startBlock),
		mnemonicStr,
	)

	// Export genesis to file using proper U2U format
	err := exportGenesisToFile(store, *outputFile)
	if err != nil {
		log.Crit("Failed to export genesis", "err", err)
	}

	log.Info("Genesis file created successfully", "file", *outputFile)

	// Print validator info
	validators := makefakegenesis.GetFakeValidatorsWithMnemonic(idx.Validator(*numValidators), mnemonicStr)
	fmt.Println("\n=== Validator Information ===")
	for _, v := range validators {
		// Generate the private key for this validator
		var key *ecdsa.PrivateKey
		var err error
		if mnemonicStr != "" {
			key, err = makefakegenesis.DeriveKeyFromMnemonic(mnemonicStr, v.ID)
			if err != nil {
				fmt.Printf("Failed to derive key for validator %d: %v\n", v.ID, err)
				continue
			}
		} else {
			key = makefakegenesis.FakeKey(v.ID)
		}

		fmt.Printf("Validator %d:\n", v.ID)
		fmt.Printf("  Address:     %s\n", v.Address.Hex())
		fmt.Printf("  PubKey:      %s\n", v.PubKey.String())
		fmt.Printf("  Private Key: 0x%x\n", crypto.FromECDSA(key))
		fmt.Printf("  Balance:     %s wei\n", balanceBig.String())
		fmt.Printf("  Stake:       %s wei\n\n", stakeBig.String())
	}
}

type dropableFile struct {
	io.ReadWriteSeeker
	io.Closer
	path string
}

func (f dropableFile) Drop() error {
	return os.Remove(f.path)
}

type realUnitWriter struct {
	plain            io.WriteSeeker
	fileshasher      *fileshash.Writer
	gziper           *gzip.Writer
	dataStartPos     int64
	uncompressedSize uint64
}

func newRealUnitWriter(plain io.WriteSeeker) *realUnitWriter {
	return &realUnitWriter{
		plain: plain,
	}
}

func (w *realUnitWriter) Start(header genesis.Header, name, tmpDirPath string) error {
	// Write unit marker and version
	_, err := w.plain.Write(append(genesisstore.FileHeader, genesisstore.FileVersion...))
	if err != nil {
		return err
	}

	// write genesis header
	err = rlp.Encode(w.plain, genesisstore.Unit{
		UnitName: name,
		Header:   header,
	})
	if err != nil {
		return err
	}

	// Reserve space for hash (32 bytes) + compressed size (8 bytes) + uncompressed size (8 bytes)
	w.dataStartPos, err = w.plain.Seek(8+8+32, io.SeekCurrent)
	if err != nil {
		return err
	}

	// Create gzip writer
	w.gziper, _ = gzip.NewWriterLevel(w.plain, gzip.BestCompression)

	// Create fileshash writer on top of gzip writer
	w.fileshasher = fileshash.WrapWriter(w.gziper, genesisstore.FilesHashPieceSize, func(tmpI int) fileshash.TmpWriter {
		tmpI++
		tmpPath := path.Join(tmpDirPath, fmt.Sprintf("genesis-%s-tmp-%d", name, tmpI))
		_ = os.MkdirAll(tmpDirPath, os.ModePerm)
		tmpFh, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Crit("File opening error", "path", tmpPath, "err", err)
		}
		return dropableFile{
			ReadWriteSeeker: tmpFh,
			Closer:          tmpFh,
			path:            tmpPath,
		}
	})
	return nil
}

func (w *realUnitWriter) Write(b []byte) (n int, err error) {
	n, err = w.fileshasher.Write(b)
	w.uncompressedSize += uint64(n)
	return
}

func (w *realUnitWriter) Flush() (hash.Hash, error) {
	h, err := w.fileshasher.Flush()
	if err != nil {
		return hash.Hash{}, err
	}

	err = w.gziper.Close()
	if err != nil {
		return hash.Hash{}, err
	}

	endPos, err := w.plain.Seek(0, io.SeekCurrent)
	if err != nil {
		return hash.Hash{}, err
	}

	// Go back and write the hash and sizes in the reserved space
	_, err = w.plain.Seek(w.dataStartPos-(8+8+32), io.SeekStart)
	if err != nil {
		return hash.Hash{}, err
	}

	// Write hash (32 bytes)
	_, err = w.plain.Write(h.Bytes())
	if err != nil {
		return hash.Hash{}, err
	}
	// Write compressed size (8 bytes)
	_, err = w.plain.Write(bigendian.Uint64ToBytes(uint64(endPos - w.dataStartPos)))
	if err != nil {
		return hash.Hash{}, err
	}
	// Write uncompressed size (8 bytes)
	_, err = w.plain.Write(bigendian.Uint64ToBytes(w.uncompressedSize))
	if err != nil {
		return hash.Hash{}, err
	}

	// Return to end of file
	_, err = w.plain.Seek(0, io.SeekEnd)
	if err != nil {
		return hash.Hash{}, err
	}
	return h, nil
}

func exportGenesisToFile(store *genesisstore.Store, filePath string) error {
	gen := store.Genesis()

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	header := gen.Header
	tmpPath := "/tmp/makegenesis-export"
	os.MkdirAll(tmpPath, os.ModePerm)
	defer os.RemoveAll(tmpPath)

	// Export epochs
	log.Info("Exporting epochs")
	writer := newRealUnitWriter(f)
	err = writer.Start(header, "ers-1", tmpPath)
	if err != nil {
		return err
	}
	gen.Epochs.ForEach(func(er ier.LlrIdxFullEpochRecord) bool {
		b, _ := rlp.EncodeToBytes(er)
		writer.Write(b)
		return true
	})
	_, err = writer.Flush()
	if err != nil {
		return err
	}
	log.Info("Exported epochs")

	// Export blocks
	log.Info("Exporting blocks")
	writer = newRealUnitWriter(f)
	err = writer.Start(header, "brs-1", tmpPath)
	if err != nil {
		return err
	}
	gen.Blocks.ForEach(func(br ibr.LlrIdxFullBlockRecord) bool {
		b, _ := rlp.EncodeToBytes(br)
		writer.Write(b)
		return true
	})
	_, err = writer.Flush()
	if err != nil {
		return err
	}
	log.Info("Exported blocks")

	// Export EVM data
	log.Info("Exporting EVM data")
	writer = newRealUnitWriter(f)
	err = writer.Start(header, "evm-1", tmpPath)
	if err != nil {
		return err
	}
	gen.RawEvmItems.ForEach(func(key, value []byte) bool {
		// Use the same format as iodb.Write
		// Key length (4 bytes, uint32)
		writer.Write(bigendian.Uint32ToBytes(uint32(len(key))))
		// Key data
		writer.Write(key)
		// Value length (4 bytes, uint32)
		writer.Write(bigendian.Uint32ToBytes(uint32(len(value))))
		// Value data
		writer.Write(value)
		return true
	})
	_, err = writer.Flush()
	if err != nil {
		return err
	}
	log.Info("Exported EVM data")

	return nil
}

