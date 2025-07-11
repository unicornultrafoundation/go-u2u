package integrationtests

import (
	"os"

	u2u "github.com/unicornultrafoundation/go-u2u/cmd/u2u/launcher"
)

func DumpSFCStorage(dir string) error {
	// start the SFC storage dump process
	// equivalent to running `u2u db dump-sfc --experimental` but in this local process
	os.Args = []string{
		"u2u",
		"db",
		"dump-sfc",
		"--datadir", dir,
		"--experimental",
	}
	return u2u.Run()
}
