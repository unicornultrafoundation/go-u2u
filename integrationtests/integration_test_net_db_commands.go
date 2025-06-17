package integrationtests

import (
	"os"

	u2u "github.com/unicornultrafoundation/go-u2u/cmd/u2u/launcher"
)

func (n *IntegrationTestNet) DumpSFCStorage(dir string) error {
	// start the SFC storage dump process
	// equivalent to running `u2u db dump-sfc --experimental --sfc` but in this local process
	os.Args = []string{
		"u2u",
		"db",
		"dump-sfc",
		"--experimental",
		"--datadir", dir,
		"--sfc",
	}
	return u2u.Run()
}

func (n *IntegrationTestNet) HealDB(dir string) error {
	// start the DB healing process
	// equivalent to running `u2u db heal --experimental` but in this local process
	os.Args = []string{
		"u2u",
		"db",
		"heal",
		"--datadir", dir,
		"--experimental",
	}
	return u2u.Run()
}
