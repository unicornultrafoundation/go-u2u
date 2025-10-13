package main

import (
	"fmt"
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
)

func main() {
	fmt.Println("Validator Keys for Local Network:")
	fmt.Println("=================================")

	for i := idx.ValidatorID(1); i <= 5; i++ {
		key := makefakegenesis.FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		pubkeyraw := crypto.FromECDSAPub(&key.PublicKey)

		fmt.Printf("\nValidator %d:\n", i)
		fmt.Printf("  Address: %s\n", addr.Hex())
		fmt.Printf("  PubKey:  0x%x\n", pubkeyraw)
	}
}