package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/accounts/keystore"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/native/validatorpk"
	"github.com/unicornultrafoundation/go-u2u/valkeystore"
)

func main() {
	var mnemonic string
	flag.StringVar(&mnemonic, "mnemonic", "", "BIP39 mnemonic for deterministic validator key generation (optional)")
	flag.Parse()

	if len(flag.Args()) < 2 {
		fmt.Println("Usage: setup_validator_node [--mnemonic \"MNEMONIC\"] <validator_id> <datadir>")
		os.Exit(1)
	}

	validatorID := flag.Args()[0]
	dataDir := flag.Args()[1]

	var id int
	fmt.Sscanf(validatorID, "%d", &id)

	if mnemonic != "" {
		fmt.Printf("Using mnemonic for validator %d key generation\n", id)
	} else {
		fmt.Printf("Using deterministic hardcoded key for validator %d\n", id)
	}
	fmt.Printf("Setting up validator node %d in %s\n", id, dataDir)

	// Get the validator private key
	var key *ecdsa.PrivateKey
	var err error
	if mnemonic != "" {
		key, err = makefakegenesis.DeriveKeyFromMnemonic(mnemonic, idx.ValidatorID(id))
		if err != nil {
			fmt.Printf("Failed to derive key from mnemonic: %v\n", err)
			os.Exit(1)
		}
	} else {
		key = makefakegenesis.FakeKey(idx.ValidatorID(id))
	}

	address := crypto.PubkeyToAddress(key.PublicKey)
	pubkeyraw := crypto.FromECDSAPub(&key.PublicKey)
	pubkey := validatorpk.PubKey{
		Raw:  pubkeyraw,
		Type: validatorpk.Types.Secp256k1,
	}

	// Create directories
	keystoreDir := filepath.Join(dataDir, "keystore")
	valKeystoreDir := filepath.Join(keystoreDir, "validator")
	os.MkdirAll(valKeystoreDir, 0755)

	fmt.Printf("Address: %s\n", address.Hex())
	fmt.Printf("PubKey: 0x%x\n", pubkeyraw)
	fmt.Printf("U2U PubKey: %s\n", pubkey.String())
	fmt.Printf("Private Key: 0x%x\n", crypto.FromECDSA(key))

	// 1. Setup regular account keystore (like integration.SetAccountKey)
	regularKeystore := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// Import key with empty password (matching our password file)
	account, err := regularKeystore.ImportECDSA(key, "")
	if err != nil {
		fmt.Printf("Failed to import to regular keystore: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Imported to regular keystore: %s\n", account.Address.Hex())

	// 2. Setup validator keystore
	valKeystore := valkeystore.NewDefaultFileKeystore(valKeystoreDir)
	privateKeyBytes := crypto.FromECDSA(key)
	err = valKeystore.Add(pubkey, privateKeyBytes, "")
	if err != nil {
		fmt.Printf("Failed to add to validator keystore: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Added to validator keystore\n")

	// 3. Verify both keystores work
	if regularKeystore.HasAddress(address) {
		fmt.Printf("✅ Regular keystore has address %s\n", address.Hex())
	}

	if valKeystore.Has(pubkey) {
		fmt.Printf("✅ Validator keystore has pubkey 0x%x\n", pubkeyraw)
	}

	fmt.Printf("\n🎯 Validator node %d setup complete!\n", id)
	fmt.Printf("Use these flags:\n")
	fmt.Printf("  --validator.id %d\n", id)
	fmt.Printf("  --validator.pubkey %s\n", pubkey.String())
	fmt.Printf("  --validator.password /path/to/empty_password_file.txt\n")
	fmt.Printf("  --unlock %s\n", address.Hex())
	fmt.Printf("  --password /path/to/empty_password_file.txt\n")
}