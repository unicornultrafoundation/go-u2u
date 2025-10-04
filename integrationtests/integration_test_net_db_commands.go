package integrationtests

import (
	"os"

	u2u "github.com/unicornultrafoundation/go-u2u/cmd/u2u/launcher"
)

func DumpSFCStorage(dir string) error {
	// start the SFC storage dump process
	// equivalent to running `u2u db dump-sfc --experimental --sfc` but in this local process
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
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
