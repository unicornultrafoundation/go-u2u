package launcher

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"

	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/native/validatorpk"
)

func TestFakeNetFlag_NonValidator(t *testing.T) {
	// Configure the instance for IPC attachment
	var ipc string
	if runtime.GOOS == "windows" {
		ipc = `\\.\pipe\u2u.ipc` + strconv.Itoa(trulyRandInt(100000, 999999))
	} else {
		ipc = filepath.Join(t.TempDir(), "u2u.ipc")
	}
	// Start an u2u console, make sure it's cleaned up and terminate the console
	cli := exec(t,
		"--fakenet", "0/3", "--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none", "--cache", "7923",
		"--ipcpath", ipc, "--datadir.minfreedisk", "1", "console")

	// Gather all the infos the welcome message needs to contain
	cli.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	cli.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	cli.SetTemplateFunc("gover", runtime.Version)
	cli.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	cli.SetTemplateFunc("niltime", genesisStart)
	cli.SetTemplateFunc("apis", func() string { return ipcAPIs })

	waitForEndpoint(t, ipc, 4*time.Second)

	// Verify the actual welcome message to the required template
	cli.Expect(`
Welcome to the Consensus JavaScript console!

instance: go-u2u/v{{version}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{.Coinbase}}
at block: 1 ({{niltime}})
 datadir: {{.Datadir}}
 modules: {{apis}}

To exit, press ctrl-d
> {{.InputLine "exit"}}
`)
	cli.ExpectExit()

	wantMessages := []string{
		"Unlocked fake validator",
	}
	for _, m := range wantMessages {
		if strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text contains %q", m)
		}
	}
}

func TestFakeNetFlag_Validator(t *testing.T) {
	// Configure the instance for IPC attachment
	var ipc string
	if runtime.GOOS == "windows" {
		ipc = `\\.\pipe\u2u.ipc` + strconv.Itoa(trulyRandInt(100000, 999999))
	} else {
		ipc = filepath.Join(t.TempDir(), "u2u.ipc")
	}
	// Start an u2u console, make sure it's cleaned up and terminate the console
	cli := exec(t,
		"--fakenet", "3/3", "--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none", "--cache", "7923",
		"--ipcpath", ipc, "--datadir.minfreedisk", "1", "console")

	// Gather all the infos the welcome message needs to contain
	va := readFakeValidator("3/3")
	cli.Coinbase = "0x0000000000000000000000000000000000000000"
	cli.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	cli.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	cli.SetTemplateFunc("gover", runtime.Version)
	cli.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	cli.SetTemplateFunc("niltime", genesisStart)
	cli.SetTemplateFunc("apis", func() string { return ipcAPIs })

	waitForEndpoint(t, ipc, 8*time.Second)

	// Verify the actual welcome message to the required template
	cli.Expect(`
Welcome to the Consensus JavaScript console!

instance: go-u2u/v{{version}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{.Coinbase}}
at block: 1 ({{niltime}})
 datadir: {{.Datadir}}
 modules: {{apis}}

To exit, press ctrl-d
> {{.InputLine "exit"}}
`)
	cli.ExpectExit()

	wantMessages := []string{
		"Unlocked validator key",
		"pubkey=" + va.String(),
	}
	for _, m := range wantMessages {
		if !strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func readFakeValidator(fakenet string) *validatorpk.PubKey {
	n, _, err := parseFakeGen(fakenet)
	if err != nil {
		panic(err)
	}

	if n < 1 {
		return nil
	}

	return &validatorpk.PubKey{
		Raw:  crypto.FromECDSAPub(&makefakegenesis.FakeKey(n).PublicKey),
		Type: validatorpk.Types.Secp256k1,
	}
}
